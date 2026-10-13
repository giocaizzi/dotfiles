# Repo checks. Templates (*.tmpl) need chezmoi to render; only the script templates are
# rendered and linted (lint-tmpl).
ifeq ($(OS),Windows_NT)
# Git for Windows provides the POSIX shell required by the recipes, but does not add it to PATH.
SHELL := $(shell git --exec-path)/../../../bin/sh.exe
# Git may retain CRLF in an existing working tree even when .gitattributes specifies LF.
SHELLCHECK_EXCLUDES := -e SC1017
endif

SHELLCHECK_EXCLUDES ?=
SCRIPTS := $(shell git ls-files '*.sh')
SH_TMPL := $(shell git ls-files '.chezmoiscripts/*.sh.tmpl')
# --override-data per render: OS × role × harness selection.
TMPL_DATA := '"light":false' '"light":true' '"work":true' '"harnesses":["opencode"]' 
# rc files have no shebang and source runtime paths shellcheck cannot follow.
RC   := dot_bashrc dot_bash_profile dot_profile
SH   := $(SCRIPTS) $(RC)
ZSH  := dot_zshrc dot_zprofile
# settings.json is VS Code JSONC (comments), not strict JSON.
JSON := $(filter-out dot_config/Code/User/settings.json,$(shell git ls-files '*.json' | xargs grep -L '{{'))

.PHONY: help lint lint-tmpl test check

help: ## List targets
	@grep -E '^[a-z-]+:.*##' $(MAKEFILE_LIST) | awk -F':.*## ' '{printf "%-9s %s\n", $$1, $$2}'

lint: lint-tmpl ## shellcheck shell scripts, validate JSON
	shellcheck $(SHELLCHECK_EXCLUDES) $(SCRIPTS)
	shellcheck -s bash -e SC1090,SC1091 $(SHELLCHECK_EXCLUDES) $(RC)
	@for f in $(JSON); do jq empty "$$f" || { echo "invalid JSON: $$f"; exit 1; }; done

# Isolated empty config, so the renders never read this machine's answers. A render for one
# OS leaves the other OS's package lists unset-but-unused (SC2034).
lint-tmpl: ## Render script templates with chezmoi, shellcheck the output
	@dir=$$(mktemp -d) && touch "$$dir/chezmoi.toml" && \
	for f in $(SH_TMPL); do for os in darwin linux; do for d in $(TMPL_DATA); do \
		chezmoi execute-template --config "$$dir/chezmoi.toml" --source . \
			--override-data "{\"chezmoi\":{\"os\":\"$$os\"},$$d}" < "$$f" > "$$dir/out.sh" && \
		shellcheck -e SC2034 "$$dir/out.sh" || { echo "lint-tmpl: $$f os=$$os $$d"; exit 1; }; \
	done; done; done; rm -rf "$$dir"

test: ## Syntax-check bash and zsh files
	@for f in $(SH); do bash -n "$$f" || exit 1; done
	@if command -v zsh >/dev/null 2>&1; then \
		for f in $(ZSH); do zsh -n "$$f" || exit 1; done; \
	elif [ "$(OS)" = "Windows_NT" ]; then \
		echo "Skipping zsh syntax checks on Windows: zsh is not installed."; \
	else \
		echo "zsh is required to syntax-check zsh files."; exit 1; \
	fi

check: lint test ## lint + test
