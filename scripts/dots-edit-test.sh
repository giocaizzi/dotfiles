#!/bin/sh
# Test the dots edit command in bash, zsh,
# pwsh and, on Windows (Git Bash), Windows PowerShell 5.1 (each where installed) against a
# throwaway copy of this source: dummy encrypted files under a throwaway age key, a temp
# destination, and a chezmoi wrapper pinning --config/--source/--destination/
# --persistent-state. The real key, encrypted files and home are never read or written.
# Run from the repo root (make test).
set -eu
# Pinned so the target-mode cases are deterministic.
umask 022

# Windows: every path in C:/x form, which Git Bash and PowerShell both read; PowerShell
# runs the wrappers as .cmd shims over the sh ones.
win=
case $(uname -s) in MINGW* | MSYS* | CYGWIN*) win=1 ;; esac

real=$(command -v chezmoi)
jq=$(command -v jq)
T=$(mktemp -d)
if [ -n "$win" ]; then T=$(cygpath -m "$T"); fi
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
# Keep jq available when the no-fzf cases remove its shared WinGet-shim directory.
cat >"$T/bin/jq" <<EOF
#!/bin/sh
exec "$jq" "\$@"
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
# Fake Windows Terminal: wt -w 0 split-pane -V <command> runs the command in the
# background (output in $T/pane); the pane ends itself.
cat >"$T/bin/wt" <<EOF
#!/bin/sh
shift 4
"\$@" >"$T/pane" 2>&1 &
EOF
chmod +x "$T/bin/chezmoi" "$T/bin/editor" "$T/bin/jq" "$T/bin/tmux" "$T/bin/wt"
if [ -n "$win" ]; then
	sh=$(cygpath -w "$(command -v sh)")
	printf '@"%s" --config "%s" --source "%s" --destination "%s" --persistent-state "%s" %%*\r\n' \
		"$(cygpath -w "$real")" "$T/chezmoi.toml" "$S" "$T/home" "$T/state.boltdb" >"$T/bin/chezmoi.cmd"
	for f in editor wt; do printf '@"%s" "%s" %%*\r\n' "$sh" "$T/bin/$f" >"$T/bin/$f.cmd"; done
fi

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
# PowerShell, Write-Error is made terminating so a reported failure gives a non-zero exit.
# $live set: the live pane runs (tmux; in PowerShell on Windows, Windows Terminal).
run() {
	shell=$1 answers=$2 mode=$3 path=$4
	shift 4
	git -C "$S" checkout -q -- .
	rm -rf "${T:?}/home" "$T/edited" "$T/state.boltdb" "$T/pane" "$T/pane.log"
	# The parents a real home already has (apply of one target needs them).
	mkdir -p "$T/home/.claude" "$T/home/.config/shell" "$T/home/.config/powershell" "$T/home/.ssh"
	eval "${pre:-}"
	# PATH needs the /c/x form: the colon of C:/x would split it.
	bin=$T/bin
	if [ -n "$win" ]; then bin=$(cygpath -u "$bin"); fi
	p=$bin:$PATH
	if [ "$path" = nofzf ]; then
		p=$bin:/usr/bin:/bin
	fi
	editor=$T/bin/editor
	if [ -n "$win" ]; then editor=$T/bin/editor.cmd; fi
	set +e
	# shellcheck disable=SC2016 # expanded by the shell under test
	case $shell in
	pwsh | powershell)
		answer_input=$(printf '%b' "$answers")
		DOTS_EDIT_TEST_ANSWERS="$answer_input" TMUX=${live:-} WT_SESSION=${live:-} EDITOR="$editor $mode" PATH=$p "$(command -v "$shell")" \
			-NoProfile -ExecutionPolicy Bypass -Command "\
				\$global:dotsEditTestAnswers = [System.Collections.Queue]::new(); \
				foreach (\$answer in (\$env:DOTS_EDIT_TEST_ANSWERS -split [System.Environment]::NewLine)) { [void]\$global:dotsEditTestAnswers.Enqueue(\$answer) }; \
				function global:Read-Host { param([string]\$Prompt); if (\$global:dotsEditTestAnswers.Count) { \$global:dotsEditTestAnswers.Dequeue() } else { '' } }; \
				. '$T/aliases.ps1'; \$ErrorActionPreference = 'Stop'; $*"
		;;
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
	if [ -n "$win" ]; then f=$(cygpath -m "$f"); fi
	case $f in "$S"/*) ;; *) [ ! -e "$f" ] || return 1 ;; esac
	[ -z "$(find "$S" -name '*.new')" ]
}
in_powershell() { [ "$shell" = pwsh ] || [ "$shell" = powershell ]; }
usage() { status -ne 0 && has usage: "$T/out" && { in_powershell || status -eq 2; }; }

suite() {
	run "$shell" 'y\n' text full dots agents/personal
	expect 'dots agents/personal: edit + apply' status -eq 0
	expect 'dots agents/personal: applied' has edited-marker "$T/home/.claude/CLAUDE.md"
	expect 'dots agents/personal: re-encrypted, armored' armored .chezmoitemplates/agents/personal.md.age
	expect 'dots agents/personal: temp removed' cleaned
	if [ -z "$win" ]; then
		expect 'dots agents/personal: umask restored for apply' test -n "$(find "$T/home/.claude/CLAUDE.md" -perm 0644)"
	fi
	expect 'dots: ship hint' has 'ship a PR' "$T/out"

	run "$shell" 'n\n' text full dots edit agents/core
	expect 'dots edit agents/core: edited, not applied' has edited-marker "$S/.chezmoitemplates/agents/core.md"
	expect 'dots edit agents/core: no target written' lacks edited-marker "$T/home/.claude/CLAUDE.md"

	run "$shell" 'n\n' noop full dots agents/work
	expect 'dots agents/work no-op: byte-identical' unchanged .chezmoitemplates/agents/work.md.age
	expect 'dots agents/work no-op: temp removed' cleaned
	expect 'dots agents/work no-op: diff in the default mode' lacks 100600 "$T/out"

	pre="'$T/bin/chezmoi' apply '$T/home/.claude/CLAUDE.md'"
	run "$shell" '' noop full dots agents/personal
	pre=
	expect 'dots up to date: no apply prompt' lacks 'apply now' "$T/out"
	expect 'dots up to date: says so' has 'up to date' "$T/out"

	run "$shell" 'n\n' fail full dots agents/personal
	expect 'editor failure: non-zero' status -ne 0
	expect 'editor failure: source unchanged' unchanged .chezmoitemplates/agents/personal.md.age
	expect 'editor failure: temp removed' cleaned

	if ! in_powershell; then
		run "$shell" '1\nn\n' text nofzf dots
		expect 'dots picker without fzf: first entry' has edited-marker "$S/.chezmoitemplates/agents/core.md"

		run "$shell" '99\n' noop nofzf dots
		expect 'dots picker without fzf, invalid choice: non-zero' status -ne 0
	fi

	run "$shell" 'y\n' alias full dots aliases/public
	expect 'dots aliases/public: edit + apply' status -eq 0
	expect 'dots aliases/public: applied' has 'alias zzedit=' "$T/home/.config/shell/aliases.sh"
	expect 'dots aliases/public: new-shell reminder' has 'open a new shell' "$T/out"

	run "$shell" 'y\n' alias full dots aliases/private
	expect 'dots aliases/private: edit + apply' status -eq 0
	expect 'dots aliases/private: applied' has 'alias zzedit=' "$T/home/.config/shell/aliases.sh"
	expect 'dots aliases/private: armored' armored .aliases.private.toml.age
	expect 'dots aliases/private: temp removed' cleaned

	run "$shell" 'y\n' noop full dots aliases/private
	expect 'dots aliases/private no-op: byte-identical' unchanged .aliases.private.toml.age

	run "$shell" 'y\n' badalias full dots aliases/public
	expect 'dots aliases lint failure: non-zero' status -ne 0
	expect 'dots aliases lint failure: no apply prompt' lacks 'apply now' "$T/out"
	expect 'dots aliases lint failure: nothing applied' lacks zzbad "$T/home/.config/shell/aliases.sh"

	run "$shell" 'y\n' ssh full dots ssh/config
	expect 'dots ssh/config: edit + apply' status -eq 0
	expect 'dots ssh/config: applied' has edited-marker "$T/home/.ssh/config"
	expect 'dots ssh/config: armored' armored private_dot_ssh/encrypted_private_config.age
	expect 'dots ssh/config: temp removed' cleaned

	run "$shell" 'y\n' noop full dots ssh/config
	expect 'dots ssh/config no-op: byte-identical' unchanged private_dot_ssh/encrypted_private_config.age

	run "$shell" 'y\n' badssh full dots ssh/config
	expect 'dots ssh/config syntax error: non-zero' status -ne 0
	expect 'dots ssh/config syntax error: no apply prompt' lacks 'apply now' "$T/out"
	expect 'dots ssh/config syntax error: nothing applied' lacks NotAnOption "$T/home/.ssh/config"
	expect 'dots ssh/config syntax error: temp removed' cleaned

	run "$shell" '' ssh full dots "$T/home/.ssh/config.d/work"
	expect 'dots local: created' has edited-marker "$T/home/.ssh/config.d/work"
	expect 'dots local: source untouched' test -z "$(git -C "$S" status --short)"
	expect 'dots local: no apply prompt' lacks 'apply now' "$T/out"
	expect 'dots local: temp removed' cleaned

	run "$shell" '' noop full dots "$T/home/.ssh/config.d/work"
	expect 'dots local no-op: no file created' test ! -e "$T/home/.ssh/config.d/work"

	pre="mkdir -p '$T/home/.ssh/config.d' && echo 'Host zzold' >'$T/home/.ssh/config.d/work'"
	run "$shell" 'n\n' badssh full dots "$T/home/.ssh/config.d/work"
	pre=
	expect 'dots local check failure, give up: non-zero' status -ne 0
	expect 'dots local check failure, give up: live file unchanged' test "$(cat "$T/home/.ssh/config.d/work")" = 'Host zzold'
	expect 'dots local check failure, give up: temp removed' cleaned

	run "$shell" '\n' badthenssh full dots "$T/home/.ssh/config.d/work"
	expect 'dots local check failure, edit again: fixed' status -eq 0
	expect 'dots local check failure, edit again: installed' has edited-marker "$T/home/.ssh/config.d/work"
	expect 'dots local check failure, edit again: bad line gone' lacks NotAnOption "$T/home/.ssh/config.d/work"

	if ! in_powershell; then
		pre="mkdir -p '$T/home/.ssh/config.d' && : >'$T/home/.ssh/config.d/zzlocal'"
		run "$shell" '1\n' ssh nofzf dots ssh/local
		pre=
		expect 'dots local entry, picker without fzf: file edited' has edited-marker "$T/home/.ssh/config.d/zzlocal"
	fi

	run "$shell" '' noop full dots ssh/local
	expect 'dots local entry, empty folder: non-zero' status -ne 0
	expect 'dots local entry, empty folder: says how to create' has 'create a file' "$T/out"

	run "$shell" '' noop full dots "$T/home/.ssh/config.d/sub/x"
	expect 'dots local nested path: usage' usage

	run "$shell" '' noop full dots a b
	expect 'dots two arguments: usage' usage

	run "$shell" '' noop full dots list
	expect 'dots list: entries' has 'aliases/private' "$T/out"
	expect 'dots list: local entry' has 'ssh/local' "$T/out"

	run "$shell" '' noop full dots help
	expect 'dots help: usage' has 'usage: dots' "$T/out"

	run "$shell" 'n\n' mcp full dots mcp
	expect 'dots mcp: edited' has edited-marker "$S/.chezmoidata/mcp.toml"
	expect 'dots mcp: report per configuration' has 'darwin/personal: .*mcp' "$T/out"
	expect 'dots mcp: identical lists grouped' has 'linux/personal: same as darwin/personal' "$T/out"
	expect 'dots mcp: work machines unaffected' has 'darwin/work: nothing' "$T/out"
	expect 'dots mcp: light not managed' has 'windows/light: nothing' "$T/out"

	run "$shell" 'y\n' badtmpl full dots agents/env
	expect 'render failure: non-zero' status -ne 0
	expect 'render failure: reported' has 'render FAILED' "$T/out"
	expect 'render failure: no apply prompt' lacks 'apply now' "$T/out"

	if ! in_powershell; then
		run "$shell" '4\nn\n' text nofzf dots "$T/home/.claude/CLAUDE.md"
		expect 'dots <target>: tree node picked' has edited-marker "$S/.chezmoitemplates/agents/core.md"
		expect 'dots <target>: nodes named' has '\[agents/personal\]' "$T/out"
	fi

	run "$shell" 'n\n' text full dots dot_config/tmux/tmux.conf
	expect 'dots <source file>: edited' has edited-marker "$S/dot_config/tmux/tmux.conf"

	run "$shell" '' noop full dots no/such/thing
	expect 'dots unknown argument: non-zero' status -ne 0
	expect 'dots unknown argument: says so' has 'no registry entry' "$T/out"

	run "$shell" '' noop full dots --bogus
	expect 'dots bad flag: usage' usage

	run "$shell" '' noop full dots view "$T/home/.claude/CLAUDE.md"
	expect 'dots view <target>: tree' has 'include .chezmoitemplates/agents/core.md  \[agents/core\]' "$T/out"
	expect 'dots view <target>: rendered' has 'dummy personal layer' "$T/out"

	run "$shell" '' noop full dots view "$T/home/.claude/CLAUDE.md" --as windows,work
	expect 'dots view --as: other role' has 'dummy work layer' "$T/out"
	expect 'dots view --as: labelled' has 'rendered on windows/work' "$T/out"

	run "$shell" '' noop full dots view "$T/home/.claude/CLAUDE.md" --as plan9
	expect 'dots view --as unknown: non-zero' status -ne 0

	run "$shell" '' noop full dots view "$T/home/.claude/CLAUDE.md" --matrix
	expect 'dots view --matrix: differences as diffs' has '+dummy work layer' "$T/out"
	expect 'dots view --matrix: unmanaged listed' has 'not managed on: darwin/light' "$T/out"

	run "$shell" '' noop full dots view "$T/home/.config/tmux/tmux.conf" --matrix
	expect 'dots view --matrix: identical renders grouped' has 'rendered on darwin/personal, darwin/work, darwin/light, linux/personal' "$T/out"

	run "$shell" '' noop full dots view mcp
	expect 'dots view <entry>: reaches this machine' has 'this machine: ' "$T/out"

	live=1
	run "$shell" 'n\n' slowtext full dots agents/core
	live=
	# A failure report shows the pane too.
	{ echo '--- pane'; cat "$T/pane" 2>/dev/null; } >>"$T/out"
	expect 'live preview: diff of the changed target' has '+edited-marker' "$T/pane"
	# Windows Terminal has no kill-pane: there the pane ends itself on the stop file.
	if [ -z "$win" ] || ! in_powershell; then
		expect 'live preview: pane closed' has killed "$T/pane.log"
	fi
}

if [ -z "$win" ] && PATH=/usr/bin:/bin command -v fzf >/dev/null 2>&1; then
	echo "fzf is in /usr/bin: the picker fallback cases will fail" >&2
fi
# Windows PowerShell 5.1 exists only on Windows; zsh is no Windows dependency.
# A supplied shell list lets make run the independent fixtures concurrently.
shells=$*
if [ -z "$shells" ]; then
	shells='bash zsh pwsh'
	if [ -n "$win" ]; then shells='bash powershell pwsh'; fi
fi
for shell in $shells; do
	if ! command -v "$shell" >/dev/null 2>&1; then
		echo "Skipping the $shell edit command tests: $shell is not installed."
		continue
	fi
	suite
done
echo "edit commands: $pass passed, $fail failed"
[ "$fail" -eq 0 ]
