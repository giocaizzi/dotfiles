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
# Where `make tools` installs the pinned CLIs the checks run: a repo-local, gitignored dir by
# default, never ~/.local/bin (override TOOLS_DIR to install there; CI does). TOOLS_PLATFORM
# (os/arch: darwin|linux / amd64|arm64) defaults to this machine's.
TOOLS_DIR ?= $(CURDIR)/.tools/bin
TOOLS_PLATFORM ?= $(shell uname -s | tr '[:upper:]' '[:lower:]')/$(shell uname -m | sed -e 's/x86_64/amd64/' -e 's/aarch64/arm64/')
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
tools: ## Install the pinned chezmoi, chezmoi-dots and agents into TOOLS_DIR (default .tools/bin) for TOOLS_PLATFORM
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
		case "$$sum" in ""|*[!0-9a-f]*) echo "no valid sha256 for $$n on $(TOOLS_PLATFORM) in $$reg" >&2; exit 1;; esac; \
		[ $${#sum} -eq 64 ] || { echo "no valid sha256 for $$n on $(TOOLS_PLATFORM) in $$reg" >&2; exit 1; }; \
		url=$$(toml $$reg release.$$n url | sed -e "s/{version}/$$v/g" -e "s/{target}/$$t/g"); \
		curl -fsSL "$$url" -o "$(TOOLS_DIR)/$$n.part"; \
		got=$$(sha256sum "$(TOOLS_DIR)/$$n.part" 2>/dev/null || shasum -a 256 "$(TOOLS_DIR)/$$n.part"); got=$${got%% *}; \
		[ "$$got" = "$$sum" ] || { echo "sha256 mismatch for $$n: got $$got, want $$sum" >&2; exit 1; }; \
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

# Render the agents config template and validate it with the agents CLI, hermetically: explicit
# chezmoi data (no chezmoi config, a temp HOME/XDG) for both profiles with every harness selected.
# The layers come from the rendered config; each is the plain source layer, or an empty stand-in
# where the source is encrypted (*.age). Skipped when agents is not installed.
HARNESSES := claude copilot vscode codex opencode pi antigravity

lint-agents:
	@if ! command -v $(AGENTS) >/dev/null 2>&1; then \
		echo "Skipping agents config check: $(AGENTS) is not installed."; exit 0; \
	fi; \
	t=$$(mktemp -d) && mkdir -p "$$t/agents/instructions" "$$t/home" && \
	export HOME="$$t/home" XDG_CONFIG_HOME="$$t" APPDATA="$$t" && \
	hs=$$(echo $(HARNESSES) | tr ' ' ','); \
	rc=0; \
	for w in true false; do \
		p=$$([ $$w = true ] && echo work || echo personal); \
		data="{\"work\":$$w,\"light\":false,\"harnesses\":$$(echo $(HARNESSES) | jq -R -c 'split(" ")')}"; \
		chezmoi --config "$$t/chezmoi.toml" --source "$(CURDIR)" execute-template --override-data "$$data" \
			< $(SRC)/.chezmoitemplates/agents-config.toml > "$$t/config-$$p.toml" || { rc=1; break; }; \
		for l in $$($(AGENTS) config show -json -file "$$t/config-$$p.toml" | jq -r '.instructions.layers | keys[]'); do \
			src="$(SRC)/dot_config/agents/instructions/$$l.md"; \
			if [ -f "$$src" ]; then cp "$$src" "$$t/agents/instructions/"; else : > "$$t/agents/instructions/$$l.md"; fi; \
		done; \
		echo "agents check ($$p)"; \
		$(AGENTS) check -file "$$t/config-$$p.toml" -profile $$p -harnesses "$$hs" && \
		$(AGENTS) prompt check -file "$$t/config-$$p.toml" -profile $$p -harnesses "$$hs" >/dev/null || { rc=1; break; }; \
	done; \
	rm -r "$$t"; exit $$rc

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
