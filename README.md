# fn2fm - Filename to forScore Metadata

fn2fm turns a carefully named score PDF into forScore-ready PDF metadata
(title, composers, arrangers or genres, key, and accompaniment tags).
Arrangers are the default.

If you only want the command-line program, you do **not** need to download this
project’s files. Choose an install path below; `go install …@latest` fetches
what it needs without a local copy of the repository.

## Install

Pick one:

- **With Go** (recommended for updates from source): follow
  [Install with Go](docs/install-go.md) — Git, Go, `go install`, PATH, and
  optional project files.
- **Without Go** (pre-built ZIP): follow [INSTALL.md](INSTALL.md) — download a
  release package for your computer and run the executable.

## Try it

Keep a `names.json` dictionary beside your scores (copy from
`names.example.json`), name PDFs with the tilde convention, then run:

```text
fn2fm "O Holy Night ~ AdAd_DaFo[Bb]+.pdf"
```

Full naming rules, the dictionary editor, and more examples are in
[Using fn2fm](docs/usage.md).

## Docs

- [Using fn2fm](docs/usage.md) — run the command, score naming, arrangers or genres, `names.json`, examples
- [Filenames and metadata](docs/filename-and-metadata.md) — character limits, forScore/PDF metadata, how PDFs are updated
- [Install with Go](docs/install-go.md) — Git, Go, `go install`, PATH, update, optional clone
- [INSTALL.md](INSTALL.md) — install from a release ZIP (no Go required)
- [Developing](docs/developing.md) — tests, local builds, CI checks
- [Backlog](BACKLOG.md) — outstanding work
- [Development history](DEVELOPMENT_HISTORY.md) — completed work and decisions

## Licence

fn2fm is licensed under the [MIT licence](LICENSE). Bundled dependencies retain
their own licences; downloadable packages include `THIRD_PARTY_NOTICES.txt` and
the corresponding licence files in `THIRD_PARTY_LICENSES/`.
