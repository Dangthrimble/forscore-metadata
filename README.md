# fn2fm - Filename to forScore Metadata

See the [backlog](BACKLOG.md) for outstanding work and the
[development history](DEVELOPMENT_HISTORY.md) for completed work and decisions.

If you only want the command-line program, you do **not** need to download this
project’s files. Skip **This project’s files (optional)** below. The **Install**
step uses `go install …@latest`, which fetches what it needs.

## Git

Git is a tool that downloads and updates project files on your computer.

- **Install:** get Git from [git-scm.com/downloads](https://git-scm.com/downloads) (choose the installer for your computer) and run it.
- **Upgrade:** download a newer installer from the same page and run it again, or use your computer’s usual update method.

Open a **new** Terminal window and check:

```bash
git --version
```

## Go

Go is the programming toolkit used to download and install this command-line program.

- **Install:** get Go **1.25** or newer from [go.dev/dl](https://go.dev/dl/). Choose the installer for your computer, run it, and accept the defaults unless you know you need otherwise.
- **Check:** open a **new** Terminal window and run:

  ```bash
  go version
  ```

  You should see a version of 1.25 or higher.

- **Update:** later, download a newer installer from the same page and run it again (or follow that page’s update steps), then open a new Terminal and run `go version` to confirm.

## Install

In Terminal, run:

```bash
go install github.com/Dangthrimble/forscore-metadata/cmd/fn2fm@latest
```

## Check that it worked

```bash
command -v fn2fm
```

The command should print a path (often ending in `go/bin/…`).

Then try:

```bash
fn2fm --version
```

If you see version text, the program is installed.

## If `command -v` prints nothing (PATH)

Go puts the program in a `bin` folder that your computer may not search yet.

1. Show the folder:

   ```bash
   go env GOBIN
   go env GOPATH
   ```

   Use `GOBIN` if it printed a path; otherwise use `GOPATH` plus `/bin` (example: `/Users/you/go/bin`).

2. **zsh** (default on Mac) — add this line to `~/.zshrc`:

   ```bash
   export PATH="$(go env GOPATH)/bin:$PATH"
   ```

   If `go env GOBIN` printed a path, use that instead of `$(go env GOPATH)/bin`.

3. Save the file. Open a **new** Terminal window (or run `source ~/.zshrc`).

4. Run `command -v fn2fm` again.

**bash:** put the same `export PATH=…` line in `~/.bashrc` or `~/.bash_profile`.

**Windows:** add the folder from `go env GOPATH` + `\bin` (or `GOBIN`) to your user Path under Environment Variables, open a new Command Prompt, and run `where fn2fm`.

## Update

1. Upgrade Git if needed (see **Git**).
2. Update Go if needed (see **Go**).
3. If you keep a local copy of the project, run `git pull` (see **This project’s files**).
4. Re-run the same `go install …@latest` command as in **Install**.
5. Run `command -v fn2fm` again.

## This project’s files (optional)

Only if you want a local copy of the project files. Not required to install or
update the command-line program.

- **First time:** download a full copy of the project. In Terminal:

  ```bash
  git clone https://github.com/Dangthrimble/forscore-metadata.git
  cd forscore-metadata
  ```

  This creates a folder named `fn2fm` with the project files inside.

- **Later:** get the latest project files:

  ```bash
  cd forscore-metadata
  git pull
  ```

Updating these files does **not** by itself update the installed command. To refresh the command, follow **Update** above.

## Running fn2fm

fn2fm writes PDF metadata directly using the bundled pdfcpu library. ExifTool is
no longer required. Keep your editable `names.json` in the folder from which you
run fn2fm (or configure a path as described under Name dictionary), and pass the
PDF filenames as arguments, quoting names with spaces:

```text
fn2fm "O Holy Night ~ AdAd_DaFo[Bb]+.pdf"
fn2fm --names /path/to/names.json "O Holy Night ~ AdAd_DaFo[Bb]+.pdf"
fn2fm names
```

The writer supports unencrypted PDF versions 1.4 through 1.7, including PDFs with
no existing metadata dictionary. It currently rejects signed PDFs, signature
fields, hybrid cross-reference files and files requiring cross-reference repair.
Unsupported files and failed PDF updates keep their original filename and contents;
the command reports the error and exits with a nonzero status if any PDF update
fails. Existing filename/tag validation still renames rejected filenames with
the `_rename` suffix.

Each update is written to a temporary file in the PDF's folder and verified before
replacing the PDF. The first original is saved as `<filename>.pdf_original`;
an existing regular backup is kept without being overwritten. The folder must
be writable, and symbolic links are not followed. Close other PDF editors before
processing the file. Basic file permissions are retained; filesystem timestamps,
extended attributes and ACLs are not copied to the replacement.

Only the PDF Info Title, Author, Subject and Keywords are updated. An absent
composer, arranger, key or accompaniment removes the corresponding metadata
where applicable. Existing PDF dates, Producer and XMP are preserved. Acrobat
may therefore continue to display older XMP values, as it did with the legacy
ExifTool workflow. Incremental updates retain previous PDF revisions; this is
not a method for permanently erasing old metadata.

Automated runtime tests have passed on Windows Server 2025 x64; desktop PDF
reader behaviour has not been manually checked on Windows.

## Automated testing

From a local checkout, run `go test ./...` and `go vet ./...` from the
repository root. For a local binary during development:

```sh
go build -o dist/fn2fm ./cmd/fn2fm
```

On Windows use `dist/fn2fm.exe`. Tests use generated PDF fixtures and temporary
folders; no personal score collection is required.

The [build workflow](.github/workflows/build.yml) runs on pushes and pull requests
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

## forScore Metadata

By default, scores and bookmarks in forScore 14.0 can be tagged with the following [forScore metadata](https://forscore.co/documentation/metadata/):
- Title
- Rating
- Difficulty
- Reference
- Time (i.e. duration)
- Key
- Composer(s)
- Genre(s)
- Tag(s)
- Label(s)

 Some of these types might not fit individual needs and all bar "Title" can be renamed to better suit the needs of the individual.

 ## PDF Metadata

 PDF files can also be tagged with [forScore-specific PDF metadata](https://forscore.co/developers-pdf-metadata/), and these can be fetched by forScore and used to populate forScore metadata with the exception of "Reference" and "Label(s)", as follows:

### PDF Specification

| PDF Value | forScore Value | Format                                                                            |
| --------- | -------------- | --------------------------------------------------------------------------------- |
| Title     | Title          | Single value                                                                      |
| Author    | Composers      | One or more comma-separated values                                                |
| Subject   | Genres         | One or more comma-separated values                                                |
| Keywords  | Tags           | One or more comma-separated values (excluding keywords listed in the chart below) |

### Specialty Keywords

| Keyword      | forScore Value | Format                                 |
| ------------ | -------------- | -------------------------------------- |
| rating:N     | Rating         | Whole number between 0 and 5           |
| difficulty:N | Difficulty     | Whole number between 0 and 3           |
| duration:N   | Duration       | Non-negative whole number (in seconds) |
| keysf:N      | Key            | A whole number between -7 and 7        |
| keymi:N      |                | 0 (major) or 1 (minor)                 |

 The purpose of fn2fm is to allow a PDF filename to be constructed and tagged in such a way that, when parsed, PDF metadata can be derived and added to the file prior to it being imported into forScore. The additional benefit is that, if the file's naming convention is chosen well, it will reduce the likelihood of filename clashes (e.g. scores with the same name by different composers or arrangers; the same score in different keys) while keeping the filenames relatively short.

## Supported Character Set

### forScore title restrictions

In a support reply dated 6 March 2024 (ticket #9110202429965819), forScore
Customer Support relayed the developer's explanation:

- Titles cannot start with a period (`.`).
- Titles cannot contain a pipe (`|`), forward slash (`/`), backslash (`\`) or
  semicolon (`;`).
- Titles cannot contain system-defined control characters, identified in the
  reply as Unicode categories `Cc` and `Cf`. The reply noted that this definition
  could change with an Apple OS update.
- forScore strips whitespace and newlines from the beginning and end of filenames.

The original report concerned a semicolon disappearing from the PDF metadata
title when fetched into forScore. Because fn2fm writes the full filename stem
as the PDF Title, it excludes semicolons throughout that stem and uses `~` as
the metadata separator. This restriction therefore has a forScore compatibility
reason in addition to the separate cloud-storage considerations below.

These are the restrictions reported in the 2024 correspondence, not a new test
of current forScore versions. fn2fm's printable-ASCII policy is stricter: it also
excludes non-ASCII characters and periods anywhere in the filename stem, rather
than only a leading period.

### Cloud storage and project filename policy

An additional possible constraint on PDF filenames (e.g. for choirs) is that, where forScore files are shared via cloud storage, the supported character set has to comply with that of the cloud storage used, as well as that of the device(s) on which the PDF files are prepared for sharing (and on which fn2fm is run). Two popular cloud storage solutions are [Dropbox](https://www.dropbox.com) and [Box](https://www.box.com/) which both provide guidance on filenaming:
- [Naming Dropbox files and folders](https://help.dropbox.com/organize/file-names)
- [Troubleshooting Uploads to Box](https://support.box.com/hc/en-us/articles/360044196773-Troubleshooting-Uploads-to-Box)

Based on those two cloud storage solutions, the characters to avoid are:
- non-printable ASCII characters, which comprise the [ASCII Control Characters](https://www.ascii-code.com/characters/control-characters)
- " (double quote, quotation mark)
- \* (asterisk)
- . (period, full stop)
- / (forward slash, solidus)
- : (colon)
- < (less than)
- \> (greater than)
- ? (question mark)
- \ (backslash, reverse solidus)
- | (vertical bar or pipe)
- emojis or emoticons

In addition:
- Dropbox says to avoid special characters (without specifying what constitutes a special character) and doesn't support names with trailing spaces
- Box only supports the [Unicode Basic Multilingual Plane (BMP)](https://codepoints.net/basic_multilingual_plane) characters and doesn't support names with leading or trailing spaces

Semicolons (`;`) are also forbidden throughout the filename. Only `~` is supported as the metadata separator; semicolon-based filenames are rejected.

## fn2fm Score Naming Convention

The naming convention used by fn2fm has been developed to support the needs of our choir [Cambrensis](https://www.cambrensis.org.uk/), uses only [7-bit ASCII printable characters](https://www.ascii-code.com/characters/printable-characters) constrained by the characters to be avoided listed above, and supports the following metadata:
- Score Title (the same as the PDF filename, but without the ".pdf" extension)
- Composer(s)
- Arranger(s) (used by repurposing "Genre(s)")
- Initial key signature
- Whether or not the score includes an accompaniment (this is visible in the forScore "Tag(s)")

The filename is constructed from two parts separated by a tilde ("~"):
- The score title
- The score metadata

To minimise uncertainty, the filename must contain exactly one tilde. Filenames without a tilde are rejected. If there are two or more, fn2fm will not know which separates the score title from the metadata and the filename will be rejected.

The score metadata is formatted as follows and, apart from the limitation of no trailing spaces in the filename, each can have leading and trailing spaces if they aid readability (we just use a single space before and after the tilde):
- <composer(s)\>_<arranger(s)\>\[<initialKeySignature\>\]<accompanimentIndicator\>

<composer(s)\> and <arranger(s)\> are entered in abbreviated form, with the abbreviation being checked against a list of abbreviations and expanded forms maintained in `names.json` (e.g. "JoRu" for "John Rutter"; "CtEcMr" for "Chris Tomlin, Ed Cash, Matt Redman"). If there are no composers, no text is required between the tilde ("~") and the underscore ("_"). if there are no arrangers, the underscore ("_") is not required and no text is required before the open square bracket ("\["). The choice of abbreviations is up to the individual.

The key brackets are required, but <initialKeySignature\> may be empty. Specify the actual major or minor key where known, using `#` for sharp, `b` for flat and `m` for minor (e.g. `[C]`, `[F#]` or `[F#m]`). This allows fn2fm to write the forScore "Key" metadata.

If the actual key is not known, use the number of accidentals as a fallback: `[0]` for no sharps or flats, `[1#]` through `[7#]` for sharps, or `[1b]` through `[7b]` for flats. These counts produce no forScore key metadata and do not imply major or minor. They retain key-signature information in the filename, helping distinguish copies of the same song with different key signatures.

Use empty brackets (`[]`) when no key can be specified, for example for an atonal score: `Atonal Score ~ JoRu[]+.pdf`. Brackets containing only spaces are also accepted. These produce no forScore key metadata; any accompaniment tag is still included. Different keys can share the same accidental count, so use the actual key where known.

<accompanimentIndicator\> can be a plus ("+") for a tag that states "With Accompaniment", a hyphen/minus ("-") for a tag that states "Without Accompaniment", or left blank for no accompaniment tag.

## Name dictionary

The repository supplies `names.example.json` as a starter dictionary. Copy it to
`names.json` beside your scores and customise that working copy.

By default, fn2fm loads `names.json` from the current working directory (the
folder from which you run the command). It does not need to be beside the
executable. You can also:

- pass `--names /path/to/names.json` for a single run, or
- open the dictionary editor and choose **Use this dictionary next time**, which
  stores that path in the per-user config file (`fn2fm/config.json` under the
  OS application-config directory).

Resolution order is: `--names`, then the saved path, then `./names.json`. A
saved path that is missing is an error; fn2fm does not fall back silently.

When loading the dictionary, fn2fm warns if an abbreviation or name has leading
or trailing whitespace. Warnings identify the affected entry. Processing continues
with leading and trailing whitespace removed from names before they are used as
composer or arranger metadata. Abbreviations are not automatically corrected, and
the dictionary file is not rewritten during PDF runs. Spaces within names, such
as `Dan Forrest`, are preserved.

### Browser dictionary editor

Launch a local editor in your default browser:

```text
fn2fm names
fn2fm names /path/to/names.json
```

The command starts a server bound to `127.0.0.1` on a random port, prints a URL
that includes an access token, and opens that URL. Press Ctrl+C in the terminal
to stop. No separate application or internet service is required.

The editor provides:

- a searchable, sortable list of abbreviations and names
- add, edit and delete forms
- suggested abbreviations (you can accept or amend them)
- uniqueness and filename-safety checks
- the path of the dictionary being edited
- Open / New, Reload, and **Use this dictionary next time**

Separate several people in one entry with commas, for example
`Chris Tomlin, Ed Cash, Matt Redman`. Changing or deleting an abbreviation does
not rename existing PDF files; those filenames will fail lookup until updated.

Each successful save writes the JSON with sorted keys and keeps a single rolling
backup beside the file as `names.json.bak`. If the file changed on disk since it
was loaded (for example after a cloud sync), the editor refuses to overwrite and
asks you to reload.

Abbreviation suggestions follow a documented heuristic and are not guaranteed to
match every historical entry. Existing abbreviations are never rewritten
automatically.

### Sharing a dictionary

A small group can keep one `names.json` in a Box, Dropbox or OneDrive folder:

1. Place the shared `names.json` in the synchronised folder.
2. Each person runs `fn2fm names /path/to/shared/names.json`, then **Use this
   dictionary next time**.
3. Prefer one editor at a time. If two people save close together, reload when
   warned, and recover by hand from `names.json.bak` or a cloud “conflicted
   copy” sibling if one appears.
4. There is no locking or automatic merge; informal coordination is enough for
   infrequent concurrent edits.

PDF processing then uses the same saved path, so the dictionary you maintain is
the one filenames are checked against.

## Examples

In all these examples, the forScore Title will be the same as the filename, without the ".pdf" file extension

- O Magnum Mysterium ~ MoLa[2#].pdf
  - Composers: Morten Lauridsen
  - Arrangers:
  - Key:
  - Tags:
- Spirit Of The Season ~ GbAs_DaHa[Ab]+.pdf
  - Composers: Glen Ballard, Alan Silvestri
  - Arrangers: David Hamilton
  - Key: A♭
  - Tags: With Accompaniment
- African Noel ~ _AnTh[2b]+.pdf
  - Composers: 
  - Arrangers: André J Thomas
  - Key:
  - Tags: With Accompaniment
- Total Praise ~ RiSm[5b]-.pdf
  - Composers: Richard Smallwood
  - Arrangers:
  - Key:
  - Tags: Without Accompaniment
- Sing with Joy at Christmas (Stella Natalis) ~ KaJe[C]+.pdf
  - Composers: Karl Jenkins
  - Arrangers:
  - Key: C
  - Tags: With Accompaniment
- The Witness 16 The Victor ~ JaOc[0]+.pdf
  - Composers: Jamie Owens Collins
  - Arrangers:
  - Key:
  - Tags: With Accompaniment

## Licence

fn2fm is licensed under the [MIT licence](LICENSE). Bundled dependencies retain
their own licences; downloadable packages include `THIRD_PARTY_NOTICES.txt` and
the corresponding licence files in `THIRD_PARTY_LICENSES/`.
