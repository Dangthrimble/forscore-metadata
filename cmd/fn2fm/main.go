/*
Fn2fm parses the filename of a music score in PDF format, deriving and writing
PDF metadata to that file that can subsequently be fetched by forScore
[https://forscore.co/].
It uses specific characters to deliniate the metadata embedded in the filename.
It requires a JSON file names.json (by default in the folder from which it is
invoked), or a configured / --names path.

Usage:

	fn2fm [flags] [file ...]
	fn2fm names [path]
	fn2fm --version
*/
package main

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"

	"github.com/Dangthrimble/fn2fm/internal/filename"
	"github.com/Dangthrimble/fn2fm/internal/names"
	"github.com/Dangthrimble/fn2fm/internal/namesedit"
	"github.com/Dangthrimble/fn2fm/internal/pdfmeta"
)

// Set by the build workflow; source builds retain the development label.
var version = "dev"

func main() {
	if len(os.Args) == 2 && os.Args[1] == "--version" {
		fmt.Printf("fn2fm %s (%s/%s)\n", version, runtime.GOOS, runtime.GOARCH)
		return
	}

	if len(os.Args) >= 2 && os.Args[1] == "names" {
		if err := namesedit.Run(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "fn2fm names: %v\n", err)
			os.Exit(1)
		}
		return
	}

	namesFlag, files := parseArgs(os.Args[1:])
	path, err := names.ResolvePath(namesFlag)
	if err != nil {
		fmt.Printf("Unable to resolve names dictionary path: %v\n", err)
		os.Exit(1)
	}

	dict, err := names.Load(path)
	if err != nil {
		// Match prior behaviour: log the underlying error for missing/unreadable files.
		if !strings.Contains(err.Error(), "invalid ") {
			log.Println(err)
		}
		fmt.Printf("Unable to load names.json: %v\n", err)
		os.Exit(1)
	}
	for _, warning := range dict.Warnings {
		log.Println(warning)
	}
	ca := dict.Entries

	var failed bool
	for _, file := range files {
		fmt.Printf("Parsing %q...\n", file)
		fn, err := filename.ValidateAndExtension(file)
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

		comp, err := filename.ParseComposers(md, ca)
		if err != nil {
			fmt.Printf("  ERROR: %v\n", err)
			continue
		}

		arr, err := filename.ParseArrangers(md, ca)
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
		fmt.Printf("    Arranger(s): %q\n", arr)
		fmt.Printf("            Key: %q\n", key)
		fmt.Printf("  Accompaniment: %q\n\n", acc)

		err = pdfmeta.Write(file, pdfmeta.Metadata{
			Title: fn, Author: comp, Subject: arr, Keywords: pdfmeta.Keywords(key, acc),
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

func parseArgs(args []string) (namesFlag string, files []string) {
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--names":
			if i+1 >= len(args) {
				fmt.Fprintln(os.Stderr, "fn2fm: --names requires a path")
				os.Exit(2)
			}
			i++
			namesFlag = args[i]
		case strings.HasPrefix(arg, "--names="):
			namesFlag = strings.TrimPrefix(arg, "--names=")
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
	return namesFlag, files
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `Usage:
  fn2fm [--names path] [file ...]
  fn2fm names [path]
  fn2fm --version

Loads composer/arranger abbreviations from names.json in the current folder,
a path saved via the names editor, or --names for a single run.
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
