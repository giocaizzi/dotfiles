#!/bin/sh
# dots-edit live preview pane (docs/OS.md#editing-your-dotfiles), run by dots-edit in a
# tmux side pane: live.sh <source file> <edited file> <snapshot dir> <chezmoi> <source dir>.
# On every save of the edited file it copies it (re-encrypted when the source file is an
# age file) into the private source copy <snapshot dir>/shadow, renders this machine's
# target state from that copy and shows the diff of each target that differs from
# <snapshot dir>/before/here.json. Paths into the copy read as the real source dir.
file=$1 edited=$2 dir=$3 cz=$4 src=$5
last=
while :; do
	sum=$(cksum <"$edited" 2>/dev/null)
	if [ "$sum" != "$last" ]; then
		last=$sum
		case $file in
		*.age) "$cz" encrypt "$edited" >"$dir/shadow/$file" ;;
		*) cp "$edited" "$dir/shadow/$file" ;;
		esac
		printf '\033[H\033[2J'
		echo "dots-edit live preview: $file on this machine (refreshes on save)"
		if "$cz" --source "$dir/shadow" dump --exclude dirs,externals --format json 2>"$dir/live.err" |
			jq --arg s "$dir/shadow" --arg r "$src" 'walk(if type == "string" then split($s) | join($r) else . end)' >"$dir/live.json"; then
			jq -rn --slurpfile a "$dir/before/here.json" --slurpfile b "$dir/live.json" \
				'$a[0] as $a | $b[0] as $b | [($a, $b) | keys[]] | unique[] | select($a[.] != $b[.])' | tr -d '\r' |
				while IFS= read -r t; do
					jq -r --arg t "$t" '.[$t] | .contents // .linkname // ""' "$dir/before/here.json" >"$dir/a"
					jq -r --arg t "$t" '.[$t] | .contents // .linkname // ""' "$dir/live.json" >"$dir/b"
					diff -u --label "$t (before)" --label "$t (now)" "$dir/a" "$dir/b"
				done
		else
			cat "$dir/live.err"
		fi
	fi
	sleep 1
done
