# Install with Go

Back to [README](../README.md). Prefer a [release ZIP](../INSTALL.md) if you do
not want to install Go.

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

## Dictionaries

`go install` installs only the `fn2fm` program. It does not install or create `names.json` or `genres.json`.

Processing scores needs a `names.json` file. `genres.json` is needed only when PDF Subject is genres. The usual setting is arrangers, which does not read `genres.json`.

Put the working files in the folder you run `fn2fm` from (often beside your scores), or keep them somewhere else and save that path once. If you already have a customised file, do not replace it with the example.

On a first install, pick one of these:

1. **Create them with the editor.** Change to the folder that should hold the file, then run:

   ```bash
   fn2fm names
   ```

   Run `fn2fm genres` only if you want genres. The editor opens an empty dictionary. The file is written when you save. See [Using fn2fm](usage.md#browser-dictionary-editor) for the rest of the editor.

2. **Use a shared copy in a sync folder.** Keep `names.json` (and `genres.json` if you use genres) in a folder synced by Box, Dropbox, iCloud, OneDrive, Google Drive, or a similar service. Point the editor at that file:

   ```bash
   fn2fm names /path/to/shared/names.json
   ```

   For genres, run `fn2fm genres /path/to/shared/genres.json`. In the editor, choose **Use this dictionary next time**. See [Sharing a dictionary](usage.md#sharing-a-dictionary). Edit one copy at a time.

3. **Copy the starters from the repository.** Follow **This project’s files (optional)** below if you do not already have a local copy, then copy the examples into your scores folder:

   ```bash
   cd "/path/to/scores"
   cp /path/to/forscore-metadata/names.example.json names.json
   cp /path/to/forscore-metadata/genres.example.json genres.json
   ```

   The second copy is only for genres mode. A later `git pull` updates the example files in the project folder. It does not replace `names.json` or `genres.json` in your scores folder. Do not copy the examples over those working files when you update.

Path rules and day-to-day use are in [Using fn2fm](usage.md#name-dictionary).

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

  This creates a folder named `forscore-metadata` with the project files inside.

- **Later:** get the latest project files:

  ```bash
  cd forscore-metadata
  git pull
  ```

Updating these files does **not** by itself update the installed command. To refresh the command, follow **Update** above.

## Next steps

- [Using fn2fm](usage.md) — naming scores and running the command
- [Filenames and metadata](filename-and-metadata.md) — character limits and PDF behaviour
