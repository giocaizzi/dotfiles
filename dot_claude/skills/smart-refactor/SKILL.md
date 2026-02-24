---
agent: 'agent'
description: 'Smartly refactor files and strings of code. Use when you need to do complex refactoring across multiple files or large codebases.'
tools: ['vscode', 'execute', 'read', 'edit', 'search', 'web', 'agent', 'ms-python.python/getPythonEnvironmentInfo', 'ms-python.python/getPythonExecutableCommand', 'ms-python.python/installPythonPackage', 'ms-python.python/configurePythonEnvironment', 'todo']
---

You have to refactor a huge codebase, with changes across filenames and code.

In order to do this effectively,

- **Do not try to intervene manually in every file**. Instead, devise a systematic plan to apply the changes across the entire codebase.

## Tools
- use `mv` to move files, rename files and directories,
- use `find` to locate files and directories matching specific patterns,
- use `grep` or `rg` to search for patterns across files,
- use `sed` for bulk string replacements across multiple files,
- Use `pear` or `python` for complex refactoring string manipulations that are hard to do with `sed`.