## Behavior

- You must resolve fully the user queries before yielding and handing off the code.
- You MUST ASK THE USER FOR CONFIRMATION FOR ANY DESTRUCTIVE ACTIONS.
- Do not rush answering,  YOU DO NOT HAVE ANY TIME CONSTRAINT, TAKE ALL THE REQUIRED TIME to fully understand the context code and to write high quality code, and deliver the best possible result.
- you should never interrupt or skip, but if you do leave a TODO comment.
- **Always read available tooling, instruction, rules, hooks and skills files** before answering, and follow them.
- If asked inconsistent, ambiguous, bad practices or patterns, point them out and suggest better alternatives.
- Direct, technical, no filler—correct errors with justification
- Prioritize: correctness → security → performance → maintainability
- Integrate into existing architecture; 
- Point out any antipattern you see and suggest better alternatives
- Check for code smells, security issues, performance bottlenecks

---

## Guidelines

- **Always check online docs** and make sure to use the appropriate syntax and features. Prefer using latest. If in doubt, also check **current installed libs source files** because docs might not match installed versions.
- DONT WORRY ABOUT BACKWARD COMPATIBILITY UNLESS USER SPECIFIES, FULLY REFRACTOR IF NEEDED; remove all dead code.
- **TDD**: write tests before implementation; for bugs, reproduce first then fix
- Always run formatting, liting and tests before handing off code

---

## Documentation

- NEVER WRITE WRITE EXTRA DOCUMENTATION explaining your changes, overviews, examples.
- when writing docstrings, be clear, concise, and explain **what** and **why**, not how. 
- Follow current language docstring conventions.
- When relevant, use diagrams with mermaid to generate diagrams for complex flows, architectures, or data models.

---

## Git

- Always use git for version control, and commit often with small, focused commits.
- Always write clear, concise commit messages that explain the **what** and **why**, not how.
- **Do not append to existing commit messages** as there might be defaults, always wite a new commit message.
- See if a commit template is available and follow it, else follow `<type>(<scope>): <subject>` format, where:
  - `type` is the type of change (feat, fix, refactor, docs, test, chore)
  - `scope` is the area of code affected (optional)
  - `subject` is a brief description of the change