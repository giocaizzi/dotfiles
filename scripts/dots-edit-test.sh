#!/bin/sh
# Test the edit commands (agents-edit, aliases-edit, ssh-edit) in bash, zsh and pwsh
# (each where installed) against a throwaway copy of this source: dummy encrypted files
# under a throwaway age key, a temp destination, and a chezmoi wrapper pinning
# --config/--source/--destination/--persistent-state. The real key, encrypted files and
# home are never read or written. Run from the repo root (make test).
set -eu
# Pinned so the target-mode cases are deterministic.
umask 022

case $(uname -s) in MINGW* | MSYS* | CYGWIN*)
	echo "Skipping the edit command tests on Windows: CI runs them on Linux."
	exit 0
	;;
esac

real=$(command -v chezmoi)
T=$(mktemp -d)
trap 'rm -rf "$T"' EXIT
S=$T/src
mkdir "$T/bin"

"$real" age-keygen -o "$T/key.txt" 2>/dev/null
recipient=$(sed -n 's/^# public key: //p' "$T/key.txt")
cat >"$T/chezmoi.toml" <<EOF
encryption = "age"
[age]
identity = "$T/key.txt"
recipient = "$recipient"
[data]
gitUser = "test"
gitEmail = "test@example.invalid"
light = false
work = false
workEmail = ""
githubUser = ""
personalGitRemote = ""
bwsProjectId = ""
otelEndpoint = ""
harnesses = ["claude"]
EOF
cat >"$T/bin/chezmoi" <<EOF
#!/bin/sh
exec "$real" --config "$T/chezmoi.toml" --source "$S" --destination "$T/home" \
	--persistent-state "$T/state.boltdb" "\$@"
EOF
# Fake editor: logs the path it was given, then edits it per mode.
cat >"$T/bin/editor" <<EOF
#!/bin/sh
printf '%s\n' "\$2" >"$T/edited"
case \$1 in
fail) exit 1 ;;
text) printf '\nedited-marker\n' >>"\$2" ;;
slowtext) printf '\nedited-marker\n' >>"\$2" && sleep 3 ;;
mcp) printf 'zzedit = "https://edited-marker.invalid/mcp"\n' >>"\$2" ;;
badtmpl) printf '\n{{ .nope.nope }}\n' >>"\$2" ;;
alias) printf '\n[alias.zzedit]\ncmd = "echo edited-marker"\n' >>"\$2" ;;
badalias) printf '\n[alias.zzbad]\nposix = "true"\nshells = ["posix"]\n' >>"\$2" ;;
ssh) printf '\nHost zzedit\n  HostName edited-marker.invalid\n' >>"\$2" ;;
badssh) printf '\nNotAnOption yes\n' >>"\$2" ;;
badthenssh)
	if grep -q NotAnOption "\$2"; then
		sed '/NotAnOption/d' "\$2" >"\$2.x" && mv "\$2.x" "\$2"
		printf '\nHost zzedit\n  HostName edited-marker.invalid\n' >>"\$2"
	else
		printf '\nNotAnOption yes\n' >>"\$2"
	fi
	;;
esac
EOF
# Fake tmux for the live preview: split-window runs the pane command in the background
# (output in $T/pane), kill-pane stops it.
cat >"$T/bin/tmux" <<EOF
#!/bin/sh
case \$1 in
split-window)
	for a do cmd=\$a; done
	sh -c "\$cmd" >"$T/pane" 2>&1 &
	echo \$! >"$T/pane.pid"
	echo %1
	;;
kill-pane) kill "\$(cat "$T/pane.pid")" 2>/dev/null; echo killed >>"$T/pane.log" ;;
esac
EOF
chmod +x "$T/bin/chezmoi" "$T/bin/editor" "$T/bin/tmux"

# The source copy: no scripts, externals or removals, dummy encrypted files, a git repo.
mkdir "$S"
git ls-files -z | tar -cf - --null -T - | tar -xf - -C "$S"
rm -rf "$S/.chezmoiscripts" "$S/.chezmoiexternal.toml.tmpl" "$S/.chezmoiremove" "$S/.chezmoi.toml.tmpl"
enc() { printf '%b' "$2" | "$T/bin/chezmoi" encrypt >"$S/$1"; }
enc .aliases.private.toml.age '[alias.zzprivate]\ncmd = "echo private"\n'
enc .chezmoitemplates/agents/personal.md.age 'dummy personal layer\n'
enc .chezmoitemplates/agents/work.md.age 'dummy work layer\n'
enc private_dot_ssh/encrypted_private_config.age 'Host dummy\n  HostName dummy.invalid\n'
git -C "$S" init -q
git -C "$S" add -A
git -C "$S" -c user.name=test -c user.email=test@example.invalid -c commit.gpgsign=false \
	-c core.hooksPath=/dev/null commit -qm fixture
"$T/bin/chezmoi" execute-template <"$S/dot_config/shell/aliases.sh.tmpl" >"$T/aliases.sh"
"$T/bin/chezmoi" execute-template <"$S/dot_config/powershell/aliases.ps1.tmpl" >"$T/aliases.ps1"

pass=0
fail=0
# run <shell> <answers> <editor mode> <path: full|nofzf> <command...>: status in $st,
# output in $T/out. Resets the source and the destination first, then evals $pre. In
# pwsh, Write-Error is made terminating so a reported failure gives a non-zero exit.
run() {
	shell=$1 answers=$2 mode=$3 path=$4
	shift 4
	git -C "$S" checkout -q -- .
	rm -rf "${T:?}/home" "$T/edited" "$T/state.boltdb" "$T/pane" "$T/pane.log"
	# The parents a real home already has (apply of one target needs them).
	mkdir -p "$T/home/.claude" "$T/home/.config/shell" "$T/home/.ssh"
	eval "${pre:-}"
	p=$T/bin:$PATH
	[ "$path" = full ] || p=$T/bin:/usr/bin:/bin
	set +e
	# shellcheck disable=SC2016 # expanded by the shell under test
	case $shell in
	pwsh) printf '%b' "$answers" | TMUX=${live:-} EDITOR="$T/bin/editor $mode" PATH=$p "$pwsh" -NoProfile -Command ". '$T/aliases.ps1'; \$ErrorActionPreference = 'Stop'; $*" ;;
	*) printf '%b' "$answers" | TMUX=${live:-} EDITOR="$T/bin/editor $mode" PATH=$p "$shell" -c '. "$1"; shift; "$@"' sh "$T/aliases.sh" "$@" ;;
	esac >"$T/out" 2>&1
	st=$?
	set -e
}
# expect <description> <test...>
expect() {
	desc=$1
	shift
	if "$@"; then
		pass=$((pass + 1))
	else
		fail=$((fail + 1))
		echo "FAIL [$shell] $desc"
		sed 's/^/    /' "$T/out"
	fi
}
status() { test "$st" "$1" "$2"; }
has() { grep -q "$1" "$2" 2>/dev/null; }
lacks() { ! grep -q "$1" "$2" 2>/dev/null; }
armored() { [ "$(head -n 1 "$S/$1")" = '-----BEGIN AGE ENCRYPTED FILE-----' ]; }
unchanged() { git -C "$S" diff --quiet -- "$1"; }
# The decrypted temp file the editor got is gone, and no .new file is left.
cleaned() {
	f=$(cat "$T/edited")
	case $f in "$S"/*) ;; *) [ ! -e "$f" ] || return 1 ;; esac
	[ -z "$(find "$S" -name '*.new')" ]
}
usage() { status -ne 0 && has usage: "$T/out" && { [ "$shell" = pwsh ] || status -eq 2; }; }

suite() {
	run "$shell" 'y\n' text full agents-edit personal
	expect 'agents-edit personal: edit + apply' status -eq 0
	expect 'agents-edit personal: applied' has edited-marker "$T/home/.claude/CLAUDE.md"
	expect 'agents-edit personal: re-encrypted, armored' armored .chezmoitemplates/agents/personal.md.age
	expect 'agents-edit personal: temp removed' cleaned
	expect 'agents-edit personal: umask restored for apply' test -n "$(find "$T/home/.claude/CLAUDE.md" -perm 0644)"
	expect 'agents-edit: ship hint' has 'ship a PR' "$T/out"

	run "$shell" 'n\n' text full agents-edit core
	expect 'agents-edit core: edited, not applied' has edited-marker "$S/.chezmoitemplates/agents/core.md"
	expect 'agents-edit core: no target written' lacks edited-marker "$T/home/.claude/CLAUDE.md"

	run "$shell" 'n\n' noop full agents-edit work
	expect 'agents-edit work no-op: byte-identical' unchanged .chezmoitemplates/agents/work.md.age
	expect 'agents-edit work no-op: temp removed' cleaned
	expect 'agents-edit work no-op: diff in the default mode' lacks 100600 "$T/out"

	pre="'$T/bin/chezmoi' apply '$T/home/.claude/CLAUDE.md'"
	run "$shell" '' noop full agents-edit personal
	pre=
	expect 'agents-edit up to date: no apply prompt' lacks 'apply now' "$T/out"
	expect 'agents-edit up to date: says so' has 'up to date' "$T/out"

	run "$shell" 'n\n' fail full agents-edit personal
	expect 'editor failure: non-zero' status -ne 0
	expect 'editor failure: source unchanged' unchanged .chezmoitemplates/agents/personal.md.age
	expect 'editor failure: temp removed' cleaned

	run "$shell" '' noop full agents-edit bogus
	expect 'agents-edit bad argument: usage' usage

	run "$shell" '1\nn\n' text nofzf agents-edit
	expect 'agents-edit picker without fzf: core' has edited-marker "$S/.chezmoitemplates/agents/core.md"

	run "$shell" '9\n' noop nofzf aliases-edit
	expect 'aliases-edit picker without fzf, invalid choice: usage' usage

	run "$shell" '' noop full aliases-edit bogus
	expect 'aliases-edit bad argument: usage' usage

	run "$shell" 'y\n' alias full aliases-edit public
	expect 'aliases-edit public: edit + apply' status -eq 0
	expect 'aliases-edit public: applied' has 'alias zzedit=' "$T/home/.config/shell/aliases.sh"
	expect 'aliases-edit: new-shell reminder' has 'open a new shell' "$T/out"

	run "$shell" 'y\n' alias full aliases-edit private
	expect 'aliases-edit private: edit + apply' status -eq 0
	expect 'aliases-edit private: applied' has 'alias zzedit=' "$T/home/.config/shell/aliases.sh"
	expect 'aliases-edit private: armored' armored .aliases.private.toml.age
	expect 'aliases-edit private: temp removed' cleaned

	run "$shell" 'y\n' noop full aliases-edit private
	expect 'aliases-edit private no-op: byte-identical' unchanged .aliases.private.toml.age

	run "$shell" 'y\n' badalias full aliases-edit public
	expect 'aliases-edit lint failure: non-zero' status -ne 0
	expect 'aliases-edit lint failure: no apply prompt' lacks 'apply now' "$T/out"
	expect 'aliases-edit lint failure: nothing applied' lacks zzbad "$T/home/.config/shell/aliases.sh"

	run "$shell" 'y\n' ssh full ssh-edit config
	expect 'ssh-edit: edit + apply' status -eq 0
	expect 'ssh-edit: applied' has edited-marker "$T/home/.ssh/config"
	expect 'ssh-edit: armored' armored private_dot_ssh/encrypted_private_config.age
	expect 'ssh-edit: temp removed' cleaned

	run "$shell" 'y\n' noop full ssh-edit config
	expect 'ssh-edit no-op: byte-identical' unchanged private_dot_ssh/encrypted_private_config.age

	run "$shell" 'y\n' badssh full ssh-edit config
	expect 'ssh-edit syntax error: non-zero' status -ne 0
	expect 'ssh-edit syntax error: no apply prompt' lacks 'apply now' "$T/out"
	expect 'ssh-edit syntax error: nothing applied' lacks NotAnOption "$T/home/.ssh/config"
	expect 'ssh-edit syntax error: temp removed' cleaned

	run "$shell" '' ssh full ssh-edit work
	expect 'ssh-edit config.d: created' has edited-marker "$T/home/.ssh/config.d/work"
	expect 'ssh-edit config.d: source untouched' test -z "$(git -C "$S" status --short)"
	expect 'ssh-edit config.d: no apply prompt' lacks 'apply now' "$T/out"
	expect 'ssh-edit config.d: temp removed' cleaned

	run "$shell" '' noop full ssh-edit work
	expect 'ssh-edit config.d no-op: no file created' test ! -e "$T/home/.ssh/config.d/work"

	pre="mkdir -p '$T/home/.ssh/config.d' && echo 'Host zzold' >'$T/home/.ssh/config.d/work'"
	run "$shell" 'n\n' badssh full ssh-edit work
	pre=
	expect 'ssh-edit config.d syntax error, give up: non-zero' status -ne 0
	expect 'ssh-edit config.d syntax error, give up: live file unchanged' test "$(cat "$T/home/.ssh/config.d/work")" = 'Host zzold'
	expect 'ssh-edit config.d syntax error, give up: temp removed' cleaned

	run "$shell" '\n' badthenssh full ssh-edit work
	expect 'ssh-edit config.d syntax error, edit again: fixed' status -eq 0
	expect 'ssh-edit config.d syntax error, edit again: installed' has edited-marker "$T/home/.ssh/config.d/work"
	expect 'ssh-edit config.d syntax error, edit again: bad line gone' lacks NotAnOption "$T/home/.ssh/config.d/work"

	pre="mkdir -p '$T/home/.ssh/config.d' && : >'$T/home/.ssh/config.d/zzlocal'"
	run "$shell" '2\n' ssh nofzf ssh-edit
	pre=
	expect 'ssh-edit picker without fzf: config.d file' has edited-marker "$T/home/.ssh/config.d/zzlocal"

	run "$shell" '' noop full ssh-edit ../x
	expect 'ssh-edit path argument: usage' usage

	run "$shell" '' noop full ssh-edit a b
	expect 'ssh-edit two arguments: usage' usage

	run "$shell" '' noop full dots-edit --list
	expect 'dots-edit --list: entries' has 'aliases/private' "$T/out"

	run "$shell" '' noop full dots-edit --help
	expect 'dots-edit --help: usage' has 'usage: dots-edit' "$T/out"

	run "$shell" 'n\n' mcp full dots-edit mcp
	expect 'dots-edit mcp: edited' has edited-marker "$S/.chezmoidata/mcp.toml"
	expect 'dots-edit mcp: report per configuration' has 'darwin/personal: .*mcp' "$T/out"
	expect 'dots-edit mcp: identical lists grouped' has 'linux/personal: same as darwin/personal' "$T/out"
	expect 'dots-edit mcp: work machines unaffected' has 'darwin/work: nothing' "$T/out"
	expect 'dots-edit mcp: light not managed' has 'windows/light: nothing' "$T/out"

	run "$shell" 'y\n' badtmpl full dots-edit agents/env
	expect 'render failure: non-zero' status -ne 0
	expect 'render failure: reported' has 'render FAILED' "$T/out"
	expect 'render failure: no apply prompt' lacks 'apply now' "$T/out"

	run "$shell" '4\nn\n' text nofzf dots-edit "$T/home/.claude/CLAUDE.md"
	expect 'dots-edit <target>: tree node picked' has edited-marker "$S/.chezmoitemplates/agents/core.md"
	expect 'dots-edit <target>: nodes named' has '\[agents/personal\]' "$T/out"

	run "$shell" 'n\n' text full dots-edit dot_config/tmux/tmux.conf
	expect 'dots-edit <source file>: edited' has edited-marker "$S/dot_config/tmux/tmux.conf"

	run "$shell" '' noop full dots-edit no/such/thing
	expect 'dots-edit unknown argument: non-zero' status -ne 0
	expect 'dots-edit unknown argument: says so' has 'no registry entry' "$T/out"

	run "$shell" '' noop full dots-edit --bogus
	expect 'dots-edit bad flag: usage' usage

	run "$shell" '' noop full dots-edit --view "$T/home/.claude/CLAUDE.md"
	expect 'dots-edit --view <target>: tree' has 'include .chezmoitemplates/agents/core.md  \[agents/core\]' "$T/out"
	expect 'dots-edit --view <target>: rendered' has 'dummy personal layer' "$T/out"

	run "$shell" '' noop full dots-edit --view "$T/home/.claude/CLAUDE.md" --as windows,work
	expect 'dots-edit --view --as: other role' has 'dummy work layer' "$T/out"
	expect 'dots-edit --view --as: labelled' has 'rendered on windows/work' "$T/out"

	run "$shell" '' noop full dots-edit --view "$T/home/.claude/CLAUDE.md" --as plan9
	expect 'dots-edit --view --as unknown: non-zero' status -ne 0

	run "$shell" '' noop full dots-edit --view "$T/home/.claude/CLAUDE.md" --matrix
	expect 'dots-edit --view --matrix: differences as diffs' has '+dummy work layer' "$T/out"
	expect 'dots-edit --view --matrix: unmanaged listed' has 'not managed on: darwin/light' "$T/out"

	run "$shell" '' noop full dots-edit --view "$T/home/.config/tmux/tmux.conf" --matrix
	expect 'dots-edit --view --matrix: identical renders grouped' has 'rendered on darwin/personal, darwin/work, darwin/light, linux/personal' "$T/out"

	run "$shell" '' noop full dots-edit --view mcp
	expect 'dots-edit --view <entry>: reaches this machine' has 'this machine: ' "$T/out"

	live=1
	run "$shell" 'n\n' slowtext full dots-edit agents/core
	live=
	expect 'live preview: diff of the changed target' has '+edited-marker' "$T/pane"
	expect 'live preview: pane closed' has killed "$T/pane.log"
}

if PATH=/usr/bin:/bin command -v fzf >/dev/null 2>&1; then
	echo "fzf is in /usr/bin: the picker fallback cases will fail" >&2
fi
pwsh=$(command -v pwsh || true)
for shell in bash zsh pwsh; do
	if [ "$shell" = pwsh ] && [ -z "$pwsh" ] || ! command -v "$shell" >/dev/null 2>&1; then
		echo "Skipping the $shell edit command tests: $shell is not installed."
		continue
	fi
	suite
done
echo "edit commands: $pass passed, $fail failed"
[ "$fail" -eq 0 ]
