# WIP: voice input across AI harnesses

Status: plan, not implemented; waits on the open questions at the end. Docs checked on 2026-11-06. Delete this file when the feature ships; its lasting content moves to `docs/OS.md` → Shared harness policy.

## Goal

One shared preference for voice input (speech to prompt text), `[harnessPolicy.voice]` in `.chezmoidata/harness-policy.toml`, mapped to each harness's native setting where one exists and documented as a gap, with a fallback, where none does. Same pattern as the bell (`ui.terminalBell` → `preferredNotifChannel`, `beep`, `[tui] notification_method`).

## What each harness supports

| Harness (version checked) | Dictation | Configurable from a file | Keys | Requirements and limits | Source |
|---|---|---|---|---|---|
| Claude Code (2.1.319) | Yes, `/voice` | Yes: `voice.enabled`, `voice.mode` (`hold` or `tap`), `voice.autoSubmit` in `settings.json`; key `voice:pushToTalk` in `keybindings.json` | Space: hold to record (default) or tap to start and send; Esc cancels | claude.ai sign-in only (no API key, Bedrock, Vertex, Foundry); audio streamed to Anthropic; no SSH or cloud sessions; language comes from the `language` setting, which **also sets Claude's reply language**, empty = English; 20 languages | [code.claude.com/docs/en/voice-dictation](https://code.claude.com/docs/en/voice-dictation) |
| Copilot CLI (1.0.105) | Yes, `/voice` | No documented key: enabled state and model are saved in `~/.copilot/settings.json` under undocumented names; `copilot help config` lists none | Hold Space; toggle Ctrl+X V | First `/voice` downloads a voice runtime and a model (multilingual default, or English or Spanish); local transcription; macOS Apple Silicon, Linux x64, Windows x64 and ARM64 | [docs.github.com … copilot-cli/voice-input](https://docs.github.com/en/copilot/how-tos/copilot-cli/use-copilot-cli/voice-input) |
| VS Code Copilot | Yes, built in | Yes: `dictation.enabled`, `dictation.model`, `agents.voice.language`, `agents.voice.enabled`, `accessibility.voice.speechTimeout` | Cmd/Ctrl+I in chat; hold the shortcut for push-to-talk | Local on desktop: Windows x64/ARM64, macOS Apple Silicon, Linux x64/ARM64 with glibc ≥ 2.34; `ms-vscode.vscode-speech` only for Intel Macs, ARM32, musl | [code.visualstudio.com/docs/configure/accessibility/voice](https://code.visualstudio.com/docs/configure/accessibility/voice) |
| Antigravity CLI (`agy` 1.5.2) | Yes, F5 or `/voice dictate` | No documented `settings.json` key (`/settings` → Voice Mode only chooses what plain `/voice` starts); rebinding through `/keybindings` (`voice.start_dictation`) | F5 start/stop, Enter keeps, Esc discards | One-time `/logout` and sign-in to grant the permission; not for business or enterprise accounts; docs cover macOS and Windows only; SSH via `agy mic-serve` + `ANTIGRAVITY_MIC` | [antigravity.google/docs/cli/commands/voice](https://antigravity.google/docs/cli/commands/voice/) |
| Codex CLI (0.166/0.167) | **No dictation.** `/voice` in the TUI starts a realtime voice conversation, not dictation | `[realtime] type = "transcription"` exists in source but is marked "Experimental / do not use"; `[audio] microphone` picks a device | — | The official Voice page covers only the ChatGPT desktop app (voice chat and dictation); the old `[features] voice_transcription` flag is gone | [learn.chatgpt.com/codex/features/voice](https://learn.chatgpt.com/codex/features/voice); `openai/codex` at `rust-v0.167.0`: `codex-rs/config/src/config_toml.rs` (`RealtimeToml`), `codex-rs/features/src/lib.rs` (`InAppDictation` is a desktop-app requirements gate) |
| opencode (2.2.5) | **No** | — | — | Keybinds docs list no voice action; only community plugins (whisper or cloud STT) | [opencode.ai/docs/keybinds](https://opencode.ai/docs/keybinds/) |
| pi (1.5.0) | **No** | — | — | Only community packages on pi.dev (`pi-voice-stt`, `pi-voice-input`, …) | pi.dev package index |

OS-level fallback for the harnesses without dictation (types into any focused text field, so into the terminal):

| OS | Fallback | Notes | Source |
|---|---|---|---|
| macOS | Dictation (Microphone key, or the shortcut in System Settings → Keyboard → Dictation, e.g. Fn twice) | Terminal needs no microphone grant of its own; Ghostty receives the text as input | [support.apple.com mh40584](https://support.apple.com/en-ie/guide/mac-help/mh40584/26/mac) |
| Windows | Voice typing, Win+H | Online (Azure Speech) except on Copilot+ PCs; works in Windows Terminal | [support.microsoft.com voice typing](https://support.microsoft.com/en-us/windows/use-voice-typing-to-talk-instead-of-type-on-your-pc-fec94565-c4bd-329d-e59a-af033fa5689f) |
| Linux | None built in | Gap: needs a third-party tool; documented, not installed | — |

## Proposed design

### Policy key

```toml
# --- D9: Voice input (dictation) ---
[harnessPolicy.voice]
enabled = true
mode = "hold"        # "hold" (push-to-talk) | "tap" (toggle)
autoSubmit = false   # send the prompt when recording stops
language = ""        # BCP 47 code; "" = each harness's default (English)
```

### Native mapping

| Harness | Adapter | Writes |
|---|---|---|
| Claude Code | `.chezmoitemplates/claude-settings.json` (merged by `dot_claude/modify_settings.json`) | `"voice": {"enabled", "mode", "autoSubmit"}`; `language` per open question 2 |
| VS Code | `dot_config/Code/User/settings.json` (static, checked by `chezmoi-agents check` against `translate.VSCodeSettings`) | `"dictation.enabled"`; `"agents.voice.language"` when `language` is set |
| Copilot CLI | none (no documented key) | Documented: run `/voice` once per machine (downloads runtime and model); hold Space matches `mode = "hold"` |
| Antigravity | none (no documented key) | Documented: `/logout` + sign in once, then F5; no business accounts |
| Codex CLI | none | Gap: OS dictation fallback; GitHub issue |
| opencode | none | Gap: OS dictation fallback; GitHub issue |
| pi | none | Gap: OS dictation fallback; GitHub issue |

A managed Claude `voice` key wins on every apply: `/voice off` in a session lasts until the next `chezmoi apply`. Change it in the policy instead.

### Code changes (`tools/agents`)

- `internal/policy/policy.go`: `Voice` struct (`enabled`, `mode`, `autoSubmit`, `language`), field `Voice` on `Policy`.
- `internal/policy/validate.go`: `mode` ∈ {`hold`, `tap`}; `language` empty or a BCP 47-shaped code.
- `internal/translate/claude.go`, `vscode.go`: emit the keys above. `codex.go`, `copilot.go`, `opencode.go`, `pi.go`, `antigravity.go` unchanged.
- `internal/parity/parity.go`: check Claude `voice.enabled` and VS Code `dictation.enabled` follow the policy.
- `internal/translate/testdata/harness-policy.golden.json` and tests: regenerate, add a voice case.

`chezmoi-agents` needs no release for this PR: `make lint` builds it from the tree. A later package bump follows the usual release flow.

### Docs

`docs/OS.md` → Shared harness policy: add "voice input" to the dimension list and a row cell per harness in the adapter table; a short "Voice input" subsection after the table with the per-harness one-time steps, the OS fallbacks above and the gaps, linking the issues. `AGENTS.md` unchanged (it links to the section).

### Per-OS differences

- macOS: Claude (native module, mic prompt for the terminal app; reset with `tccutil reset Microphone com.mitchellh.ghostty`), Copilot (Apple Silicon only), VS Code (Apple Silicon built in), Antigravity documented.
- Linux: Claude falls back to `arecord` or SoX when its native module fails; WSL needs WSLg and `sox libsox-fmt-pulse`. Copilot x64 only. Antigravity undocumented. No OS fallback. The Pi home server (light role) gets no harness configs, so nothing changes there.
- Windows: Claude, Copilot, VS Code, Antigravity native; Settings → Privacy & security → Microphone → desktop apps; Win+H fallback.

No package or install-script change: every native feature ships with its harness, and the SoX fallback is not added to `packages.toml` (Claude's native module covers macOS and Windows, and Linux desktops here are not a target yet).

## Shared files and the other drivers

- models-sync (`docs/wip/models-sync.md` in its worktree): touches `tools/agents/cli.go`, `internal/term`, a new models package, `.chezmoidata/models.json`, `docs/MODELS.md`, and a new "Models and subscriptions" section in `docs/OS.md`. It reads `harnessPolicy.models` only. Overlap with this plan: none in code; `docs/OS.md` different sections.
- usage tool: no plan file found in its worktree at the time of writing; this plan does not touch `cli.go` or add a subcommand.
- This plan's shared-file touch: one new table in `harness-policy.toml`, one new struct in `policy.go`, two translators, one parity block.

## Delivery

One PR on `docs/voice-input-plan` (renamed `feat/voice-input` at implementation), two workers in parallel:

1. Policy, adapters and Go: `harness-policy.toml`, `claude-settings.json`, `dot_config/Code/User/settings.json`, `tools/agents/internal/{policy,translate,parity}`. Checks: `make check`.
2. Docs: `docs/OS.md`, delete this file. Checks: `make lint`.

Before the PR: one `parity` + priority issue per gap in `giocaizzi/dotfiles` (Codex CLI, opencode, pi, Linux OS fallback), linked from the docs.

## Open questions

1. Scope: map only documented keys (Claude, VS Code), or also write the undocumented ones seen in binaries (Antigravity `voiceMode`; Copilot's names unknown)? Recommended: documented only.
2. Language: Claude's `language` also sets its reply language. (a) map `language` to VS Code only, Claude dictation stays English; (b) map it to Claude too, replies follow it; (c) no `language` key. Recommended: (a), with `language = ""`.
3. Mode: `hold` + `autoSubmit = false`, or `tap`? Recommended: `hold`, no auto-submit (the live Claude value, and the Copilot and VS Code push-to-talk default).
4. Work role: same on work machines, or `enabled = false` there? Recommended: same everywhere (harmless where unavailable).
5. Gap issues: one issue per gap at P3, or one umbrella issue? Recommended: one per gap, P3, label `parity`.
