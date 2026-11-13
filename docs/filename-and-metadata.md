# Filenames and metadata

Back to [README](../README.md). See also [Using fn2fm](usage.md) for naming
syntax, the dictionary, and examples.

## forScore metadata

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

## PDF metadata

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

## Supported character set

### Why filenames stay restricted

Score filenames have to work in three places at once: on the computer that
runs fn2fm, in cloud storage used to share PDFs, and inside forScore (which
uses the filename stem as the Title). Different systems ban different
characters. fn2fm therefore keeps to printable ASCII and excludes punctuation
that commonly breaks Windows sync or cloud uploads (`" * . / : ; < > ? \ |`),
plus leading or trailing spaces.

Accented letters and other Unicode can already appear in composer and arranger
names via `names.json` and PDF metadata. Putting them in the filename as well
would risk sync clashes (the same name stored in more than one Unicode form)
and has not been re-checked against current forScore behaviour, so filenames
remain ASCII for reliability.

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

## PDF writing behaviour

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
