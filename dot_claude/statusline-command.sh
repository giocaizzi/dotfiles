#!/usr/bin/env bash
# Claude Code status line - mirrors ~/.mytheme.omp.json style
# Receives JSON on stdin from Claude Code

input=$(cat)

# --- colours (ANSI; terminal renders these dimmed) ---
BLUE='\033[38;2;126;184;218m'   # #7eb8da
PINK='\033[38;2;255;165;216m'   # #ffa5d8
WHITE='\033[38;2;255;255;255m'  # #ffffff
RED='\033[38;2;240;80;50m'      # #F05032  (git)
YELLOW='\033[38;2;244;180;0m'   # #F4B400  (gcp)
RESET='\033[0m'

# --- extract Claude JSON fields ---
cwd=$(echo "$input"         | jq -r '.workspace.current_dir // .cwd // empty')
model=$(echo "$input"       | jq -r '.model.display_name // empty')
used=$(echo "$input"        | jq -r '.context_window.used_percentage // empty')
session=$(echo "$input"     | jq -r '.session_name // empty')
rl_5h=$(echo "$input"       | jq -r '.rate_limits.five_hour.used_percentage // empty')
rl_7d=$(echo "$input"       | jq -r '.rate_limits.seven_day.used_percentage // empty')

# --- colour for a percentage (blue <50, yellow 50-79, red ≥80) ---
pct_color() {
    local pct=$1
    if [ "$pct" -ge 80 ] 2>/dev/null; then
        printf '\033[38;2;240;80;50m'
    elif [ "$pct" -ge 50 ] 2>/dev/null; then
        printf '\033[38;2;244;180;0m'
    else
        printf '\033[38;2;126;184;218m'
    fi
}

# --- user / host ---
user=$(whoami)
host=$(hostname -s)

# --- path (replace $HOME with ~) ---
if [ -n "$cwd" ]; then
    display_path="${cwd/#$HOME/\~}"
else
    display_path="~"
fi

# --- git branch + dirty indicator ---
git_part=""
if [ -n "$cwd" ] && git -C "$cwd" rev-parse --is-inside-work-tree >/dev/null 2>&1; then
    branch=$(git -C "$cwd" symbolic-ref --short HEAD 2>/dev/null \
             || git -C "$cwd" rev-parse --short HEAD 2>/dev/null)
    dirty=""
    if ! git -C "$cwd" diff --quiet 2>/dev/null || ! git -C "$cwd" diff --cached --quiet 2>/dev/null; then
        dirty=" *"
    fi
    git_part=$(printf "${RED}[${WHITE}${RESET}${RED} %s%s]${RESET}" "$branch" "$dirty")
fi

# --- GCP project ---
gcp_part=""
gcp_project=$(gcloud config get-value project 2>/dev/null)
if [ -n "$gcp_project" ]; then
    gcp_part=$(printf " ${YELLOW}[%s]${RESET}" "$gcp_project")
fi

# --- session name ---
session_part=""
if [ -n "$session" ]; then
    session_part=$(printf " ${BLUE}(%s)${RESET}" "$session")
fi

# --- model ---
model_part=""
if [ -n "$model" ]; then
    model_part=$(printf "${WHITE}%s${RESET}" "$model")
fi

# --- context usage ---
ctx_part=""
if [ -n "$used" ]; then
    used_int=${used%.*}
    ctx_part=$(printf " $(pct_color "$used_int")ctx:%s%%${RESET}" "$used_int")
fi

# --- rate limits (Pro/Max only; field absent otherwise) ---
rl_part=""
if [ -n "$rl_5h" ]; then
    rl_5h_int=${rl_5h%.*}
    rl_part="${rl_part}$(printf " $(pct_color "$rl_5h_int")5h:%s%%${RESET}" "$rl_5h_int")"
fi
if [ -n "$rl_7d" ]; then
    rl_7d_int=${rl_7d%.*}
    rl_part="${rl_part}$(printf " $(pct_color "$rl_7d_int")7d:%s%%${RESET}" "$rl_7d_int")"
fi

# --- assemble ---
# Line 1: ┌[user from host][git][gcp][session]
# Line 2: └| [path]  model  ctx%

line1=$(printf "${BLUE}┌${RESET}[${WHITE}%s${RESET} from ${WHITE}%s${RESET}]" "$user" "$host")
[ -n "$git_part" ]     && line1="${line1} ${git_part}"
[ -n "$gcp_part" ]     && line1="${line1}${gcp_part}"
[ -n "$session_part" ] && line1="${line1}${session_part}"

line2=$(printf "${BLUE}└|${RESET} ${PINK}[%s]${RESET}" "$display_path")
[ -n "$model_part" ] && line2="${line2}  ${model_part}"
[ -n "$ctx_part" ]   && line2="${line2}${ctx_part}"
[ -n "$rl_part" ]    && line2="${line2}${rl_part}"

printf "%b\n%b\n" "$line1" "$line2"
