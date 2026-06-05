# SuperCode — Roadmap

A CLI-based AI coding agent written in Go. It explores a repository, edits code using patches, runs terminal commands, and completes tasks autonomously through an agentic reasoning loop with tool calling.

This roadmap is intentionally minimal and built for an intermediate-level developer. No fancy features — just the core pieces needed for a working coding agent.

---

## Tech Stack

- **Language:** Go
- **LLM:** Any provider with a chat + tool-calling API (e.g. Anthropic or OpenAI)
- **CLI:** Standard library `flag`, or `cobra` if you want subcommands
- **HTTP:** Standard library `net/http`

---

## Phase 0 — Project Setup

Goal: Get a runnable Go CLI skeleton in place.

- Initialize the module: `go mod init supercode`
- Set up the folder structure:
  ```
  supercode/
    main.go
    internal/
      agent/      # the reasoning loop
      llm/        # API client
      tools/      # tool definitions and execution
      config/     # API keys, settings
  ```
- Read the API key from an environment variable (e.g. `SUPERCODE_API_KEY`).
- Print a hello message and exit. Confirm `go run .` works.

**Deliverable:** A CLI that builds and runs.

---

## Phase 1 — LLM Client

Goal: Talk to the model and get a response back.

- Build a small `llm` package with one function: send a list of messages, get a reply.
- Define basic message types: `role` (system / user / assistant / tool) and `content`.
- Handle the HTTP request, JSON encoding/decoding, and errors.
- Test it with a hardcoded prompt and print the model's reply.

**Deliverable:** `supercode "say hello"` returns a model response.

---

## Phase 2 — Tool System

Goal: Let the model request actions instead of just chatting.

- Define a `Tool` interface or struct: name, description, JSON schema for inputs, and an `Execute` function.
- Register tools in a map so the agent can look them up by name.
- Send the tool definitions to the model in each request.
- Parse tool-call requests from the model's response.

**Deliverable:** The model can ask to call a tool, and you can route that call to the right function.

---

## Phase 3 — Core Tools

Goal: Give the agent the minimum set of tools to be useful.

Build these four tools:

1. **`read_file`** — Read the contents of a file given a path.
2. **`list_files`** — List files/directories (this covers repository exploration).
3. **`run_command`** — Run a shell command and return its output (use Go's `os/exec`).
4. **`edit_file`** — Apply a patch-based edit to a file.

For `edit_file`, keep the patch format simple: have the model provide the file path, the exact text to find, and the replacement text. Read the file, do a string replace, and write it back. Fail clearly if the text isn't found.

**Deliverable:** Each tool works when called manually with test inputs.

---

## Phase 4 — The Agentic Loop

Goal: Tie everything together so the agent runs autonomously.

- Take a task from the user (e.g. `supercode "add a hello function to main.go"`).
- Loop:
  1. Send the conversation + tools to the model.
  2. If the model returns a tool call → execute it → add the result back to the conversation → repeat.
  3. If the model returns a final text answer → print it and stop.
- Add a max-iteration limit so it can't loop forever.
- Write a clear system prompt telling the model it's a coding agent and how to use its tools.

**Deliverable:** Give it a small task and watch it explore, edit, and finish on its own.

---

## Phase 5 — Polish

Goal: Make it pleasant and safe to use.

- Print each tool call and its result so the user can follow along.
- Ask for confirmation before running commands or editing files (optional but recommended).
- Handle errors gracefully — don't crash on a bad tool call; feed the error back to the model.
- Write a short `README.md` with setup and usage instructions.

**Deliverable:** A clean, understandable CLI you can demo.

---

## Stretch Goals (Optional)

Only after the core works:

- Conversation memory across multiple turns in one session.
- A `--dry-run` flag that shows planned edits without applying them.
- Simple syntax-highlighted diffs for file edits.
- Support for more than one LLM provider behind the same interface.

---

## Suggested Timeline

| Phase | Focus | Rough Effort |
|-------|-------|--------------|
| 0 | Setup | Half a day |
| 1 | LLM client | 1 day |
| 2 | Tool system | 1 day |
| 3 | Core tools | 1–2 days |
| 4 | Agentic loop | 1–2 days |
| 5 | Polish | 1 day |

Total: roughly **1 week** of focused work for an intermediate developer.

---

## Progress Checklist

### Foundation
- [ ] CLI interface
- [ ] Working directory support
- [ ] Chat history
- [ ] Tool framework

### Tools
- [ ] read_file
- [ ] search
- [ ] list_files
- [ ] write_file
- [ ] execute
- [ ] finish

### Agent
- [ ] OpenAI/Anthropic integration
- [ ] Structured JSON tool calls
- [ ] Agent execution loop
- [ ] Tool result feedback

### Upgrades
- [ ] apply_patch
- [ ] Automatic build verification
- [ ] Repository tree context

### Stretch
- [ ] Git integration
- [ ] Symbol search
- [ ] Multi-file editing

---

## Definition of Done

The project is complete when you can run a single command with a coding task, and the agent autonomously reads the repo, makes a patch-based edit, optionally runs a command to verify, and reports back — all from the terminal.
