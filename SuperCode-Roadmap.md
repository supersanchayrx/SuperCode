# SuperCode — Roadmap

A CLI-based AI coding agent written in Go. It explores a repository, edits code using patches, runs terminal commands, and completes tasks autonomously through an agentic reasoning loop with tool calling.

This roadmap is intentionally minimal and built for an intermediate-level developer. No fancy features — just the core pieces needed for a working coding agent.

---

## Tech Stack

- **Language:** Go
- **LLM:** OpenRouter chat completions API (currently `nvidia/nemotron-3.5-lightning:free`)
- **CLI:** `cobra` for commands and flags
- **HTTP:** Standard library `net/http`
- **Config:** `godotenv` for loading the API key from `.env`

---

## Current Status (15 Nov 2026)

### Done
- [x] **Phase 0 (project setup):** module, folder structure, API key loaded from `.env`.
- [x] **Phase 1 (LLM client):** the `llm` package sends a system prompt + user message to OpenRouter and returns the reply text.
- [x] **Cobra CLI:**
  - `rootCmd` in `cli/root.go`;
  - a `sayhello` command with a `--name` / `-n` flag;
  - a `chat` command with a `--mode` / `-M` flag.
- [x] **Agent package skeleton:**
  - `mode.go`: the `Mode` type (`auto`, `chat`, `plan`, `edit`) and `ParseMode`.
  - `router.go`: `classify(input)` asks the model whether a message is chat, plan or edit.
  - `agent.go`: the `Agent` struct and `Handle(input)`, which picks a mode for each message (using a local copy, so `auto` stays on) and calls the matching handler.
  - `chat.go`, `plan.go`, `edit.go`: handler methods `Chat`, `Plan`, `Edit`. `Plan` and `Edit` print "Not built yet".
- [x] **The REPL gets the agent:** `cli/repl.go` creates the `Agent` and passes it into the loop. The loop no longer imports `llm`.
- [x] **Chat prompt rewritten:** it's now `chatPrompt` (a `const`), written for terminal output, and names only `/mode plan` and `/mode edit`.
- [x] **`.env` is in `.gitignore`.**

### Loose ends in the current code
- **The project doesn't build right now.** Two causes:
  - `llm/llmChat.go` imports `cli`, which creates an import cycle (`cli → agent → llm → cli`). This is part of the half-done streaming work. Step 1 removes it; streaming comes back properly in Step 6.
  - `HandleRequest` now returns only `error`, but `agent/chat.go`, `agent/router.go` and `cli/sayhello.go` still expect `(error, string)`. Once the cycle is fixed, these will be the next errors.
- **The REPL calls `agent.Chat` instead of `agent.Handle`,** so the router and auto mode are skipped and every message goes to chat.
- **The mode `switch` in `chat()` (`cli/repl.go`) does nothing useful.** Both cases call `chatMode()`. The mode decision belongs in `agent.Handle`.
- **`/mode` is advertised but not implemented.** The welcome text tells users to use it, but a line starting with `/` is sent to the model like any other message.
- **`agent/chat.go` swallows errors.** On failure it prints nothing and returns `nil`, so the user sees an empty reply.
- **`log.Fatal` in `llm/llmChat.go` and `config/config.go`** ends the whole program on any network or API error (rate limits on the free model are common). The error fallbacks in `classify` and the REPL never get to run.
- **`main.go`** prints "Hello Reading api key --" before every command, including `--help`, and loads keys before Cobra runs.
- **The `.env` path is relative** (`internal/config/.env`), so `supercode` only works when run from the project folder.
- **`go.mod`** still marks godotenv as `// indirect`.
- **The old `internal/tools/` files show as deleted** in git (`tool.go`, `read_file.go`, `write_file.go`, `list_files.go`, `run_command.go`). They're needed again in Milestone B.
- **Most new files are untracked:** all of `agent/`, `cli/repl.go`, `cli/sayhello.go`, `llm/llmChat.go`. Nothing protects them until they're committed.

---

## Next Steps

Work top to bottom. Each step ends with a **Check** you can run, so you always know the project works before moving on. Don't start a step until the previous one's check passes.

| Order | Step | Why it comes here |
|-------|------|-------------------|
| 1 | Make it build, then commit | Nothing else can be tested until it compiles, and most files are untracked. |
| 2 | Restructure the CLI | The session shape everything else plugs into. |
| 3 | Tidy the agent | Small; makes chat mode answer properly through `Handle`. |
| 4 | Errors that don't kill the session | Needed before streaming, where a mid-answer failure would kill the REPL. |
| 5 | Housekeeping, then commit | Milestone A without streaming is done here. |
| 6 | Streaming chat replies | Nice-to-have on top of a working session. |
| 7 | Milestone B: tool calling | The big one. |
| 8 | Milestone C: plan and edit modes | Built on the tools from B. |

### Milestone A: a working session with modes (do this now)

**Done when:**
- `supercode` on its own opens a session.
- `supercode chat` does the same.
- `supercode --mode plan` starts in plan mode.
- `/mode` switches modes mid-session.
- Chat mode answers through the agent.
- An API error prints a message and the session keeps going.

#### Step 1. Make it build, then commit
Goal: the smallest changes that get `go build ./...` passing again. Streaming is put on hold until Step 6.
- [ ] **Remove the `cli` import from `llm/llmChat.go`** and the `cli.AddToOutputQueue()` call. `cli → agent → llm → cli` is an import cycle, and Go won't build it.
- [ ] **Set `Stream: false` for now** in `BuildChatRequest`, and make `HandleRequest` read the body and call `ParseResponse` again, so it returns the reply text.
- [ ] **Settle on one signature for `HandleRequest`** and update every caller (`agent/chat.go`, `agent/router.go`, `cli/sayhello.go`). The Go convention is `(string, error)`, and now is the cheapest time to switch, since you're touching every caller anyway. Do the same for `ParseResponse` and `config.LoadKeys`.
- [ ] **Leave `cli/output.go` alone, or delete it.** It's unused until Step 6, where it gets deleted anyway (see below).
- [ ] **Commit.** Most files are untracked, so nothing protects them until they're committed. Decide about `internal/tools` first: run `git checkout HEAD -- internal/tools` to bring it back (recommended, Milestone B needs it), or commit the deletion on purpose.

**Check:** `go build ./...` passes, and `supercode chat` answers a question.

#### Step 2. Restructure the CLI
- [ ] **`cli/root.go`** (partly done):
  - [x] Give `rootCmd` a `RunE` that starts the REPL, so plain `supercode` opens a session like `claude` does. (`RunE: chat` is in place.)
  - [ ] Move the `--mode` / `-M` flag here as a **persistent flag**, so it works on both `supercode` and `supercode chat`. Right now it's registered only on `chatCmd`, so `supercode --mode plan` fails with "unknown flag".
- [x] **Keep `chat` as an alias:** a subcommand whose `RunE` calls the same start-session function as root. (Both use `chat()`.) Cobra's `Aliases` field only gives a subcommand extra names, so `chat` stays its own small command.
- [ ] **`cli/repl.go`:** keep only the session loop (partly done).
  - [ ] Move `chatCmd` to its own file or `root.go`, and remove the mode `switch` from this file.
  - [x] The start function: `ParseMode(modeFlag)` → `agent.New(mode)` → run the loop with the agent.
  - [x] The loop takes the agent as a parameter, and on error prints `Error: …` and keeps looping.
  - [ ] The loop calls `a.Handle(input)` instead of `a.Chat(input)`.
  - [x] Remove the `llm` import from the REPL. The CLI should never call `llm` or `tools` directly. (`cli/sayhello.go` still imports `llm`; that's fine while it's a test command, or route it through the agent later.)
  - [ ] Show the current mode in the prompt, e.g. `[auto] >`, by reading `a.Mode` on every loop.
  - [ ] Replace plain `exit` / `quit` with `/exit`. Update the welcome line to mention `/help`.
- [ ] **New `cli/commands.go`:** `handleCommand(a, input) bool`, which returns `true` when the session should end. The loop calls it for any line starting with `/`, then `continue`s, so commands never reach the model.
  - `/mode <auto|chat|plan|edit>`: run `ParseMode` on the argument; on success set `a.Mode`, on error print it.
  - `/mode` on its own: print the current mode and the usage line.
  - `/chat`, `/plan`, `/edit`, `/auto`: shortcuts for `/mode <name>`. These match what the chat prompt already tells users.
  - `/help`: list all commands.
  - `/exit`, `/quit`: end the session.
  - Anything else: "Unknown command, type /help".

**Check:**
- `supercode`, `supercode chat` and `supercode --mode plan` all open a session, and the prompt shows the right mode.
- In auto mode, a question prints `(auto → chat)` and gets an answer.
- `/mode edit` changes the prompt to `[edit] >`, and a message prints "Not built yet".
- `/help`, `/exit` and an unknown `/foo` all behave as above.

#### Step 3. Tidy the agent
- [ ] **New `agent/prompts.go`:** move every system prompt here as a `const`.
  - `routerPrompt` from `router.go`.
  - `chatPrompt` from `chat.go` (already renamed and made a `const`).
  - Placeholder `planPrompt` and `editPrompt` for later.
- [x] **Make every prompt name only commands that really exist.** The chat prompt now says `/mode plan` and `/mode edit`, which exist after Step 2.
- [ ] **`agent/chat.go`:** if `HandleRequest` returns an error, return it; otherwise print the answer and return `nil`.

**Check:** with the API key temporarily wrong in `.env`, a chat message prints `Error: …` instead of an empty line. (Until Step 4, `log.Fatal` may still end the program here. That's the next step's job.)

#### Step 4. Errors that don't kill the session
- [ ] **`llm/llmChat.go`:** replace every `log.Fatal` / `log.Fatalf` with returning an error.
  - `SendRequest` returns `(*http.Request, error)`.
  - The non-200 status check returns `fmt.Errorf("API error (%d): %s", ...)`.
  - `ParseResponse` returns its `Unmarshal` error.
- [ ] **`config.LoadKeys`:** return errors instead of calling `log.Fatal`.

**Check:** with a wrong API key, or with Wi-Fi off, a message prints an error and the `>` prompt comes back.

#### Step 5. Housekeeping, then commit
- [ ] **`main.go`:** remove the greeting print. Move `config.LoadKeys()` into `rootCmd`'s `PersistentPreRunE`, so it runs before every command, can return an error, and isn't needed for `--help`.
- [ ] **Decide where the `.env` lives** so `supercode` works from any folder. Options:
  - a real environment variable,
  - `~/.supercode/.env` via `os.UserHomeDir()`,
  - a path next to the executable.
- [x] **Check `.env` is in `.gitignore`.**
- [ ] **Run `go mod tidy`.**
- [ ] **Commit.**

**Check:** `supercode --help` works without a `.env`, and `supercode` works when run from a different folder. Every "Done when" item above is true.

#### Step 6. Streaming chat replies
The reply should appear as the model writes it, not all at once at the end.

**Design: what to keep and what to drop**
- [ ] **Delete `outputQueue`, `AddToOutputQueue` and `streamText` from `cli/output.go`.** The stream already arrives a few tokens at a time, which is the typing effect. Printing each delta as it arrives is enough. A queue plus `time.Sleep` per character only adds delay on top of the network's.
- [ ] **Never import `cli` from `llm`** (removed in Step 1). It would be an import cycle, and it breaks the layering rule: `llm` must not know about the terminal.
- [ ] **Pass the output down instead of reaching up:**
  - `llm` takes either a callback (`onToken func(string)`) or an `io.Writer`, and calls it or writes to it for every delta.
  - `agent` passes it through.
  - `cli` supplies the real one: `os.Stdout`, or a function that prints.
  - This is the same "`io.Writer` output" item as in Milestone B, done early.

**`llm` package changes**
- [ ] **Add a stream chunk struct.** In a streaming chunk the text is in `choices[0].delta.content`, not `message.content`:
  - shape: `Choices []struct{ Delta struct{ Content string } }`;
  - add an optional `Error` field: a failure after the stream has started arrives inside a chunk, not as an HTTP status;
  - later, `delta.tool_calls` also arrives in pieces (Milestone B), so leave room for it.
  - Keep `ChatResponse` as it is for non-streaming calls.
- [ ] **Don't stream every request.**
  - `classify` in the router needs the whole one-word answer and must not print it to the user.
  - `ParseResponse` expects the non-streaming format.
  - Make `stream` a parameter, or have two functions (e.g. `Complete` and `Stream`). Only chat mode streams.
- [ ] **Read the body line by line, not with `io.ReadAll`.** `ReadAll` waits for the whole response, which defeats streaming.
  - Check the status code first. Only call `ReadAll` when the status isn't 200, to show the error body.
  - On 200, wrap `resp.Body` in a `bufio.Scanner`. For each line:
    - skip empty lines;
    - skip lines starting with `:` (OpenRouter keep-alive comments like `: OPENROUTER PROCESSING`);
    - strip the `data: ` prefix;
    - stop on `[DONE]`;
    - otherwise unmarshal into the chunk struct. Check `len(Choices) > 0` first (some chunks, such as a final usage-only one, have none), then send the delta to the callback or writer.
  - Print a final newline when the stream ends. That belongs to the caller, so `llm` stays out of the terminal.
- [ ] **Return the full text as well.** Collect the deltas in a `strings.Builder` and return `(string, error)`. History and tool-call parsing will need the complete reply later.
- [ ] **No `log.Fatal` in the stream code.** A dropped connection halfway through an answer would kill the whole REPL. Return the error (as in Step 4).

**Callers**
- [ ] **`agent/chat.go`:** call the streaming function with the output from the CLI. Don't print the returned string again, because the user has already seen it.

**Reference:** a working version of this loop exists in the last commit, as `ChatStream` in the old `client.go`. View it with `git show HEAD:internal/llm/client.go`. It has the scanner loop, the `data: ` handling, the `[DONE]` check and the `strings.Builder`. The main change is to replace its direct `fmt.Print` with the callback or writer.

**Check:** a long chat answer appears word by word, auto mode still prints `(auto → chat)` without the router's answer leaking to the screen, and `go build ./...` passes. Commit.

### Milestone B: tool calling (Phases 2–4) — step 7
- [ ] **Restore and review `internal/tools`:** the `Tool` struct, `Registry`, `read_file`, `list_files`, `write_file`, `run_command`.
- [ ] **Check that the chosen OpenRouter model supports tool calling.** Many free models don't. Switch models if needed.
- [ ] **Extend the `llm` types:**
  - send a `tools` list with each request,
  - parse `tool_calls` from the response,
  - support `role: "tool"` messages with a `tool_call_id`.
- [ ] **Change `BuildChatRequest`** to take a `[]Message` instead of two strings, so a conversation can grow inside one turn.
- [ ] **New `agent/loop.go`:** the agent loop, used by every mode handler.
  1. Send the messages and tools.
  2. If the reply has tool calls: run them, append the results, go back to step 1.
  3. If the reply is plain text: show it and stop.
  
  Add a maximum number of iterations.
- [ ] **Give each mode its own tool set:**
  - chat: read-only (`read_file`, `list_files`)
  - plan: read-only, plus writing `PLAN.md`
  - edit: read, write and `run_command`
- [ ] **Let the agent change its own mode** (e.g. a `switch_mode` tool, or a handler deciding a plan is needed first) by setting `a.Mode`. The REPL reads `a.Mode` for the prompt, so it stays in sync automatically.
- [ ] **Stop the agent printing directly.** Pass it an output destination (`io.Writer`) or callbacks from the CLI, so the CLI controls how tool progress ("reading main.go…") and answers look.
- [ ] **Ask the user before** writing files or running commands. The question goes through the CLI, not `fmt.Scanln` inside a tool.
- [ ] **Feed tool errors back to the model** as tool results instead of ending the turn.

### Milestone C: plan and edit modes — step 8
- [ ] **Plan mode:** call the model with the planner prompt (read-only tools, so it can look at the repo), write the result to `PLAN.md`, and ask before overwriting an existing one.
- [ ] **Edit mode:** an `edit_file` tool that does find-and-replace (see Phase 3), shows what will change, and asks before writing.
- [ ] **Conversation history across messages** on the `Agent` (a `[]Message`), plus a `/clear` command to reset it.

### Target structure after Milestone A
```
internal/
  cli/
    root.go       rootCmd: RunE starts a session, persistent --mode flag, PersistentPreRunE loads keys
    repl.go       session loop: prompt, read line, command vs message, call a.Handle, print errors
    commands.go   slash commands: /mode, /chat /plan /edit /auto, /help, /exit
    chat.go       `chat` subcommand: alias that starts the same session (optional, can live in root.go)
    sayhello.go
  agent/
    agent.go      Agent struct + Handle (dispatch per message)
    mode.go       Mode type + ParseMode
    router.go     classify (auto mode)
    prompts.go    every system prompt
    chat.go       chat mode handler
    plan.go       plan mode handler (stub)
    edit.go       edit mode handler (stub)
  llm/
    llmChat.go    HTTP client, request/response types
  config/
    config.go     model, endpoint, LoadKeys
  tools/          restored in Milestone B
```

**Rule of thumb:** `cli` reads input and shows output, `agent` decides what to do, `llm` does HTTP, `tools` touch the filesystem. If a `cli` file imports `llm` or `tools`, logic has leaked into the wrong layer.

---

## Phase 0 — Project Setup ✅

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

## Phase 1 — LLM Client ✅

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
- [x] CLI interface
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
- [x] LLM API integration (OpenRouter, OpenAI-compatible)
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
