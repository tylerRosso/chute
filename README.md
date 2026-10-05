# chute

Drop files between your computer and your phone over the LAN. `chute` is a lightweight, zero-configuration file sharing
server written in Go: start it, scan the QR code it prints, and the phone's browser becomes the drop zone, in both
directions.

---

## Features

- **Terminal QR Code**: Automatically prints a QR code in the terminal for instant pairing with smartphone cameras.
- **Minimalist & Fast**: Built primarily with the Go standard library for speed, low memory usage, and zero runtime bloat. Pre-parses templates at startup and leverages HTTP conditional caching (`ETag` / `304 Not Modified`) for instant page reloads.
- **Modern Web Interface**:
  - Drag-and-drop file uploads with real-time progress, transfer speed, and estimated time remaining.
  - File browser with human-readable file sizes and instant download links.
  - **Efficient Caching**: Conditional validation ensures instantaneous reloads and zero redundant data transfer over Wi-Fi when file contents are unchanged.
  - **Theme Support**: Seamless switching between **Light**, **Dark**, and **Auto** (system preference) modes, with no flash of unstyled content (FOUC) on load.
- **Graceful Shutdown**: Handles `SIGINT` and `SIGTERM` cleanly, providing a 10-second drain window for active uploads and instant termination on a second signal.
- **Cross-Platform**: Static binaries, cross-compiled for Linux, macOS, and Windows with one command.

---

## Quick Start

### Requirements

- Go 1.26.5 or newer (the version in `go.mod`).
- For `./build.sh debug`, `run` and `test` only: a C compiler: `$CC` if set, otherwise Go's default, then gcc or clang, whichever is installed. The race detector needs cgo, and cgo needs a C compiler. Release builds need none.

### Running from Source

```bash
# Run directly with Go (e.g., on port 5000)
go run . -port 5000
```

Or install the latest release without cloning:

```bash
go install github.com/tylerRosso/chute@latest
```

Once running, `chute` prints:
- Local desktop URL (`http://localhost:5000`)
- Mobile network URL (`http://<LAN-IP>:5000`)
- A terminal-rendered QR code for mobile scanning

### Building Binaries

```bash
./build.sh              # build the release binary (the default)
./build.sh release      # same thing, spelled out
./build.sh debug        # unoptimised, with the race detector
./build.sh run [args]   # build debug, then run it with any extra arguments
./build.sh test [name]  # build, then gofmt, go vet and go test -race
./build.sh install      # build, then copy the binary to $PREFIX/bin
./build.sh uninstall    # remove it again
./build.sh clean        # remove bin/
```

The release binary is fully static and lands in `bin/release/chute`; the debug one in `bin/debug/chute`.

`PREFIX` defaults to `~/.local`. Override it for anywhere else, or add `DESTDIR` to stage an install for a package:

```bash
PREFIX=/usr/local ./build.sh install            # needs root
DESTDIR=/tmp/stage PREFIX=/usr ./build.sh install
```

To cross-compile, set Go's own `GOOS` and `GOARCH`. The binary lands in `bin/$GOOS-$GOARCH/`:

```bash
GOOS=windows GOARCH=arm64 ./build.sh    # bin/windows-arm64/chute.exe
GOOS=darwin GOARCH=arm64 ./build.sh     # bin/darwin-arm64/chute
```

---

## Command-Line Options

| Flag | Default | Description |
|------|---------|-------------|
| `-port` | `8080` | Port to run the HTTP server on |
| `-dir` | `./uploads` | Directory where uploaded files are saved |
| `-no-qr` | `false` | Disable printing the QR code in the terminal |

### Examples

```bash
# Store uploads in a custom directory on port 5000
./bin/release/chute -port 5000 -dir /path/to/my/folder

# Run without the terminal QR code
./bin/release/chute -port 5000 -no-qr
```

---

## Security

`chute` has no authentication. It listens on every network interface, and anyone who can reach the port can upload
files and download every file in the upload directory. That is the design: it exists to move files between your own
devices on a network you trust, such as a home or office LAN, with nothing to install or sign in to on the phone.

Treat it accordingly. Do not run it on public Wi-Fi or on a machine reachable from the Internet; point `-dir` at a
directory holding nothing you would not hand to everyone on the network; and stop it when the transfer is done.
Uploaded files are stored as received and served back as they are, so open them with the care you would give any file
from another device.

Known hardening work is listed under Security in [TODO.md](TODO.md).

---

## Project Structure

```
.
├── main.go            # Entrypoint, CLI flag parsing, and graceful server lifecycle
├── handlers.go        # HTTP route handlers (index, upload, static uploads) and ETag caching
├── template.go        # Pre-parsed HTML/CSS/JS frontend template with theme engine
├── utils.go           # Local IP resolution and byte size formatting
├── build.sh           # Build, test, install and cross-compile script
├── main_test.go       # Server lifecycle and router integration tests
├── handlers_test.go   # Upload, ETag caching, and file-serving unit tests
├── utils_test.go      # IP detection and size formatting unit tests
├── AGENTS.md          # Project conventions and development rules
└── TODO.md            # Backlog of agreed, unimplemented work
```

---

## Development & Testing

Run the full check: gofmt, `go vet`, then the tests under the race detector, stopping at the first failure:

```bash
./build.sh test
./build.sh test TestHandleIndex    # just one test
```

`go fmt ./...` fixes whatever the gofmt check reports. Plain `go test ./...` still works without a C compiler, minus the race detector.

---

## License

ISC. See [LICENSE](LICENSE).
