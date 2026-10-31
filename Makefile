# Repo checks. Templates (*.tmpl) need chezmoi to render; the script templates
# (lint-tmpl) and the alias templates (lint-aliases) are rendered and linted.
ifeq ($(OS),Windows_NT)
# Git for Windows provides the POSIX shell required by the recipes, but does not add it to PATH.
# make cannot take a SHELL path with spaces (C:/Program Files): it finds sh.exe on PATH instead.
export PATH := $(shell git --exec-path)/../../../bin;$(PATH)
SHELL := sh.exe
# Git may retain CRLF in an existing working tree even when .gitattributes specifies LF.
SHELLCHECK_EXCLUDES := -e SC1017
endif

SHELLCHECK_EXCLUDES ?=
# The source state (.chezmoiroot); chezmoi --source . resolves it on its own.
SRC := home
SCRIPTS := $(shell git ls-files '*.sh')
SH_TMPL := $(shell git ls-files '$(SRC)/.chezmoiscripts/*.sh.tmpl')
# --override-data per render: OS × role × harness selection.
TMPL_DATA := '"light":false' '"light":true' '"work":true' '"harnesses":["opencode"]' '"harnesses":["claude","copilot","vscode","codex","opencode","pi"]' 
# rc files have no shebang and source runtime paths shellcheck cannot follow.
RC   := $(addprefix $(SRC)/,dot_bashrc dot_bash_profile dot_profile)
SH   := $(SCRIPTS) $(RC)
ZSH  := $(addprefix $(SRC)/,dot_zshrc dot_zprofile)
# settings.json is VS Code JSONC (comments), not strict JSON.
JSON := $(filter-out $(SRC)/dot_config/Code/User/settings.json,$(shell git ls-files '*.json' | xargs grep -L '{{'))

.PHONY: help lint lint-tmpl lint-aliases lint-plugins lint-edit aliases-audit test test-syntax check check-fast

help: ## List targets
	@grep -E '^[a-z-]+:.*##' $(MAKEFILE_LIST) | awk -F':.*## ' '{printf "%-13s %s\n", $$1, $$2}'

lint: lint-tmpl lint-aliases lint-plugins lint-edit ## shellcheck shell scripts, validate JSON
	shellcheck $(SHELLCHECK_EXCLUDES) $(SCRIPTS)
	shellcheck -s bash -e SC1090,SC1091 $(SHELLCHECK_EXCLUDES) $(RC)
	@for f in $(JSON); do jq empty "$$f" || { echo "invalid JSON: $$f"; exit 1; }; done
	@dir=$$(mktemp -d) && touch "$$dir/chezmoi.toml" && \
	for t in vscode-mcp.json claude-settings.json copilot-settings.json; do for d in '{}' '{"work":true}'; do \
		chezmoi execute-template --config "$$dir/chezmoi.toml" --source . --override-data "$$d" \
			< $(SRC)/.chezmoitemplates/$$t | jq empty || { echo "invalid JSON: $$t $$d"; exit 1; }; \
	done; done; rm -rf "$$dir"

# Isolated empty config, so the renders never read this machine's answers. A render for one
# OS leaves the other OS's package lists unset-but-unused (SC2034).
lint-tmpl: ## Render script templates with chezmoi, shellcheck the output
	@dir=$$(mktemp -d) && touch "$$dir/chezmoi.toml" && \
	for f in $(SH_TMPL); do for os in darwin linux; do for d in $(TMPL_DATA); do \
		chezmoi execute-template --config "$$dir/chezmoi.toml" --source . \
			--override-data "{\"chezmoi\":{\"os\":\"$$os\"},$$d}" < "$$f" > "$$dir/out.sh" && \
		shellcheck -e SC2034 "$$dir/out.sh" || { echo "lint-tmpl: $$f os=$$os $$d"; exit 1; }; \
	done; done; done; rm -rf "$$dir"

# Validate the registry, then render the alias files per OS × role: aliases.sh is
# shellchecked as bash, omz-aliases.zsh syntax-checked by zsh, aliases.ps1 parsed by
# pwsh and, on Windows, Windows PowerShell 5.1 (CI runners ship pwsh; skipped where it is
# missing). On Windows the temp dir is passed in C:/x form, which PowerShell can open.
lint-aliases: ## Validate the alias registry and lint the rendered alias files
	@sh scripts/aliases-lint.sh
	@dir=$$(mktemp -d) && { ! command -v cygpath >/dev/null 2>&1 || dir=$$(cygpath -m "$$dir"); } && \
	touch "$$dir/chezmoi.toml" && \
	render() { chezmoi execute-template --config "$$dir/chezmoi.toml" --source . --override-data "$$1" < "$$2"; } && \
	for os in darwin linux windows; do for d in $(TMPL_DATA); do \
		render "{\"chezmoi\":{\"os\":\"$$os\"},$$d}" $(SRC)/dot_config/shell/aliases.sh.tmpl > "$$dir/aliases.sh" && \
		shellcheck -s bash "$$dir/aliases.sh" || { echo "lint-aliases: aliases.sh os=$$os $$d"; exit 1; }; \
	done; done && \
	render '{}' $(SRC)/dot_config/shell/omz-aliases.zsh.tmpl > "$$dir/omz.zsh" && \
	{ ! command -v zsh >/dev/null 2>&1 || zsh -n "$$dir/omz.zsh"; } && \
	for d in $(TMPL_DATA); do \
		render "{\"chezmoi\":{\"os\":\"windows\"},$$d}" $(SRC)/dot_config/powershell/aliases.ps1.tmpl > "$$dir/aliases.ps1" && \
		for ps in pwsh powershell; do \
			if command -v $$ps >/dev/null 2>&1; then \
				$$ps -NoProfile -Command "\$$e = \$$null; [void][System.Management.Automation.Language.Parser]::ParseFile('$$dir/aliases.ps1', [ref]\$$null, [ref]\$$e); if (\$$e) { \$$e; exit 1 }" \
					|| { echo "lint-aliases: aliases.ps1 ($$ps) $$d"; exit 1; }; \
			fi; \
		done; \
	done; \
	command -v pwsh >/dev/null 2>&1 || echo "Skipping aliases.ps1 parse: pwsh is not installed."; \
	rm -rf "$$dir"

lint-plugins: ## Check the harness plugin list against giocaizzi/harness (HARNESS_DIR = local checkout)
	@sh scripts/harness-plugins-lint.sh

lint-edit: ## Validate the edit registry (.chezmoidata/edit.toml) and the dots-edit matrix
	@sh scripts/edit-lint.sh

aliases-audit: ## List live shell aliases the registry does not account for
	@sh scripts/aliases-audit.sh

test: test-syntax ## Syntax-check bash and zsh files, test the edit commands
	@pids=''; status=0; \
	for shell in bash pwsh; do sh scripts/dots-edit-test.sh "$$shell" & pids="$$pids $$!"; done; \
	if [ "$(OS)" = "Windows_NT" ]; then sh scripts/dots-edit-test.sh powershell & pids="$$pids $$!"; \
	else sh scripts/dots-edit-test.sh zsh & pids="$$pids $$!"; fi; \
	for pid in $$pids; do wait "$$pid" || status=1; done; exit $$status

test-syntax: ## Syntax-check bash and zsh files
	@for f in $(SH); do bash -n "$$f" || exit 1; done
	@if command -v zsh >/dev/null 2>&1; then \
		for f in $(ZSH); do zsh -n "$$f" || exit 1; done; \
	elif [ "$(OS)" = "Windows_NT" ]; then \
		echo "Skipping zsh syntax checks on Windows: zsh is not installed."; \
	else \
		echo "zsh is required to syntax-check zsh files."; exit 1; \
	fi

check: lint test ## lint + test

# The guard plugin's stop hook prefers check-fast: the edit command tests take minutes per shell.
check-fast: lint test-syntax ## lint + syntax checks, without the edit command tests
