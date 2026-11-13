# Repo checks. The registries, the config template and every rendered script, alias file
# and settings template are linted by the chezmoi-dots plugin (`chezmoi dots lint`),
# run against this checkout; it needs chezmoi to render.
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
DOTS := CHEZMOI_EXECUTABLE=chezmoi CHEZMOI_SOURCE_DIR="$(CURDIR)/$(SRC)" chezmoi-dots
AGENTS := agents
# rc files have no shebang and source runtime paths shellcheck cannot follow.
RC   := $(addprefix $(SRC)/,dot_bashrc dot_bash_profile dot_profile)
SH   := $(RC)
ZSH  := $(addprefix $(SRC)/,dot_zshrc dot_zprofile)
# The global git hooks deployed to ~/.config/git/hooks (POSIX sh).
HOOK_FILES := $(wildcard $(SRC)/dot_config/git/hooks/executable_*)
# settings.json is VS Code JSONC (comments), not strict JSON.
JSON := $(filter-out $(SRC)/dot_config/Code/User/settings.json,$(shell git ls-files '*.json' | xargs grep -L '{{'))

.PHONY: help lint lint-agents test test-syntax test-git-hooks check check-fast aliases-audit

help: ## List targets
	@grep -E '^[a-z-]+:.*##' $(MAKEFILE_LIST) | awk -F':.*## ' '{printf "%-13s %s\n", $$1, $$2}'

lint: ## shellcheck, validate JSON, lint registries, templates, and agents config
	shellcheck -s bash -e SC1090,SC1091 $(SHELLCHECK_EXCLUDES) $(RC)
	shellcheck -s sh $(HOOK_FILES)
	shellcheck -s bash tests/git-hooks.sh
	@for f in $(JSON); do jq empty "$$f" || { echo "invalid JSON: $$f"; exit 1; }; done
	@$(MAKE) --no-print-directory lint-agents
	$(DOTS) lint

# Render the agents config template and validate it with the agents CLI against the plain layers
# (empty stand-ins for the encrypted ones). Skipped when agents is not installed.
lint-agents:
	@if ! command -v $(AGENTS) >/dev/null 2>&1; then \
		echo "Skipping agents config check: $(AGENTS) is not installed."; exit 0; \
	fi; \
	t=$$(mktemp -d) && mkdir -p "$$t/agents/instructions" && \
	cp $(SRC)/dot_config/agents/instructions/core.md $(SRC)/dot_config/agents/instructions/env.md "$$t/agents/instructions/" && \
	touch "$$t/agents/instructions/personal.md" "$$t/agents/instructions/work.md" && \
	chezmoi --source "$(CURDIR)" execute-template < $(SRC)/.chezmoitemplates/agents-config.toml > "$$t/config.toml" && \
	XDG_CONFIG_HOME="$$t" APPDATA="$$t" $(AGENTS) check -file "$$t/config.toml" && \
	XDG_CONFIG_HOME="$$t" APPDATA="$$t" $(AGENTS) prompt check -file "$$t/config.toml" >/dev/null; \
	rc=$$?; rm -r "$$t"; exit $$rc

aliases-audit: ## List live shell aliases the registry does not account for
	$(DOTS) audit aliases

test: test-syntax test-git-hooks ## Syntax-check bash and zsh files, test the global git hooks

test-git-hooks: ## Test the global git hooks (stub gitleaks, throwaway repos)
	bash tests/git-hooks.sh

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

# The guard plugin's stop hook prefers check-fast: no plugin tests.
check-fast: lint test-syntax ## lint + syntax checks
