#!/bin/sh
# SPDX-License-Identifier: ISC

# chute build script.
#
#   ./build.sh              build the release binary (the default)
#   ./build.sh release      same thing, spelled out
#   ./build.sh debug        unoptimised, with the race detector
#   ./build.sh run [args]   build debug, then run it with any extra arguments
#   ./build.sh test [name]  build, then gofmt, go vet and go test -race
#                           (optionally just the test called name)
#   ./build.sh install      build, then copy the binary to $PREFIX/bin
#   ./build.sh uninstall    remove it again
#   ./build.sh clean        remove bin/
#
# PREFIX defaults to ~/.local, the usual per-user prefix. Override it for
# anywhere else:
#     PREFIX=/usr/local ./build.sh install     (needs root)
#
# DESTDIR stages an install for a package: the binary goes under it, and PATH
# is not checked, since a staging directory is never on it.
#     DESTDIR=/tmp/stage PREFIX=/usr ./build.sh install
# Both must be absolute: the script works from its own directory.
#
# Cross-compiling uses Go's own variables, and the binary lands in
# bin/$GOOS-$GOARCH/ rather than bin/release/:
#     GOOS=windows GOARCH=arm64 ./build.sh
# Only a release build can target another platform; the other commands need a
# binary that runs here.

set -eu

# Work from the script's own directory, resolved through symlinks, so this can
# be run from any cwd, by absolute path, or through a symlink in ~/bin.
cd "$(dirname "$(readlink -f "$0")")"

PROGRAM="chute"

# ------------------------------------------------------------------- dispatch

COMMAND="${1:-release}"

case "$COMMAND" in
	clean)
		# -dir defaults to ./uploads in the current directory, so a binary
		# started from inside bin/ keeps whatever it receives there. Those are
		# someone's files, not build output: refuse rather than delete them.
		if [ -d bin ]; then
			strays=$(find bin -type f ! -name "$PROGRAM" ! -name "$PROGRAM.exe")

			if [ -n "$strays" ]; then
				echo "$0: bin/ holds files the build did not make:" >&2
				printf '%s\n' "$strays" | head -5 | sed 's/^/    /' >&2
				echo "Move them out first, or remove bin/ by hand" \
					"if they can go." >&2
				exit 1
			fi
		fi

		rm -rf bin
		echo "cleaned"
		exit 0
		;;
	uninstall)
		bindir="${DESTDIR:-}${PREFIX:-${HOME:-}/.local}/bin"
		target="$bindir/$PROGRAM"

		if [ ! -e "$target" ] && [ ! -L "$target" ]; then
			echo "nothing installed at $target"
			exit 0
		fi

		# Diagnose an unwritable directory rather than leaving rm(1) to fail
		# with a bare "Permission denied" and no way forward.
		if [ ! -w "$bindir" ]; then
			echo "$0: $bindir is not writable." >&2
			echo "Remove it as root, for example:" >&2
			echo "    su -c \"rm -f '$target'\"" >&2
			exit 1
		fi

		rm -f "$target"
		echo "removed $target"

		exit 0
		;;
	release)
		BUILD="release"
		;;
	debug)
		BUILD="debug"
		;;
	run)
		BUILD="debug"
		;;
	test)
		BUILD="release"
		;;
	install)
		BUILD="release"
		;;
	*)
		echo "usage: $0" \
			"[release|debug|run [args...]|test [name]|install|uninstall|clean]" >&2
		exit 2
		;;
esac

if ! command -v go > /dev/null 2>&1; then
	echo "$0: go not found on PATH." >&2
	echo "See Requirements in README.md." >&2
	exit 1
fi

# has_ CMD -- CMD is installed. Only the first word is looked up: CC may carry
# arguments, as in CC="zig cc".
has_()
{
	command -v "${1%% *}" > /dev/null 2>&1
}

if [ "$BUILD" = "debug" ] || [ "$COMMAND" = "test" ]; then
	# The race detector needs cgo, and cgo needs a C compiler. $CC when set;
	# otherwise Go's own default (gcc on most systems, clang on some, or what
	# `go env -w CC=` chose), then whichever of gcc and clang is installed. A
	# release build needs no C compiler.
	if [ -z "${CC:-}" ]; then
		for cc in "$(go env CC)" gcc clang; do
			if has_ "$cc"; then
				CC="$cc"
				break
			fi
		done
	fi

	if [ -z "${CC:-}" ] || ! has_ "$CC"; then
		echo "$0: no C compiler found${CC:+: $CC is not installed}." >&2
		echo "The race detector needs cgo, which needs a C compiler." >&2
		echo "Install gcc or clang, or point CC at one:" \
			"CC=/path/to/cc $0 $COMMAND" >&2
		exit 1
	fi
fi

# go env reports GOOS and GOARCH from the environment when they are set, and
# the host's otherwise, so a mismatch with GOHOSTOS/GOHOSTARCH is exactly a
# cross build.
goos=$(go env GOOS)
goarch=$(go env GOARCH)
hostos=$(go env GOHOSTOS)
hostarch=$(go env GOHOSTARCH)

if [ "$goos/$goarch" != "$hostos/$hostarch" ]; then
	if [ "$COMMAND" != "release" ]; then
		echo "$0: '$COMMAND' needs a binary that runs on this machine" \
			"($hostos/$hostarch)," >&2
		echo "but GOOS/GOARCH select $goos/$goarch." \
			"Only a release build can cross-compile." >&2
		exit 2
	fi

	BUILD="$goos-$goarch"
fi

OUT="bin/$BUILD/$PROGRAM$(go env GOEXE)"

mkdir -p "bin/$BUILD"

# Delete the output first. go build leaves the old binary untouched when a
# compile fails -- same contents, same mtime -- so without this a failed build
# silently leaves you running the previous one.
rm -f "$OUT"

# CGO_ENABLED=0 is pinned, not inherited: Go turns cgo off by itself only when
# it finds no C compiler, and with one present net links against libc and the
# release binary quietly turns dynamic. The flags are spelled out here rather
# than collected in a variable: '-s -w' and 'all=-N -l' each contain a space,
# and a word-split variable cannot carry one.
if [ "$BUILD" = "debug" ]; then
	CC="$CC" CGO_ENABLED=1 go build -race -gcflags='all=-N -l' -o "$OUT" .
else
	CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o "$OUT" .
fi

echo "built $PWD/$OUT"

if [ "$COMMAND" = "run" ]; then
	# Runs from the project root, so the default -dir puts uploads in
	# ./uploads here, never inside bin/ where clean would refuse to run.
	shift
	exec "./$OUT" "$@"
fi

if [ "$COMMAND" = "install" ]; then
	# --------------------------------------------------------------- install
	#
	# Installs a COPY, not a symlink: the installed binary must not change under
	# you when you rebuild for development, and `./build.sh clean` must not be
	# able to leave a dangling link where your file server used to be.
	bindir="${DESTDIR:-}${PREFIX:-$HOME/.local}/bin"

	if ! mkdir -p "$bindir"; then
		echo "$0: cannot create $bindir" >&2
		exit 1
	fi

	# Normalise to an absolute, symlink-resolved path so the messages below and
	# the PATH comparison all speak about the same string. Without this a
	# trailing slash in PREFIX gives "$HOME/.local//bin", which matches nothing
	# in PATH and warns about a directory that is perfectly well on it.
	bindir=$(cd "$bindir" && pwd -P)
	target="$bindir/$PROGRAM"

	if [ ! -w "$bindir" ]; then
		echo "$0: $bindir is not writable." >&2
		echo "The build above succeeded, so copy it as root, for example:" >&2
		echo "    su -c \"install -T -m 755 '$PWD/$OUT' '$target'\"" >&2
		echo "Do not re-run this script under plain su: su resets HOME," \
			"so go would use root's" >&2
		echo "caches, fetch every module again," \
			"and leave a root-owned binary in bin/." >&2
		exit 1
	fi

	# -T, so that a directory sitting at $target is an error rather than being
	# installed INTO while the script reports success for a path that is not the
	# binary.
	install -T -m 755 "$OUT" "$target"
	echo "installed $target"

	# A staging directory is never on PATH, so the check below would only put
	# a false warning into every package build.
	if [ -n "${DESTDIR:-}" ]; then
		exit 0
	fi

	# A prefix's bin is not necessarily on PATH (~/bin often is not) -- an easy
	# way to install something and then wonder why the shell cannot find it.
	# Both sides are resolved before comparing, so a different spelling of the
	# same directory does not produce a false warning.
	on_path=$(
		IFS=:
		for dir in $PATH; do
			if [ -z "$dir" ]; then
				dir=.
			fi

			if resolved=$(cd "$dir" 2> /dev/null && pwd -P); then
				if [ "$resolved" = "$bindir" ]; then
					echo yes
					exit 0
				fi
			fi
		done

		echo no
	)

	if [ "$on_path" != yes ]; then
		echo "warning: $bindir is not on your PATH," \
			"so '$PROGRAM' will not be found by name." >&2
		echo "         Add it, or run it as $target" >&2
	fi

	exit 0
fi

if [ "$COMMAND" != "test" ]; then
	exit 0
fi

# ----------------------------------------------------------------------- test
#
# The gate, in order, stopping at the first failure: every file gofmt-clean,
# go vet clean, then the tests under the race detector.

echo

unformatted=$(gofmt -l .)

if [ -n "$unformatted" ]; then
	echo "not gofmt-clean:" >&2
	printf '%s\n' "$unformatted" | sed 's/^/    /' >&2
	echo "Run: go fmt ./..." >&2
	exit 1
fi

CGO_ENABLED=0 go vet ./...

shift || true
selection="${1:-}"

# Deadline for each package's test binary; go test's own default is 10m. A bug
# that stops the server shutting down hangs a test, and a wedged test must
# fail, not stall the suite. Override for a slow machine:
#     TEST_TIMEOUT=120 ./build.sh test
TEST_TIMEOUT="${TEST_TIMEOUT:-60}"

# Anchored, so a name selects that test alone rather than every test it is a
# prefix of. TestName/subtest works too: go test splits the pattern on '/'.
run_pattern="."

if [ -n "$selection" ]; then
	run_pattern="^$selection\$"
fi

# -count=1, so the tests run rather than replay a cached pass.
#
# Under `set -e` a bare assignment from a failing command substitution would
# abort before the output is printed, so the status is caught explicitly.
if output=$(CC="$CC" CGO_ENABLED=1 go test -race -count=1 \
	-timeout "${TEST_TIMEOUT}s" -run "$run_pattern" ./... 2>&1); then
	status=0
else
	status=$?
fi

printf '%s\n' "$output"

# go test passes when -run matches nothing, marking each package
# "[no tests to run]". A misspelt name must not read as a pass.
if [ "$status" -eq 0 ] && [ -n "$selection" ]; then
	ran=$(printf '%s\n' "$output" | grep '^ok' |
		grep -cv '\[no tests to run\]' || true)

	if [ "$ran" -eq 0 ]; then
		echo "no test named '$selection'" >&2
		exit 2
	fi
fi

exit "$status"
