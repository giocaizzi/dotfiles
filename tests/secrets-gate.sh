#!/bin/bash
# Tests for the env-secrets gate in home/.chezmoiignore: chezmoi manages
# ~/.config/shell/secrets.sh only where this machine has a bws project ID and an installed
# bws and chezmoi-dots that can run it (`chezmoi-dots secrets run`, 0.4.0 or later).
# Read-only: `chezmoi managed` against this checkout, with a temp HOME, config and stub
# binaries; nothing is applied and no secret manager is called. Needs chezmoi; skipped on
# Windows, where the gate stats chezmoi-dots.exe and a stub script cannot stand in for it.
set -u

if ! command -v chezmoi >/dev/null 2>&1; then
	echo "secrets-gate: chezmoi is required" >&2
	exit 1
fi
case "$(uname -s)" in
MINGW* | MSYS* | CYGWIN*)
	echo "secrets-gate: skipped on Windows"
	exit 0
	;;
esac

SRC="$(cd "$(dirname "$0")/.." && pwd)/home"
T="$(mktemp -d)"
trap 'rm -rf "$T"' EXIT

pass=0 fail=0 xfail=0

# managed <case> <bwsProjectId> <bws: yes|no> <chezmoi-dots version, "fail" or "none">:
# prints yes when chezmoi manages .config/shell/secrets.sh for that machine state, no when
# it does not, error when chezmoi fails.
managed() {
	local home="$T/$1/home" cfg="$T/$1/chezmoi.toml"
	mkdir -p "$home/.local/bin"
	if [ "$3" = yes ]; then
		printf '#!/bin/sh\necho "bws stub: must not run" >&2\nexit 9\n' > "$home/.local/bin/bws"
		chmod +x "$home/.local/bin/bws"
	fi
	if [ "$4" = fail ]; then
		# A chezmoi-dots whose version command fails.
		printf '#!/bin/sh\necho "broken" >&2\nexit 1\n' > "$home/.local/bin/chezmoi-dots"
		chmod +x "$home/.local/bin/chezmoi-dots"
	elif [ "$4" != none ]; then
		# Like chezmoi-dots 0.3.0, an unknown command is a usage error (exit 2).
		# shellcheck disable=SC2016 # $1 belongs to the stub, not to this script
		printf '#!/bin/sh\n[ "$1" = version ] && { echo "chezmoi-dots %s"; exit 0; }\nexit 2\n' "$4" \
			> "$home/.local/bin/chezmoi-dots"
		chmod +x "$home/.local/bin/chezmoi-dots"
	fi
	cat > "$cfg" <<EOF
[data]
    gitUser = "Octo Cat"
    gitEmail = "octocat@example.com"
    light = false
    work = false
    terraform = false
    workEmail = ""
    githubUser = ""
    personalGitRemote = ""
    bwsProjectId = "$2"
    otelEndpoint = ""
    harnesses = []
EOF
	# dotsRender: the release externals render empty, so nothing is downloaded.
	# A render that fails (chezmoi exits non-zero) is "error", never a pass.
	if ! HOME="$home" XDG_CONFIG_HOME="$T/$1/xdg" XDG_CACHE_HOME="$T/$1/cache" XDG_STATE_HOME="$T/$1/state" \
		chezmoi --config "$cfg" --source "$SRC" --destination "$home" --override-data '{"dotsRender":true}' \
		managed --include=files > "$T/$1.out" 2> "$T/$1.err"; then
		echo error
	elif grep -qx '.config/shell/secrets.sh' "$T/$1.out"; then
		echo yes
	else
		echo no
	fi
}

# check <case> <want> <got> [xfail <reason>]: strict expected failure, like pytest's
# xfail(strict=True): a case marked xfail that passes fails the run, so the marker is removed
# together with the fix.
check() {
	local name=$1 want=$2 got=$3 mark=${4:-} reason=${5:-}
	if [ "$mark" = xfail ]; then
		if [ "$got" = "$want" ]; then
			echo "XPASS (strict) $name: remove the xfail marker"
			fail=$((fail + 1))
		else
			echo "xfail $name: $reason"
			xfail=$((xfail + 1))
		fi
		return
	fi
	if [ "$got" = "$want" ]; then
		pass=$((pass + 1))
	else
		echo "FAIL $name: managed=$got, want $want"
		[ -s "$T/$name.err" ] && sed 's/^/  /' "$T/$name.err"
		fail=$((fail + 1))
	fi
}

P=11111111-2222-3333-4444-555555555555
check no-project no "$(managed no-project "" yes 0.4.0)"
check no-bws no "$(managed no-bws "$P" no 0.4.0)"
check no-dots no "$(managed no-dots "$P" yes none)"
check dots-0.4.0 yes "$(managed dots-0.4.0 "$P" yes 0.4.0)"
check dots-0.5.0 yes "$(managed dots-0.5.0 "$P" yes 0.5.0)"
# A machine still on chezmoi-dots 0.3.0 (no `secrets run`) must keep the secrets file out
# of the apply until the external brings 0.4.0: otherwise every render fails and chezmoi
# apply stops at .config/shell/secrets.sh, before .local/bin/chezmoi-dots is updated.
check dots-0.3.0 no "$(managed dots-0.3.0 "$P" yes 0.3.0)"
check dots-0.10.0 yes "$(managed dots-0.10.0 "$P" yes 0.10.0)"
# A version that cannot be read skips the secrets file instead of failing the render.
check dots-version-fails no "$(managed dots-version-fails "$P" yes fail)"
check dots-garbage no "$(managed dots-garbage "$P" yes garbage)"
check dots-devel no "$(managed dots-devel "$P" yes devel-0123abc)"

echo "secrets-gate: $pass passed, $fail failed, $xfail xfail"
[ "$fail" -eq 0 ]
