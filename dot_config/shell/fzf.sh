# ============================================================================
# FZF CONFIGURATION (shared by bash + zsh)
# ============================================================================
# Sourced by ~/.profile. Key bindings are loaded by .bashrc / .zshrc since
# they require shell-specific scripts.
# ============================================================================

export FZF_DEFAULT_OPTS='
  --height=40%
  --layout=reverse
  --info=inline
  --border
  --margin=1
  --padding=1
  --preview-window=right:60%:wrap
  --bind ctrl-u:preview-half-page-up,ctrl-d:preview-half-page-down
  --color fg:252,bg:233,hl:67,fg+:252,bg+:235,hl+:81
  --color info:144,prompt:161,spinner:135,pointer:135,marker:118
'

export FZF_CTRL_R_OPTS="
  --preview 'echo {2..}' --preview-window up:3:hidden:wrap
  --bind 'ctrl-/:toggle-preview'
  --bind 'ctrl-y:execute-silent(echo -n {2..} | pbcopy)+abort'
  --header 'Press CTRL-Y to copy command to clipboard'
"

# Use fd for file search when available
if command -v fd >/dev/null 2>&1; then
	export FZF_DEFAULT_COMMAND='fd --type f --hidden --follow --exclude .git'
	export FZF_CTRL_T_COMMAND="$FZF_DEFAULT_COMMAND"
fi

# History search function (used by both bash and zsh)
h() {
	local selected
	selected=$(fc -l 1 | fzf --tac --no-sort +m --query "$*" --prompt="History> " | sed -E 's/^[[:space:]]*[0-9]+[*]?[[:space:]]+//')
	if [ -n "$selected" ]; then
		print -z "$selected" 2>/dev/null || printf '%s\n' "$selected"
	fi
}
alias hist='h'
