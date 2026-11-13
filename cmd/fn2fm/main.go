/*
Fn2fm parses the filename of a music score in PDF format, deriving and writing
PDF metadata to that file that can subsequently be fetched by forScore
[https://forscore.co/].
It uses specific characters to deliniate the metadata embedded in the filename.
It requires a JSON file names.json (by default in the folder from which it is
invoked), or a configured / --names path. PDF Subject is arrangers unless the
user or the scores folder selects genres, which then also uses genres.json.

Usage:

	fn2fm [flags] [file ...]
	fn2fm names [path]
	fn2fm genres [path]
	fn2fm subject [arrangers|genres] [--library]
	fn2fm --version
*/
package main

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"

	"github.com/Dangthrimble/forscore-metadata/internal/filename"
	"github.com/Dangthrimble/forscore-metadata/internal/names"
	"github.com/Dangthrimble/forscore-metadata/internal/namesedit"
	"github.com/Dangthrimble/forscore-metadata/internal/pdfmeta"
)

// Set by the build workflow; source builds retain the development label.
var version = "dev"

func main() {
	if len(os.Args) == 2 && os.Args[1] == "--version" {
		fmt.Printf("fn2fm %s (%s/%s)\n", version, runtime.GOOS, runtime.GOARCH)
		return
	}

	if len(os.Args) >= 2 && os.Args[1] == "names" {
		if err := namesedit.Run(namesedit.Names, os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "fn2fm names: %v\n", err)
			os.Exit(1)
		}
		return
	}
	if len(os.Args) >= 2 && os.Args[1] == "genres" {
		if err := namesedit.Run(namesedit.Genres, os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "fn2fm genres: %v\n", err)
			os.Exit(1)
		}
		return
	}
	if len(os.Args) >= 2 && os.Args[1] == "subject" {
		if err := runSubject(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "fn2fm subject: %v\n", err)
			os.Exit(1)
		}
		return
	}

	flags, files := parseArgs(os.Args[1:])
	choice, err := names.EffectiveSubject(flags.subject)
	if err != nil {
		fmt.Printf("Unable to resolve subject mode: %v\n", err)
		os.Exit(1)
	}
	if flags.genres != "" && choice.Mode != names.SubjectGenres {
		fmt.Fprintln(os.Stderr, "fn2fm: --genres applies when subject is genres")
		os.Exit(2)
	}

	namesDict, err := loadDictionary("names.json", flags.names, names.ResolvePath)
	if err != nil {
		os.Exit(1)
	}
	abbreviations := namesDict.Entries
	slotDict := abbreviations
	slotLabel := "Arranger(s)"
	if choice.Mode == names.SubjectGenres {
		genresDict, err := loadDictionary("genres.json", flags.genres, names.ResolveGenresPath)
		if err != nil {
			os.Exit(1)
		}
		slotDict = genresDict.Entries
		slotLabel = "Genre(s)"
	}

	var failed bool
	for _, file := range files {
		fmt.Printf("Parsing %q...\n", file)
		fn, err := filename.ValidateFilenameAndExtension(file)
		if err != nil {
			fmt.Printf("  ERROR: %v\n", err)
			if !renameRejectedFile(file) {
				failed = true
			}
			continue
		}

		md, err := filename.ValidateMetadataTags(fn)
		if err != nil {
			fmt.Printf("  ERROR: %v\n", err)
			if !renameRejectedFile(file) {
				failed = true
			}
			continue
		}

		comp, err := filename.ParseComposers(md, abbreviations)
		if err != nil {
			fmt.Printf("  ERROR: %v\n", err)
			continue
		}

		// The underscore slot is arrangers or genres; slotDict selects which.
		subject, err := filename.ParseArrangers(md, slotDict)
		if err != nil {
			fmt.Printf("  ERROR: %v\n", err)
			continue
		}

		key, err := filename.ParseKey(md)
		if err != nil {
			fmt.Printf("  ERROR: %v\n", err)
			continue
		}

		acc, err := filename.ParseAccompaniment(md)
		if err != nil {
			fmt.Printf("  ERROR: %v\n", err)
			continue
		}

		fmt.Printf("    Composer(s): %q\n", comp)
		fmt.Printf("    %s: %q\n", slotLabel, subject)
		fmt.Printf("            Key: %q\n", key)
		fmt.Printf("  Accompaniment: %q\n\n", acc)

		err = pdfmeta.Write(file, pdfmeta.Metadata{
			Title: fn, Author: comp, Subject: subject, Keywords: pdfmeta.Keywords(key, acc),
		})
		if err != nil {
			log.Printf("Unable to update %q: %v", file, err)
			failed = true
			continue
		}
		fmt.Printf("  Updated PDF metadata\n")
	}
	if failed {
		os.Exit(1)
	}
}

type runFlags struct {
	names   string
	genres  string
	subject string
}

func parseArgs(args []string) (runFlags, []string) {
	var flags runFlags
	var files []string
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--names" || arg == "--genres" || arg == "--subject":
			if i+1 >= len(args) {
				fmt.Fprintf(os.Stderr, "fn2fm: %s requires a value\n", arg)
				os.Exit(2)
			}
			i++
			switch arg {
			case "--names":
				flags.names = args[i]
			case "--genres":
				flags.genres = args[i]
			case "--subject":
				flags.subject = args[i]
			}
		case strings.HasPrefix(arg, "--names="):
			flags.names = strings.TrimPrefix(arg, "--names=")
		case strings.HasPrefix(arg, "--genres="):
			flags.genres = strings.TrimPrefix(arg, "--genres=")
		case strings.HasPrefix(arg, "--subject="):
			flags.subject = strings.TrimPrefix(arg, "--subject=")
		case arg == "--help" || arg == "-h":
			printUsage()
			os.Exit(0)
		case strings.HasPrefix(arg, "-"):
			fmt.Fprintf(os.Stderr, "fn2fm: unknown flag %s\n", arg)
			os.Exit(2)
		default:
			files = append(files, arg)
		}
	}
	return flags, files
}

func loadDictionary(label, flag string, resolve func(string) (string, error)) (*names.Dictionary, error) {
	path, err := resolve(flag)
	if err != nil {
		fmt.Printf("Unable to resolve %s path: %v\n", label, err)
		return nil, err
	}
	dict, err := names.Load(path)
	if err != nil {
		// Match prior behaviour: log the underlying error for missing/unreadable files.
		if !strings.Contains(err.Error(), "invalid ") {
			log.Println(err)
		}
		fmt.Printf("Unable to load %s: %v\n", label, err)
		return nil, err
	}
	for _, warning := range dict.Warnings {
		log.Println(warning)
	}
	return dict, nil
}

func runSubject(args []string) error {
	library := false
	var mode string
	for _, arg := range args {
		switch {
		case arg == "--library":
			library = true
		case arg == "--help" || arg == "-h":
			printUsage()
			return nil
		case strings.HasPrefix(arg, "-"):
			return fmt.Errorf("unknown flag %s", arg)
		default:
			if mode != "" {
				return fmt.Errorf("usage: fn2fm subject [arrangers|genres] [--library]")
			}
			mode = arg
		}
	}
	if library && mode == "" {
		return fmt.Errorf("--library requires arrangers or genres")
	}
	if mode != "" {
		if library {
			if err := names.SetLibrarySubject(mode); err != nil {
				return err
			}
			fmt.Printf("Saved library subject: %s\n", strings.TrimSpace(mode))
		} else {
			if err := names.SetUserSubject(mode); err != nil {
				return err
			}
			fmt.Printf("Saved user subject: %s\n", strings.TrimSpace(mode))
		}
	}
	choice, err := names.EffectiveSubject("")
	if err != nil {
		return err
	}
	fmt.Printf("Effective subject: %s\n", choice.Label())
	return nil
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `Usage:
  fn2fm [--names path] [--genres path] [--subject arrangers|genres] [file ...]
  fn2fm names [path]
  fn2fm genres [path]
  fn2fm subject [arrangers|genres] [--library]
  fn2fm --version

Loads composer abbreviations from names.json in the current folder, a path saved
via the names editor, or --names for a single run. PDF Subject is arrangers by
default (Cambrensis uses that slot for arrangers, not genres). Choose genres
with --subject, a saved user setting, or ./fn2fm.json. Genres mode loads
genres.json the same way, and only then.
`)
}

// renameRejectedFile moves a filename/tag validation failure aside with the
// existing `_rename` suffix. It reports rename failures and returns false when
// the move did not succeed.
func renameRejectedFile(file string) bool {
	target := file + "_rename"
	if err := os.Rename(file, target); err != nil {
		fmt.Printf("  ERROR: unable to rename %q to %q: %v\n", file, target, err)
		return false
	}
	return true
}
