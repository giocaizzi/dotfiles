#!/usr/bin/env bash
# Claude Code status line - mirrors ~/.mytheme.omp.json style
# Receives JSON on stdin from Claude Code

input=$(cat)

# --- colours (Catppuccin Mocha) ---
BLUE='\033[38;2;137;180;250m'   # #89b4fa  blue
PINK='\033[38;2;203;166;247m'   # #cba6f7  mauve
WHITE='\033[38;2;205;214;244m'  # #cdd6f4  text
RED='\033[38;2;243;139;168m'    # #f38ba8  red   (git)
YELLOW='\033[38;2;249;226;175m' # #f9e2af  yellow (gcp / warnings)
GREY='\033[38;2;108;112;134m'   # #6c7086  overlay0 (dim)
RESET='\033[0m'

# --- extract Claude JSON fields ---
cwd=$(echo "$input"         | jq -r '.workspace.current_dir // .cwd // empty')
model=$(echo "$input"       | jq -r '.model.display_name // empty')
used=$(echo "$input"        | jq -r '.context_window.used_percentage // empty')
session=$(echo "$input"     | jq -r '.session_name // empty')
rl_5h=$(echo "$input"       | jq -r '.rate_limits.five_hour.used_percentage // empty')
rl_5h_reset=$(echo "$input" | jq -r '.rate_limits.five_hour.resets_at // empty')
rl_7d=$(echo "$input"       | jq -r '.rate_limits.seven_day.used_percentage // empty')
rl_7d_reset=$(echo "$input" | jq -r '.rate_limits.seven_day.resets_at // empty')

# --- colour for a percentage (blue <50, yellow 50-79, red ≥80) ---
pct_color() {
    local pct=$1
    if [ "$pct" -ge 80 ] 2>/dev/null; then
        printf '\033[38;2;243;139;168m'   # red
    elif [ "$pct" -ge 50 ] 2>/dev/null; then
        printf '\033[38;2;249;226;175m'   # yellow
    else
        printf '\033[38;2;137;180;250m'   # blue
    fi
}

# --- format epoch seconds as compact "until reset" string (e.g. 2h15m, 3d4h, 12m) ---
fmt_remaining() {
    local resets_at=$1
    local now diff days hours minutes
    now=$(date +%s)
    diff=$((resets_at - now))
    if [ "$diff" -le 0 ]; then
        printf 'now'
        return
    fi
    days=$((diff / 86400))
    hours=$(( (diff % 86400) / 3600 ))
    minutes=$(( (diff % 3600) / 60 ))
    if [ "$days" -gt 0 ]; then
        printf '%dd%dh' "$days" "$hours"
    elif [ "$hours" -gt 0 ]; then
        printf '%dh%dm' "$hours" "$minutes"
    else
        printf '%dm' "$minutes"
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
    if [ -n "$rl_5h_reset" ]; then
        rl_part="${rl_part}$(printf "${GREY}/%s${RESET}" "$(fmt_remaining "$rl_5h_reset")")"
    fi
fi
if [ -n "$rl_7d" ]; then
    rl_7d_int=${rl_7d%.*}
    rl_part="${rl_part}$(printf " $(pct_color "$rl_7d_int")7d:%s%%${RESET}" "$rl_7d_int")"
    if [ -n "$rl_7d_reset" ]; then
        rl_part="${rl_part}$(printf "${GREY}/%s${RESET}" "$(fmt_remaining "$rl_7d_reset")")"
    fi
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
