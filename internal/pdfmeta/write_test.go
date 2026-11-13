package pdfmeta

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pdfcpu/pdfcpu/pkg/api"

	"github.com/Dangthrimble/forscore-metadata/internal/pdftest"
)

func TestWritePDFMetadata(t *testing.T) {
	for _, tc := range []struct {
		name string
		data []byte
	}{
		{"existing Info", pdftest.ExistingInfoPDF(false)},
		{"no Info classic table", pdftest.ReadFile(t, "../../testdata/no-info-table.pdf")},
		{"no Info xref stream", pdftest.ReadFile(t, "../../testdata/no-info-stream.pdf")},
		{"no final newline", bytes.TrimRight(pdftest.ExistingInfoPDF(false), "\r\n")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "Score.pdf")
			pdftest.WriteFile(t, path, tc.data)
			before := pdftest.ReadPDF(t, tc.data, Configuration())
			metadata := Metadata{Title: "New score (test)", Author: "André J Thomas", Subject: "Dan Forrest", Keywords: "keysf:-2, keymi:0, With Accompaniment"}
			if err := Write(path, metadata); err != nil {
				t.Fatal(err)
			}
			updated := pdftest.ReadFile(t, path)
			if !bytes.HasPrefix(updated, tc.data) {
				t.Fatal("original bytes changed")
			}
			if !bytes.Equal(pdftest.ReadFile(t, path+"_original"), tc.data) {
				t.Fatal("backup differs from original")
			}
			after := pdftest.ReadPDF(t, updated, Configuration())
			if after.PageCount != before.PageCount || after.XRefTable.Version() != before.XRefTable.Version() || after.Read.UsingXRefStreams != before.Read.UsingXRefStreams {
				t.Fatal("page count, version or cross-reference format changed")
			}
			info, err := after.DereferenceDict(*after.Info)
			if err != nil {
				t.Fatal(err)
			}
			for key, want := range metadata.Fields() {
				got, err := after.DereferenceText(info[key])
				if err != nil || got != want {
					t.Errorf("%s = %q, %v; want %q", key, got, err, want)
				}
			}
			if before.Info != nil {
				oldInfo, _ := before.DereferenceDict(*before.Info)
				for _, key := range []string{"Producer", "CreationDate", "ModDate", "Custom"} {
					if info[key] != oldInfo[key] {
						t.Errorf("unrelated Info field %s changed", key)
					}
				}
				for _, key := range []string{"Metadata", "AcroForm"} {
					if before.RootDict[key] != after.RootDict[key] {
						t.Errorf("catalog reference %s changed", key)
					}
				}
			}
			// A later edit clears absent fields but retains the first backup.
			if err := Write(path, Metadata{Title: "Title only"}); err != nil {
				t.Fatal(err)
			}
			second := pdftest.ReadPDF(t, pdftest.ReadFile(t, path), Configuration())
			secondInfo, _ := second.DereferenceDict(*second.Info)
			for _, key := range []string{"Author", "Subject", "Keywords"} {
				if _, exists := secondInfo[key]; exists {
					t.Errorf("empty %s was not removed", key)
				}
			}
			if !bytes.Equal(pdftest.ReadFile(t, path+"_original"), tc.data) {
				t.Fatal("first backup overwritten on repeated edit")
			}
			pdftest.AssertNoTemporaryPDFs(t, filepath.Dir(path))
		})
	}
}

func TestUnsupportedPDFsRemainUnchanged(t *testing.T) {
	base := pdftest.ExistingInfoPDF(false)
	cases := []struct {
		name, message string
		data          []byte
	}{
		{"malformed", "read PDF", []byte("This is not a PDF")},
		{"old PDF", "versions 1.4 through 1.7", bytes.Replace(base, []byte("%PDF-1.4"), []byte("%PDF-1.3"), 1)},
		{"PDF 2", "versions 1.4 through 1.7", bytes.Replace(base, []byte("%PDF-1.4"), []byte("%PDF-2.0"), 1)},
		{"signature field", "signature", pdftest.ExistingInfoPDF(true)},
	}
	for _, password := range []string{"", "test-password"} {
		conf := Configuration()
		conf.UserPW = password
		conf.OwnerPW = "test-owner"
		var encrypted bytes.Buffer
		if err := api.Encrypt(bytes.NewReader(base), &encrypted, conf); err != nil {
			t.Fatal(err)
		}
		cases = append(cases, struct {
			name, message string
			data          []byte
		}{"encrypted " + password, "encrypted/password-protected", encrypted.Bytes()})
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "Score.pdf")
			pdftest.WriteFile(t, path, tc.data)
			err := Write(path, Metadata{Title: "New title"})
			if err == nil || !strings.Contains(err.Error(), tc.message) {
				t.Fatalf("got %v; expected %q", err, tc.message)
			}
			if !bytes.Equal(pdftest.ReadFile(t, path), tc.data) {
				t.Fatal("rejected input changed")
			}
			entries, _ := os.ReadDir(dir)
			if len(entries) != 1 {
				t.Fatal("rejected input left backup or temporary file")
			}
		})
	}
}

func TestPDFFileReplacementFailures(t *testing.T) {
	for _, fail := range []string{"source changed", "backup blocked", "rename failed"} {
		t.Run(fail, func(t *testing.T) {
			dir := t.TempDir()
			path, temporary := filepath.Join(dir, "Score.pdf"), filepath.Join(dir, "temporary.pdf")
			source := pdftest.ExistingInfoPDF(false)
			pdftest.WriteFile(t, path, source)
			pdftest.WriteFile(t, temporary, []byte("staged replacement"))
			original, _ := os.Stat(path)
			want := source
			switch fail {
			case "source changed":
				want = []byte("another editor changed this file")
				pdftest.WriteFile(t, path, want)
			case "backup blocked":
				if err := os.Mkdir(path+"_original", 0700); err != nil {
					t.Fatal(err)
				}
			}
			called := false
			err := replacePDFFile(path, temporary, source, original, func(string, string) error {
				called = true
				return errors.New("injected rename failure")
			})
			if err == nil || called != (fail == "rename failed") {
				t.Fatalf("unexpected replacement result: %v; rename called=%t", err, called)
			}
			if !bytes.Equal(pdftest.ReadFile(t, path), want) {
				t.Fatal("failed replacement altered source")
			}
		})
	}
}

func TestPDFBackupFailureCleansTemporaryFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "Score.pdf")
	source := pdftest.ExistingInfoPDF(false)
	pdftest.WriteFile(t, path, source)
	if err := os.Mkdir(path+"_original", 0700); err != nil {
		t.Fatal(err)
	}
	if err := Write(path, Metadata{Title: "New title"}); err == nil {
		t.Fatal("expected backup failure")
	}
	if !bytes.Equal(pdftest.ReadFile(t, path), source) {
		t.Fatal("backup failure changed input")
	}
	pdftest.AssertNoTemporaryPDFs(t, dir)
}

func TestMetadataKeywords(t *testing.T) {
	for _, tc := range []struct{ key, acc, want string }{
		{"keysf:0, keymi:0", "With Accompaniment", "keysf:0, keymi:0, With Accompaniment"},
		{"keysf:0, keymi:0", "", "keysf:0, keymi:0"},
		{"", "Without Accompaniment", "Without Accompaniment"},
		{"", "", ""},
	} {
		if got := Keywords(tc.key, tc.acc); got != tc.want {
			t.Errorf("got %q, want %q", got, tc.want)
		}
	}
}
