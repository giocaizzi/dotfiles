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
# Where `make tools` installs the pinned CLIs the checks run (CI puts it on PATH).
TOOLS_DIR ?= $(HOME)/.local/bin
TOOLS_PLATFORM ?= linux/amd64
AGENTS := agents
# rc files have no shebang and source runtime paths shellcheck cannot follow.
RC   := $(addprefix $(SRC)/,dot_bashrc dot_bash_profile dot_profile)
SH   := $(RC)
ZSH  := $(addprefix $(SRC)/,dot_zshrc dot_zprofile)
# The global git hooks deployed to ~/.config/git/hooks (POSIX sh).
HOOK_FILES := $(wildcard $(SRC)/dot_config/git/hooks/executable_*)
# settings.json is VS Code JSONC (comments), not strict JSON.
JSON := $(filter-out $(SRC)/dot_config/Code/User/settings.json,$(shell git ls-files '*.json' | xargs grep -L '{{'))

.PHONY: help tools lint lint-agents test test-syntax test-git-hooks check check-fast aliases-audit

help: ## List targets
	@grep -E '^[a-z-]+:.*##' $(MAKEFILE_LIST) | awk -F':.*## ' '{printf "%-13s %s\n", $$1, $$2}'

# CI tools at the versions the repo pins: agents and chezmoi-dots from [release.*] in the package
# registry (checksum-verified), chezmoi from .github/tools.toml (its installer verifies it).
tools: ## Install the pinned chezmoi, chezmoi-dots and agents into TOOLS_DIR (linux/amd64)
	@set -eu; \
	toml() { awk -v sec="$$2" -v key="$$3" '/^\[/ { s = $$0; gsub(/[][ ]/, "", s); next } \
		s == sec { k = $$1; gsub(/"/, "", k); if (k == key) { v = $$3; gsub(/"/, "", v); print v; exit } }' "$$1"; }; \
	reg=$(SRC)/.chezmoidata/packages.toml; \
	mkdir -p "$(TOOLS_DIR)"; \
	cv=$$(toml .github/tools.toml tools chezmoi); \
	curl -fsSL https://get.chezmoi.io | sh -s -- -b "$(TOOLS_DIR)" -t "v$$cv"; \
	for n in chezmoi-dots agents; do \
		v=$$(toml $$reg release.$$n version); \
		t=$$(toml $$reg release.$$n.target $(TOOLS_PLATFORM)); \
		sum=$$(toml $$reg release.$$n.sha256 $(TOOLS_PLATFORM)); \
		url=$$(toml $$reg release.$$n url | sed -e "s/{version}/$$v/g" -e "s/{target}/$$t/g"); \
		curl -fsSL "$$url" -o "$(TOOLS_DIR)/$$n.part"; \
		echo "$$sum  $(TOOLS_DIR)/$$n.part" | { sha256sum -c - 2>/dev/null || shasum -a 256 -c -; }; \
		chmod +x "$(TOOLS_DIR)/$$n.part"; mv "$(TOOLS_DIR)/$$n.part" "$(TOOLS_DIR)/$$n"; \
		echo "installed $$n $$v"; \
	done

lint: ## shellcheck, validate JSON, lint registries, templates, and agents config
	shellcheck -s bash -e SC1090,SC1091 $(SHELLCHECK_EXCLUDES) $(RC)
	shellcheck -s sh $(HOOK_FILES)
	shellcheck -s bash tests/git-hooks.sh
	@for f in $(JSON); do jq empty "$$f" || { echo "invalid JSON: $$f"; exit 1; }; done
	@$(MAKE) --no-print-directory lint-agents
	$(DOTS) lint

# Render the agents config template and validate it with the agents CLI. The layers come from the
# rendered config; each is the plain source layer, or an empty stand-in where the source is
# encrypted (*.age). Skipped when agents is not installed.
lint-agents:
	@if ! command -v $(AGENTS) >/dev/null 2>&1; then \
		echo "Skipping agents config check: $(AGENTS) is not installed."; exit 0; \
	fi; \
	t=$$(mktemp -d) && mkdir -p "$$t/agents/instructions" && \
	chezmoi --source "$(CURDIR)" execute-template < $(SRC)/.chezmoitemplates/agents-config.toml > "$$t/config.toml" && \
	for l in $$($(AGENTS) config show -json -file "$$t/config.toml" | jq -r '.instructions.layers | keys[]'); do \
		src="$(SRC)/dot_config/agents/instructions/$$l.md"; \
		if [ -f "$$src" ]; then cp "$$src" "$$t/agents/instructions/"; else : > "$$t/agents/instructions/$$l.md"; fi; \
	done && \
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
