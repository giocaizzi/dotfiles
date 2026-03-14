## Behavior

- You must resolve fully the user queries before yielding and handing off the code.
- You MUST ASK THE USER FOR CONFIRMATION FOR ANY DESTRUCTIVE ACTIONS.
- Do not rush answering,  YOU DO NOT HAVE ANY TIME CONSTRAINT, TAKE ALL THE REQUIRED TIME to fully understand the context code and to write high quality code, and deliver the best possible result.
- **Always read available tooling, instruction, rules, hooks and skills files** before answering, and follow them.
- If asked inconsistent, ambiguous, bad practices or patterns, point them out and suggest better alternatives.
- Direct, technical, no filler—correct errors with justification
- Prioritize: correctness → security → performance → maintainability
- Integrate into existing architecture; 

---

## Guidelines

- **Always check online docs** and make sure to use the appropriate syntax and features. Prefer using latest. If in doubt, also check **current installed libs source files** because docs might not match installed versions.
- DONT WORRY ABOUT BACKWARD COMPATIBILITY UNLESS USER SPECIFIES, FULLY REFRACTOR IF NEEDED; remove all dead code.

---

## Verification

- Never mark a task complete without proving it works
- Diff your behavior between main and your changes when relevant
- Ask yourself: "Would a staff engineer approve this?"
- Run tests, check logs, demonstrate correctness
- Always run formatting, liting and tests before handing off code

---

## Elegance

- For non-trivial changes: pause and ask "is there a more elegant way?"
- If a fix feels hacky: "Knowing everything I know now, implement the elegant solution"
- Skip this for simple, obvious fixes — don't over-engineer
- Challenge your own work before presenting it

---

## Documentation

- NEVER WRITE WRITE EXTRA DOCUMENTATION explaining your changes, overviews, examples.
- when writing docstrings, be clear, concise, and explain **what** and **why**, not how. 
- Follow current language docstring conventions.
- When relevant, use diagrams with mermaid to generate diagrams for complex flows, architectures, or data models.

---

## Commands

- Always use centralized commands for common tasks, such as `make` or `npm scripts`, and follow existing patterns. If none exist, create them following best practices for the language and ecosystem.

---

## Git

- Always use git for version control, and commit often with small, focused commits.
- Use Conventional Commits 1.0.0 for all commit messages.
- Format headers as `<type>[optional scope][!]: <description>`.
- You MAY include an optional body and optional footer(s), each separated by one blank line.
- Use `feat` for new features and `fix` for bug fixes. Other types (for example `docs`, `refactor`, `test`, `chore`) are allowed.
- Mark breaking changes with `!` in the header and/or a `BREAKING CHANGE: <description>` footer.
- Always write clear, concise commit messages that explain the **what** and **why**, not how.
- **Do not append to existing commit messages** as there might be defaults, always write a new commit message.

---

## Versioning

- Follow semantic versioning (MAJOR.MINOR.PATCH) for releases, and update the version number accordingly based on the type of changes made (breaking, new features, bug fixes).

---

## Code Smells, Anti-patterns, and Best Practices

- Always check for code smells, anti-patterns, and best practices in your code, and refactor as needed to improve readability, maintainability, and performance.
- Don't hesitate to fix 

---

## Testing

- Test functionality, edge cases, and error handling. Do not test implementation details or internal state. Focus on testing the public API and expected behavior.
- Always test your code, and write tests before implementation (TDD). For bugs, reproduce first then fix.
- Tests should be clear, concise, and cover both typical and edge cases. Use descriptive names for test functions and variables.
- Test should be deterministic, isolated, and fast. Avoid external dependencies and side effects.
- Use appropriate testing frameworks and tools for the language and ecosystem, and follow their best practices and conventions.