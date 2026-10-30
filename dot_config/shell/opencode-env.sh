# shellcheck shell=bash
# opencode LSP: the tool is experimental and off by default; servers come from the
# install list, so opencode must not download its own.
export OPENCODE_EXPERIMENTAL_LSP_TOOL=true
export OPENCODE_DISABLE_LSP_DOWNLOAD=true
