# Repo checks. Templates (*.tmpl) need chezmoi to render, so they are skipped.
SCRIPTS := $(shell git ls-files '*.sh')
# rc files have no shebang and source runtime paths shellcheck cannot follow.
RC   := dot_bashrc dot_bash_profile dot_profile
SH   := $(SCRIPTS) $(RC)
ZSH  := dot_zshrc dot_zprofile
# settings.json is VS Code JSONC (comments), not strict JSON.
JSON := $(filter-out dot_config/Code/User/settings.json,$(shell git ls-files '*.json' | xargs grep -L '{{'))

.PHONY: help lint test check

help: ## List targets
	@grep -E '^[a-z]+:.*##' $(MAKEFILE_LIST) | awk -F':.*## ' '{printf "%-6s %s\n", $$1, $$2}'

lint: ## shellcheck shell scripts, validate JSON
	shellcheck $(SCRIPTS)
	shellcheck -s bash -e SC1090,SC1091 $(RC)
	@for f in $(JSON); do jq empty "$$f" || { echo "invalid JSON: $$f"; exit 1; }; done

test: ## Syntax-check bash and zsh files
	@for f in $(SH); do bash -n "$$f" || exit 1; done
	@for f in $(ZSH); do zsh -n "$$f" || exit 1; done

check: lint test ## lint + test
