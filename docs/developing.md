# Developing

Back to [README](../README.md). Outstanding work is in [BACKLOG.md](../BACKLOG.md);
completed decisions are in [DEVELOPMENT_HISTORY.md](../DEVELOPMENT_HISTORY.md).

## Automated testing

From a local checkout, run `go test ./...` and `go vet ./...` from the
repository root. For a local binary during development:

```sh
go build -o dist/fn2fm ./cmd/fn2fm
```

On Windows use `dist/fn2fm.exe`. A local `go build` reports `dev` from
`fn2fm --version`. Release archives stamp `--version` from the tag. `go install …@latest` reports the module version from that install.

Tests use generated PDF fixtures and temporary
folders; no personal score collection is required.

## Packaging

`cmd/buildpackage` builds the release archive from a tested `fn2fm` executable.
It verifies that the executable was built with `CGO_ENABLED=0`, checks its
`--version` output, adds the documentation, the example dictionaries, dependency
notices and an internal checksum list, and writes the archive plus a
`.zip.sha256` file.

```sh
go run ./cmd/buildpackage -binary dist/fn2fm -version v0.3.1 -out dist/packages
```

`release.json` supplies the product name, module path, command name and extra
files. Archives are named `fn2fm-v<version>-<platform>.zip` for all six targets.

The [CI workflow](../.github/workflows/ci.yml) runs on pushes and pull requests
on both macOS architectures, Windows x64/ARM64 and Linux x64/ARM64. The race
detector runs in every native job except Windows ARM64, where Go does not support
it. Each job checks that Go runs natively on the intended OS and architecture,
tests the built executable with external programs unavailable, and creates a
verified package. These checks
cover metadata, paths with spaces, repeated updates, backups, rejected PDFs and
replacement failure while a Windows process holds the PDF open.

To run the executable checks locally, build fn2fm, set `FN2FM_TEST_BINARY` to its
absolute path, and run
`go test -count=1 -v -run '^TestBuiltExecutable$' ./cmd/fn2fm`. For a
version-stamped binary, also set `FN2FM_TEST_VERSION` to the embedded version.
Without that variable, the executable checks are skipped; the Windows file-lock
case also skips on other operating systems. These automated checks do not test
the display of properties or pages in desktop PDF readers.

The [release workflow](../.github/workflows/release.yml) accepts only
`vMAJOR.MINOR.PATCH` tags, calls the CI workflow with the tag as the package
version, and publishes the six archives and a combined `SHA256SUMS.txt` only
after every package has verified.
