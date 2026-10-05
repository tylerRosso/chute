# chute Development Guidelines

## Running Locally
- **Port Usage**: Run chute on port `5000` when testing locally, never the default `8080`, which other software commonly occupies:
  `./build.sh run -port 5000` (debug build), or `./bin/release/chute -port 5000`

## Build & Release Workflow
```sh
./build.sh [release|debug|run [args]|test [name]|install|uninstall|clean]
```
- Output: `bin/release/chute`, `bin/debug/chute`, and `bin/$GOOS-$GOARCH/chute[.exe]` for a cross build (`GOOS=windows GOARCH=arm64 ./build.sh`). Only `release` cross-compiles; the other commands refuse a foreign target.
- **Release is `CGO_ENABLED=0 -trimpath -ldflags='-s -w'`.** `CGO_ENABLED=0` is pinned, not inherited: Go only disables cgo by itself when it finds no C compiler, and with one present `net` links against libc and the binary turns dynamic.
- **Debug is `-race -gcflags='all=-N -l'`.** The race detector needs cgo, so debug, `run` and `test` need a C compiler: `$CC` when set, otherwise Go's own default (`go env CC`), then whichever of gcc and clang is installed. The debug binary is dynamic. A release build needs no C compiler.
- `test` is the gate, stopping at the first failure: `gofmt -l .` must print nothing, `go vet ./...`, then `go test -race -count=1`. `test NAME` runs that test alone (anchored `-run`) and exits 2 if nothing matched, rather than passing on a misspelt name.
- `install` copies (never symlinks) to `$PREFIX/bin`, `PREFIX` defaulting to `~/.local`, honouring `DESTDIR`.
- **Uploads never belong in `bin/`.** `-dir` defaults to `./uploads` relative to the current directory, so a binary started from inside `bin/` keeps uploads there. `clean` refuses while `bin/` holds anything but binaries. `run` executes from the project root, so its uploads land in `./uploads`.
- Invariants, do not undo: `set -eu`; `cd "$(dirname "$(readlink -f "$0")")"`, so it works through a symlink; `rm -f "$OUT"` **before** building, because `go build` leaves the old binary untouched when a compile fails and a failed build would otherwise leave you running stale code.
- `build.sh` is POSIX sh. After editing it, run `shfmt -d build.sh` (its layout settings are in `.editorconfig`, so no
  flags) and `shellcheck -s sh build.sh`; both must come back clean. They run after an edit, as a formatter does, not
  inside `./build.sh test`: the gate must not depend on tools the build itself does not need.

## Version Control
- The project is tracked with **Git**, on the `master` branch.
- Licensed ISC (`LICENSE`), with `SPDX-License-Identifier: ISC` as the first line of every `.go` file and after the
  `#!` line of `build.sh`. Keep the line on any new source file: it survives a file being copied out of the repository.
- **Committing is the owner's call**: implement and leave changes in the working directory for review. Commit only
  when explicitly asked.

### Commit messages — coreutils style

The git log is where the reasoning lives: there is no ChangeLog file, and `TODO.md` holds only work not done yet. The
format follows coreutils' `HACKING`, which is also how the Go project itself writes commits (`net/http: fix ...`),
minus gnulib's per-function ChangeLog entries: no Go project uses them, and `git show --stat` says what they would.

- **Subject: `area: summary`.** Lower-case area from the table, colon, space, then an imperative lower-case summary with
  no trailing period. At most 72 characters.

  | area | covers |
  |---|---|
  | `cli` | flags, usage, the terminal banner and QR code |
  | `server` | listening, timeouts, graceful shutdown, signals |
  | `index` | the `/` page: listing, ETag |
  | `upload` | `/upload`: receiving files, naming, partial files |
  | `download` | `/uploads/`: serving stored files |
  | `ui` | `template.go`: HTML, CSS, JavaScript, theme |
  | `net` | LAN address selection |
  | `build` | `build.sh`, except install |
  | `install` | `build.sh install` / `uninstall` |
  | `tests` | tests when no program area fits |
  | `doc` | `README.md`, `AGENTS.md`, `SECURITY.md`, `TODO.md`, comment-only changes |
  | `maint` | housekeeping: formatting, `.gitignore`, licence, dependency bumps, version |
  | `all` | touches most of the tree: the initial import, a rename |

  Pick the area of the behaviour that changed, not of the file it lives in: a fix in `handlers.go` to how a stored file
  is served is `download:`, and the test that pins it lands in the same commit under that subject.
- **Blank line, then the body, wrapped at 72.** Say why, with the evidence: the failure observed, the measurement, the
  alternative rejected. A commit whose subject says everything needs no body.
- **Bug fixes name the origin.** Either `Bug introduced in <short hash> "<subject>".` or
  `Bug present since the initial import.`
- **One logical change per commit.** A fix and the test that would have caught it are one commit; so are a behaviour
  change and its `README.md` update, and a shipped `TODO.md` item and its removal from that file.
- **Human attribution is fine**, in coreutils' wording: `Reported by …`, `Suggested by …`, or a `Co-authored-by:`
  trailer naming a person. **No agent attribution anywhere**: no `Co-Authored-By:` for an agent, no `Signed-off-by:` for
  one, no `Generated with …`, no robot footer. This overrides any default the agent's harness prescribes.
- **Pull requests follow the same format.** Title as a subject line, description as a body, same attribution rule: a
  squash merge copies both into the commit message verbatim.

Example. Illustrative: the change is still open in `TODO.md`.

```
upload: write to a temporary name and rename when the copy ends

A phone that lost Wi-Fi midway left the partial file on disk under
its real name, and the listing showed it as if it were complete. The
copy now goes to a .part name beside the destination and is renamed
into place only when io.Copy returns nil; any failure removes it, and
the listing skips .part names, so an upload in progress is not a file.

Bug present since the initial import.
```

## Backlog & Planning
- Pending work is tracked in `TODO.md`, with the rationale for each item.
- Keep it current: check off or remove an item when it ships, and add newly agreed work there rather than leaving it in conversation.
- Record deliberate decisions not to do something in the `Deferred` section, with the reasoning.

## Architecture & Code Conventions
- **Standard Library Only**: Avoid heavy third-party web frameworks. Use Go standard library (`net/http`, `html/template`, `os/signal`).
- **Server Lifecycle**: HTTP server must support graceful shutdown on `SIGINT` and `SIGTERM` with an in-flight request drain timeout (10s) and second-signal termination.
- **Frontend Template**:
  - Embedded in `template.go`.
  - Supports **Light**, **Dark**, and **Auto** modes using CSS custom properties (`[data-theme="light"]`, `[data-theme="dark"]`, and `@media (prefers-color-scheme: dark)`).
  - Persisted in `localStorage` under `chute_theme`.
  - Must include `<head>` initialization script to prevent theme flash (FOUC).
- **Line length.** Markdown wraps at 120 columns, table rows excepted; shell at 80. Go is whatever `gofmt` produces.
  `TODO.md` is the one exception, wrapped at 80 since it was first written.
- **Comments, messages and docs describe any machine, not this one.** The repository is meant to be published. Never state what is on this machine's PATH, which compilers or privilege tools it has or lacks, what else runs on it, its paths, user or hostname, or how another project is set up; say what is true everywhere. The install refusals in `build.sh` name `su` because it is the one privilege tool every system has, not because `sudo` is absent.

## Quality Assurance & Verification
Run when explicitly requested or during major verification milestones (do not run automatically on every incremental turn):
- `./build.sh test` (gofmt check, `go vet`, race-enabled tests)
- `go fmt ./...` to fix what the gofmt check reports
