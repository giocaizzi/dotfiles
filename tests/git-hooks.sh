#!/bin/bash
# Tests for the global git hooks (home/dot_config/git/hooks): the pre-push secret scan and
# the dispatchers that keep a repo's own .git/hooks working under core.hooksPath.
# Everything runs in a temp dir with a stub gitleaks and gh. Run with bash (Git Bash too).
set -u

SRC="$(cd "$(dirname "$0")/.." && pwd)/home/dot_config/git/hooks"
T="$(mktemp -d)"
trap 'rm -rf "$T"' EXIT

export GIT_CONFIG_GLOBAL="$T/gitconfig"
export GIT_CONFIG_NOSYSTEM=1
: > "$GIT_CONFIG_GLOBAL"
export STUB_LOG="$T/stub.log"

# Deploy the hooks the way chezmoi does: drop the executable_ prefix, mark executable.
HOOKS="$T/hooks"
mkdir -p "$HOOKS"
for f in "$SRC"/executable_*; do
	cp "$f" "$HOOKS/$(basename "$f" | sed 's/^executable_//')"
done
chmod +x "$HOOKS"/*

# Stub gitleaks: fails when the scanned range holds LEAK_MARKER (default rules) or
# PRIVATE_MARKER (when run with --config); logs every call.
STUBS="$T/stubs"
mkdir -p "$STUBS"
cat > "$STUBS/gitleaks" <<'EOF'
#!/bin/sh
echo "gitleaks $*" >> "$STUB_LOG"
cfg= range=
while [ $# -gt 0 ]; do
	case $1 in
	--config) cfg=$2; shift ;;
	--log-opts) range=$2; shift ;;
	esac
	shift
done
marker=LEAK_MARKER
[ -n "$cfg" ] && marker=PRIVATE_MARKER
# shellcheck disable=SC2086
if git log -p $range 2>/dev/null | grep -q "$marker"; then exit 1; fi
exit 0
EOF
# Stub gh: prints $STUB_VISIBILITY, or fails when it is "fail".
cat > "$STUBS/gh" <<'EOF'
#!/bin/sh
echo "gh $*" >> "$STUB_LOG"
[ "$STUB_VISIBILITY" = fail ] && exit 1
echo "$STUB_VISIBILITY"
EOF
chmod +x "$STUBS"/*

# A PATH without gitleaks: symlinks to just the tools the hooks need.
NOGL="$T/nogl"
mkdir -p "$NOGL"
for tool in git sh cat rm mktemp dirname sed grep; do
	ln -s "$(command -v "$tool")" "$NOGL/$tool"
done

PASS=0
FAIL=0
check() { # check <description> <expected: 0|1> <actual exit>
	if [ "$2" = "$3" ]; then
		PASS=$((PASS + 1))
	else
		FAIL=$((FAIL + 1))
		echo "FAIL: $1 (expected exit $2, got $3)"
	fi
}
calls() { # number of lines in the stub log matching $1
	grep -c "$1" "$STUB_LOG" 2>/dev/null || true
}

ORIG_PATH="$PATH"
export PATH="$STUBS:$ORIG_PATH"
export STUB_VISIBILITY=public
unset GITLEAKS_PRIVATE_CONFIG

REPO=
new_repo() {
	rm -rf "$T/work" "$T/remote.git"
	: > "$STUB_LOG"
	git init -q --bare "$T/remote.git"
	git init -q -b main "$T/work"
	REPO="$T/work"
	git -C "$REPO" config user.name test
	git -C "$REPO" config user.email test@example.com
	git -C "$REPO" config core.hooksPath "$HOOKS"
	git -C "$REPO" remote add origin "$T/remote.git"
}
commit() { # commit <file> <content>
	printf '%s\n' "$2" > "$REPO/$1"
	git -C "$REPO" add "$1"
	git -C "$REPO" commit -q -m "add $1"
}
push() { git -C "$REPO" push -q "$@" > "$T/push.out" 2>&1; }

# --- clean push passes
new_repo
commit a.txt clean
push origin main
check "clean push passes" 0 $?
check "clean push scanned once" 1 "$(calls '^gitleaks')"

# --- marked push is blocked
new_repo
commit a.txt LEAK_MARKER
push origin main
check "marked push blocked" 1 $?
grep -q "pushed to refs/heads/main" "$T/push.out"
check "block message names the ref" 0 $?

# --- delete push skipped
new_repo
commit a.txt clean
push origin main
git -C "$REPO" push -q origin main:doomed > /dev/null 2>&1
: > "$STUB_LOG"
push origin :doomed
check "delete push passes" 0 $?
check "delete push not scanned" 0 "$(calls '^gitleaks')"

# --- new branch scans only the new commits
new_repo
commit a.txt LEAK_MARKER
git -C "$REPO" push -q --no-verify origin main > /dev/null 2>&1
git -C "$REPO" checkout -q -b feature
commit b.txt clean
push origin feature
check "new branch with an already pushed leak in its base passes" 0 $?
grep -q -- "--not --remotes=origin" "$STUB_LOG"
check "new-branch range excludes the remote's commits" 0 $?
commit c.txt LEAK_MARKER
push origin feature
check "update with a marked commit blocked" 1 $?
grep -q "$(git -C "$REPO" rev-parse HEAD~1)\.\.$(git -C "$REPO" rev-parse HEAD)" "$STUB_LOG"
check "update range is remote..local" 0 $?

# --- the repo's own hooks still run
new_repo
mkdir -p "$REPO/.git/hooks"
cat > "$REPO/.git/hooks/pre-commit" <<EOF
#!/bin/sh
echo pre-commit >> "$T/chain.log"
exit \${CHAIN_EXIT:-0}
EOF
cat > "$REPO/.git/hooks/pre-push" <<EOF
#!/bin/sh
echo "pre-push \$1" >> "$T/chain.log"
cat > "$T/chain.stdin"
exit \${CHAIN_EXIT:-0}
EOF
chmod +x "$REPO/.git/hooks/pre-commit" "$REPO/.git/hooks/pre-push"
: > "$T/chain.log"
commit a.txt clean
grep -q '^pre-commit$' "$T/chain.log"
check "repo pre-commit runs through the dispatcher" 0 $?
push origin main
check "push passes with a passing repo pre-push" 0 $?
grep -q '^pre-push origin$' "$T/chain.log"
check "repo pre-push runs with the same args" 0 $?
grep -q 'refs/heads/main' "$T/chain.stdin"
check "repo pre-push gets the original stdin" 0 $?
commit b.txt clean
CHAIN_EXIT=1 push origin main
check "failing repo pre-push blocks the push" 1 $?
CHAIN_EXIT=1 commit c.txt clean 2> /dev/null
check "failing repo pre-commit blocks the commit" 1 $?
# no repo hooks: dispatchers pass
new_repo
commit a.txt clean
check "commit without repo hooks passes" 0 $?

# --- missing gitleaks blocks
new_repo
commit a.txt clean
PATH="$NOGL" push origin main
check "missing gitleaks blocks the push" 1 $?
grep -q "gitleaks is not installed" "$T/push.out"
check "missing gitleaks prints the install hint" 0 $?

# --- private pass
PRIV="$T/private-names.toml"
echo "# stub" > "$PRIV"
new_repo
commit a.txt PRIVATE_MARKER
push origin main
check "private marker passes without GITLEAKS_PRIVATE_CONFIG" 0 $?
check "gh not asked without the config" 0 "$(calls '^gh')"

new_repo
commit a.txt PRIVATE_MARKER
GITLEAKS_PRIVATE_CONFIG="$PRIV" push origin main
check "private marker blocked on a public remote" 1 $?
grep -q "must not reach a public repo" "$T/push.out"
check "private block message" 0 $?

new_repo
commit a.txt PRIVATE_MARKER
STUB_VISIBILITY=private GITLEAKS_PRIVATE_CONFIG="$PRIV" push origin main
check "private marker passes on a private remote" 0 $?
check "no private pass on a private remote" 0 "$(calls -- '--config')"

new_repo
commit a.txt PRIVATE_MARKER
STUB_VISIBILITY=fail GITLEAKS_PRIVATE_CONFIG="$PRIV" push origin main
check "unknown visibility counts as public" 1 $?

new_repo
commit a.txt PRIVATE_MARKER
GITLEAKS_PRIVATE_CONFIG="$T/missing.toml" push origin main
check "unreadable private config is skipped" 0 $?

echo "git-hooks: $PASS passed, $FAIL failed"
[ "$FAIL" = 0 ]
