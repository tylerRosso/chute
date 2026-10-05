# TODO

Agreed work that is not implemented yet. Each item records why it matters, so
neither the reasoning nor the trade-offs have to be re-derived later.

---

## Security (priority)

- [ ] **Prevent accidental overwrites on upload**

  `handleUpload` in `handlers.go` calls `os.Create`, which truncates an existing
  file of the same name without warning. Two devices uploading `image.png` means
  the second silently destroys the first. Check whether the destination exists
  and auto-rename instead: `report.pdf` becomes `report (1).pdf`.

  Do this together with the interrupted-upload item under Robustness: both
  change the same few lines, and the temporary-file-then-rename shape that item
  needs is also where the rename-on-collision belongs.

- [ ] **Harden `/uploads/` file serving**

  The route is a bare `http.FileServer`, which has two consequences. It serves
  Go's default directory index at `/uploads/`, listing the whole folder. And an
  uploaded `.html` or `.svg` is rendered inline rather than downloaded, so its
  JavaScript executes under chute's own origin (stored XSS) with no
  authentication required to plant it.

  Return 404 for directory paths, and serve files with
  `Content-Disposition: attachment; filename="..."` plus
  `X-Content-Type-Options: nosniff` so browsers always download rather than
  render. The `nosniff` header matters on its own: without it a browser may
  inspect the bytes, decide a file is HTML regardless of its declared type, and
  reintroduce the problem.

- [ ] **Configure `http.Server` timeouts**

  The server in `main.go` has no timeouts, so an idle or half-open connection
  can hold resources indefinitely. Add `ReadHeaderTimeout: 5 * time.Second` and
  `IdleTimeout: 120 * time.Second`.

  Deliberately do **not** set `ReadTimeout`: it caps the entire request body
  read, which would abort slow or large uploads midway. That is the opposite of
  what this tool is for.

- [ ] **Keep internal error text out of responses**

  `handleUpload` passes `err.Error()` from `os.Create` and `io.Copy` straight to
  the client, and those messages include the absolute path of the upload
  directory. Log the full error on the server and answer with a generic
  message. In the same spirit, `handleIndex` swallows a failed `os.ReadDir`: if
  the upload directory disappears while the server runs, the page says "No
  files uploaded yet" and nothing reaches the log. Log it.

---

## Publication

Around the first public release.

- [ ] **Add `SECURITY.md`**

  The README states the trust model; what is missing is the file GitHub shows
  when someone wants to report a problem. In the same shape as the author's
  other projects: how to report a vulnerability privately, what is in scope,
  and which versions get fixes.

- [ ] **Add a `-version` flag**

  Bug reports, and the `SECURITY.md` above, need a way to name the build that
  misbehaved. Stamp the version from `build.sh` via `-ldflags "-X main.version=
  ..."`, from the git tag or commit, with "dev" for a build outside a checkout.

---

## Robustness

- [ ] **Bind the port before printing the banner**

  `main` prints "chute running!", both URLs and the QR code, and only then
  calls `ListenAndServe`. When the port is taken, the user sees a convincing
  banner followed by "Server error: address already in use". Call `net.Listen`
  first, print only once it has succeeded, then hand the listener to
  `server.Serve`. As a side effect `-port 0` becomes useful, since the banner
  can show the port the kernel actually chose.

- [ ] **An interrupted upload must not leave a truncated file**

  `handleUpload` writes straight to the final name. If the connection drops
  midway, which is routine on a phone, the partial file stays on disk under
  its real name and looks complete in the listing. Write to a temporary name
  in the upload directory, rename into place when `io.Copy` finishes, and
  remove the temporary file on any failure. The listing must skip the
  temporary names so an upload in progress does not appear as a file.

- [ ] **Report upload failures the browser notices**

  The upload code in `template.go` sets `xhr.onload` only. A dropped
  connection or an aborted request fires `onerror` or `onabort`, neither of
  which is handled, so the status line stays at "Uploading: 43%" forever. Add
  both, and include the HTTP status in the failure message so a 500 and a
  network error can be told apart.

- [ ] **Create the upload directory with 0755**

  `os.MkdirAll(uploadDir, os.ModePerm)` asks for 0777 and relies on the umask
  to narrow it. Ask for 0755 directly, the conventional mode for a directory
  the owner writes and others only read.

- [ ] **Pick the LAN address from the default route**

  `getLocalIP` returns the first non-loopback IPv4 address it finds. On a
  machine with Docker, a VM bridge or a VPN, that is often an address of one of
  those interfaces, and the QR code then points at a host no phone can reach.
  The usual fix is to ask the kernel which interface the default route uses: a
  UDP `net.Dial` to a public address sends no packet but fills in the local
  address the route would use. Keep the current scan as the fallback for a
  machine with no default route.

---

## User experience

- [ ] **Clipboard paste upload**

  Listen for the `paste` event on `window` and upload any image in the payload.
  Screenshot, `Ctrl+V`, and it is on the phone, with no intermediate save to
  disk. This is one of the most common reasons to reach for chute at all.

- [ ] **QR code modal in the web UI**

  The QR code is only printed to the terminal, which does not help someone
  already looking at the web UI on a desktop browser. A small button in the
  header showing the same code on screen makes sharing the link to another
  person's phone a one-step operation.

- [ ] **Per-file delete**

  Files can currently only be removed from the host machine's filesystem. A
  delete control per row makes cleaning up after a transfer effortless. Gate it
  behind an optional `-read-only` flag for situations where the listing should
  not be mutable from the network.

- [ ] **Small template fixes**

  Three things in `template.go`, each a line or two. `Drag & Drop` has a bare
  `&`, which browsers tolerate but which should be `&amp;`. The drop zone's
  border colour is a hard-coded `#3b82f6` while every other colour comes from a
  custom property, so it is the one colour the theme cannot change. And the
  drop zone is a `div` with a click handler: it cannot be focused or activated
  from the keyboard. Give it `tabindex="0"`, `role="button"` and an Enter/Space
  handler, or make it a `label` for the file input.

---

## Tests

- [ ] **Replace the two tests that cannot fail**

  `TestGetLocalIP` asserts a non-empty string, but `getLocalIP` always returns
  one: `"localhost"` is its fallback. `TestSetupRouter` asserts a non-nil mux,
  which `http.NewServeMux` guarantees. Neither can catch a regression. Have the
  router test send a request down each of the three routes and check the
  status codes; have the IP test check that the result parses with
  `net.ParseIP` or equals `"localhost"`.

- [ ] **Cover the untested paths**

  No test exercises the `/uploads/` route, a POST that is not multipart (which
  must be a 400), several files in one request, or what an interrupted upload
  leaves behind. The `/uploads/` hardening and the interrupted-upload fix above
  should each land with their tests, and the other two are cheap to add at the
  same time.

---

## Conventions

- [ ] **Finish the conventions in `AGENTS.md`; re-wrap the Markdown**

  The line-length rule is recorded, but two others that hold in practice are
  not: errors are logged in full on the server and generic to the client;
  every test must be able to fail. Write both down. And the rule itself is not
  yet followed: `AGENTS.md` and `README.md` have lines up to 450 characters, so
  re-wrap both at 120. `build.sh` already keeps to 80, and `TODO.md` stays at
  80 as the recorded exception.

- [ ] **Refresh the indirect dependencies**

  The only direct dependency, `qrterminal`, is current, but the indirect ones
  are not: `golang.org/x/term` is at v0.13.0 from 2023, `golang.org/x/sys`,
  `go-colorable` and `go-isatty` are several releases behind. `govulncheck`
  reports nothing against them today, so this is hygiene rather than urgency.
  `go get -u ./... && go mod tidy`, then `./build.sh test`.

---

## Deferred

- **In-memory ETag cache**

  `generateDirETag` scans the upload directory on every `GET /`. Caching the
  result in memory and invalidating it on upload would remove that scan, but the
  cache would not observe files added or deleted outside chute, and the
  current cost is negligible at realistic file counts: a directory read plus a
  SHA-256 over the metadata, against the much larger cost of rendering and
  sending the page it avoids.

  Revisit only if the upload directory grows to many thousands of files.

- **Continuous integration**

  A GitHub Actions workflow running `./build.sh test` on every push is about ten
  lines. The author's other projects have none, and with a single maintainer
  running `./build.sh test` before every commit it would only repeat what was
  already done. Revisit when there are outside contributors, or when a second
  platform needs checking on every change.
