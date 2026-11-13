# Using fn2fm

Back to [README](../README.md). See also [Filenames and metadata](filename-and-metadata.md)
for character limits and how PDF fields are written.

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

Writer limits, backups, and which PDF fields are updated are described in
[Filenames and metadata](filename-and-metadata.md#pdf-writing-behaviour).

## Score naming convention

The naming convention used by fn2fm has been developed to support the needs of our choir [Cambrensis](https://www.cambrensis.org.uk/), uses only [7-bit ASCII printable characters](https://www.ascii-code.com/characters/printable-characters) constrained by the characters to be avoided (see [Filenames and metadata](filename-and-metadata.md)), and supports the following metadata:
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
