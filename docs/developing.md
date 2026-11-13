# Developing

Back to [README](../README.md). Outstanding work is in [BACKLOG.md](../BACKLOG.md);
completed decisions are in [DEVELOPMENT_HISTORY.md](../DEVELOPMENT_HISTORY.md).

## Automated testing

From a local checkout, run `go test ./...` and `go vet ./...` from the
repository root. For a local binary during development:

```sh
go build -o dist/fn2fm ./cmd/fn2fm
```

On Windows use `dist/fn2fm.exe`. Tests use generated PDF fixtures and temporary
folders; no personal score collection is required.

The [build workflow](../.github/workflows/build.yml) runs on pushes and pull requests
on both macOS architectures, Windows x64/ARM64 and Linux x64/ARM64. Each job checks
that Go runs natively on the intended OS and architecture, tests the built
executable with external programs unavailable, and creates a verified package. These checks
cover metadata, paths with spaces, repeated updates, backups, rejected PDFs and
replacement failure while a Windows process holds the PDF open.

To run the executable checks locally, build fn2fm, set `FN2FM_TEST_BINARY` to its
absolute path, and run
`go test -count=1 -v -run '^TestBuiltExecutable$' ./cmd/fn2fm`. For a
version-stamped binary, also set `FN2FM_TEST_VERSION` to the embedded version.
Without that variable, the executable checks are skipped; the Windows file-lock
case also skips on other operating systems. These automated checks do not test
the display of properties or pages in desktop PDF readers.
