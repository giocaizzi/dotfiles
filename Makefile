# Repo checks. The registries, the config template and every rendered script, alias file
# and settings template are linted by the chezmoi-dots plugin (tools/dots, `chezmoi dots
# lint`), built here and run against this checkout; it needs chezmoi to render.
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
DOTS := CHEZMOI_EXECUTABLE=chezmoi CHEZMOI_SOURCE_DIR="$(CURDIR)/$(SRC)" tools/dots/bin/chezmoi-dots
# rc files have no shebang and source runtime paths shellcheck cannot follow.
RC   := $(addprefix $(SRC)/,dot_bashrc dot_bash_profile dot_profile)
SH   := $(RC)
ZSH  := $(addprefix $(SRC)/,dot_zshrc dot_zprofile)
# settings.json is VS Code JSONC (comments), not strict JSON.
JSON := $(filter-out $(SRC)/dot_config/Code/User/settings.json,$(shell git ls-files '*.json' | xargs grep -L '{{'))

.PHONY: help dots agents lint test test-syntax check check-fast aliases-audit

help: ## List targets
	@grep -E '^[a-z-]+:.*##' $(MAKEFILE_LIST) | awk -F':.*## ' '{printf "%-13s %s\n", $$1, $$2}'

dots: ## Build the chezmoi-dots plugin into tools/dots/bin/
	@$(MAKE) --no-print-directory -C tools/dots build

agents: ## Build chezmoi-agents into tools/agents/bin/
	@$(MAKE) --no-print-directory -C tools/agents build

lint: dots agents ## shellcheck, validate JSON, lint registries, templates, policy, and plugins
	shellcheck -s bash -e SC1090,SC1091 $(SHELLCHECK_EXCLUDES) $(RC)
	@for f in $(JSON); do jq empty "$$f" || { echo "invalid JSON: $$f"; exit 1; }; done
	$(DOTS) lint
	tools/agents/bin/chezmoi-agents check
	$(MAKE) -C tools/dots lint
	$(MAKE) -C tools/agents lint

aliases-audit: dots ## List live shell aliases the registry does not account for
	$(DOTS) audit aliases

test: test-syntax ## Syntax-check bash and zsh files, test the plugins
	$(MAKE) -C tools/dots test
	$(MAKE) -C tools/agents test

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

# The guard plugin's stop hook prefers check-fast: no race detector, no plugin tests.
check-fast: lint test-syntax ## lint + syntax checks, without the plugin tests
