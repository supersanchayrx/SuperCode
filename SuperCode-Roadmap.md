# SuperCode — Roadmap

A CLI-based AI coding agent written in Go. It explores a repository, edits code using patches, runs terminal commands, and completes tasks autonomously through an agentic reasoning loop with tool calling.

This roadmap is intentionally minimal and built for an intermediate-level developer. No fancy features — just the core pieces needed for a working coding agent.

---

## Tech Stack

- **Language:** Go
- **LLM:** OpenRouter chat completions API: a model per task (`ORCHESTRATOR_MODEL`, `PLAN_MODEL`, `EDIT_MODEL` in `.env`), several API keys (`SUPERCODE_API_KEY1`…`N`), `openrouter/free` as the last fallback
- **CLI:** `cobra` for commands and flags
- **HTTP:** Standard library `net/http`
- **Config:** `godotenv` for loading the API key from `.env`

---

## Current Status (15 Nov 2026)

### Done
- [x] **Phase 0 (project setup):** module, folder structure, API key loaded from `.env`.
- [x] **Phase 1 (LLM client):** the `llm` package sends a system prompt + user message to OpenRouter and returns the reply text.
- [x] **The project builds again** (`go build ./...` passes).
- [x] **Cobra CLI:**
  - `rootCmd` in `cli/root.go`, with `RunE: startSession`, so plain `supercode` opens a session;
  - a persistent `--mode` / `-M` flag on root, so it works on `supercode` and `supercode chat`;
  - `chatCmd` in `cli/chatCli.go`, an alias that also runs `startSession`;
  - a `sayhello` command with a `--name` / `-n` flag.
- [x] **Session start and loop are split:**
  - `startSession` (`root.go`): create the agent from `modeFlag` and run the loop. The useless mode `switch` is gone.
  - `runRepl` (`repl.go`): the loop. Lines starting with `/` go to `handleCommand`; everything else goes to `a.Handle(input)`. It shows the mode in the prompt (`[auto]>`) and prints errors without exiting. It no longer imports `llm` or `cobra`.
- [x] **Slash commands started:** `cli/commands.go` has `handleCommand` with `/help`, `/exit`, `/quit`, `/mode` and an "Unknown command" fallback. (Two bugs left, see Loose ends.)
- [x] **Mode logic moved into `agent`:** `SetMode(name string) error` in `mode.go`; `New(name string) (*Agent, error)` parses the mode itself, so `cli` no longer calls `ParseMode`; `ParseMode` lowercases its input, so `PLAN` works too.
- [x] **Agent package skeleton:**
  - `mode.go`: the `Mode` type (`auto`, `chat`, `plan`, `edit`), `ParseMode` and `SetMode`.
  - `router.go`: `classify(input)` asks the model whether a message is chat, plan or edit.
  - `agent.go`: the `Agent` struct and `Handle(input)`, which picks a mode for each message (using a local copy, so `auto` stays on) and calls the matching handler.
  - `chat.go`, `plan.go`, `edit.go`: handler methods `Chat`, `Plan`, `Edit`. `Plan` and `Edit` print "Not built yet".
- [x] **`llm` cleaned up for now:** no `cli` import (the import cycle is gone), `Stream: false`, and `HandleRequest` / `ParseResponse` return `(string, error)`. All callers are updated.
- [x] **Prompts live in `agent/prompts.go`** (`routerPrompt`, `chatPrompt`).
- [x] **Errors don't kill the session:** no `log.Fatal` left. `llm` and `config` return errors, `Chat` passes them up, and `runRepl` prints `Error: …` and keeps going. A missing `.env` exits cleanly with a message from `main.go`.
- [x] **Chat prompt rewritten:** it's now `chatPrompt` (a `const`), written for terminal output, and names only `/mode plan` and `/mode edit`.
- [x] **First restructure committed** (`d9985ba restructuring - 1`).
- [x] **`.env` is in `.gitignore`.**

### Loose ends in the current code
- **Uncommitted work:** the Step 1–2 changes aren't committed yet. `cli/chatCli.go` and `cli/commands.go` are untracked.
- **`main.go`** prints "Hello Reading api key --" before every command, including `--help`, and loads keys before Cobra runs.
- **The `.env` path is relative** (`internal/config/.env`), so `supercode` only works when run from the project folder.
- **`go.mod`** still marks every dependency as `// indirect`.
- **`cli/output.go`** (the queue and `streamText`) is unused. It gets deleted in Step 6.
- **`internal/tools/` and the old `llm/client.go` were deleted in commit `d9985ba`.** Milestone B rebuilds tools from scratch. The old `ChatStream` in `1ec8f52` is still a useful reference for Step 6.

---

## Next Steps

Work top to bottom. Each step ends with a **Check** you can run, so you always know the project works before moving on. Don't start a step until the previous one's check passes.

| Order | Step | Why it comes here |
|-------|------|-------------------|
| 1 | ~~Make it build, then commit~~ ✅ | Nothing else can be tested until it compiles. |
| 2 | ~~Restructure the CLI~~ ✅ | The session shape everything else plugs into. |
| 3 | ~~Tidy the agent~~ ✅ | Small; makes chat mode answer properly through `Handle`. |
| 4 | ~~Errors that don't kill the session~~ ✅ | Needed before streaming, where a mid-answer failure would kill the REPL. |
| 5 | Housekeeping, then commit (shelved for later) | Small cleanups; nothing else depends on them. |
| 6 | Streaming chat replies (after Milestone B) | Tool calls arrive in pieces when streaming, so build tools without streaming first. |
| 7 | ~~Milestone B: tool calling~~ ✅ | The big one. |
| 8 | ~~Milestone C: plan and edit modes~~ ✅ | Built on the tools from B. |
| 9 | Housekeeping (Step 5), then streaming (Step 6) ← **you are here** | Everything core works; polish next. |

### Milestone A: a working session with modes (do this now)

**Done when:**
- [x] `supercode` on its own opens a session.
- [x] `supercode chat` does the same.
- [x] `supercode --mode plan` starts in plan mode.
- [x] `/mode` switches modes mid-session.
- [x] Chat mode answers through the agent.
- [x] An API error prints a message and the session keeps going.

#### Step 1. Make it build, then commit ✅
Goal: the smallest changes that get `go build ./...` passing again. Streaming is put on hold until Step 6.
- [x] **Remove the `cli` import from `llm/llmChat.go`** and the `cli.AddToOutputQueue()` call. `cli → agent → llm → cli` is an import cycle, and Go won't build it.
- [x] **Set `Stream: false` for now** in `BuildChatRequest`, and make `HandleRequest` read the body and call `ParseResponse` again, so it returns the reply text.
- [x] **Settle on one signature for `HandleRequest`:** `(string, error)`, with every caller updated (`agent/chat.go`, `agent/router.go`, `cli/sayhello.go`). Same for `ParseResponse`. (`config.LoadKeys` moves to Step 4.)
- [x] **Leave `cli/output.go` alone.** It's unused until Step 6, where it gets deleted.
- [ ] **Commit the current changes**, including the untracked `cli/chatCli.go`.

**Check:** `go build ./...` passes, and `supercode chat` answers a question. ✅

#### Step 2. Restructure the CLI
- [x] **`cli/root.go`:**
  - [x] `rootCmd` has `RunE: startSession`, so plain `supercode` opens a session like `claude` does.
  - [x] The `--mode` / `-M` flag is a **persistent flag** on root, so it works on both `supercode` and `supercode chat`.
  - [x] `startSession`: `agent.New(modeFlag)` → `runRepl(a)`. No mode `switch`.
  - [x] Make `New` return `(*Agent, error)` and have `startSession` return that error, so a bad `--mode` stops the program.
  - [x] Print that error once: the `fmt.Println` in `startSession` is gone, and `SilenceErrors: true` is set, so only `Execute()` prints it.
  - [x] Add `SilenceUsage: true` to `rootCmd`, so Cobra doesn't dump the usage text for an error that isn't a typing mistake.
- [x] **`chat` alias in `cli/chatCli.go`:** a subcommand with `RunE: startSession`. Cobra's `Aliases` field only gives a subcommand extra names, so `chat` stays its own small command.
- [ ] **`cli/repl.go`: only the session loop** (mostly done).
  - [x] `runRepl(a *agent.Agent)` takes the agent and doesn't import `cobra` or `llm`. The CLI should never call `llm` or `tools` directly. (`cli/sayhello.go` still imports `llm`; that's fine while it's a test command.)
  - [x] The loop calls `a.Handle(input)`, and on error prints `Error: …` and keeps looping.
  - [x] The prompt shows the current mode (`[auto]>`), read from `a.Mode` on every loop.
  - [x] Replace plain `exit` / `quit` with `/exit` (handled in `commands.go`). Update the welcome line to mention `/help`.
  - [x] Tidy the `"label: text"` spacing in all messages.
  - [ ] Optional: move the `//token to show on left` comment up to the prompt line; finish or delete `//every input gets`.
- [ ] **`cli/commands.go`:** `handleCommand(a, input) bool`, which returns `true` when the session should end (mostly done).
  - [x] Split with `strings.Fields`, lowercase the command, `switch` on it.
  - [x] `/help`, `/exit`, `/quit`, and the "Unknown command" fallback.
  - [x] **`/mode`:** if `len(args) == 0`, print the current mode and usage; otherwise call `a.SetMode(args[0])`.
  - [x] **Check `SetMode`'s error in `/mode <name>`.**
  - [x] Print the error itself there: `Can't switch mode: unknown mode "banana" (use auto, chat, plan or edit)`.
  - [x] **`/chat`, `/plan`, `/edit`, `/auto`:** call `a.SetMode(strings.TrimPrefix(cmd, "/"))`, not `args[0]`. `args` is empty here, which is the panic.
  - [x] **Error print:** `%e` is a float verb, so an error prints as `%!e(...)`. Use `fmt.Println("Failed to change mode:", err)`, or `%v` with a `\n`.
  - [x] **Delete the lowercase loop (lines 22–24).** It does nothing: `field` is a copy of each element, so changing it doesn't change `fields`. Lowercase inside `ParseMode` instead (`Mode(strings.ToLower(s))`). Then `/mode PLAN` and `--mode PLAN` both work, and the rule lives in `agent`.
  - [ ] Remove the extra `{ }` around the `/mode` case body. A `case` is already its own block.
- [ ] **Optional: make the mode private** so `SetMode` is the only way to change it: field `Mode` → `mode`, add `func (a *Agent) Mode() Mode`, and use `a.Mode()` in the prompt and in `commands.go`.

**Check:**
- [x] `supercode`, `supercode chat` and `supercode --mode plan` all open a session, and the prompt shows the right mode.
- [x] In auto mode, a question prints `(auto → chat)` and gets an answer.
- [x] `supercode --mode banana` prints the `ParseMode` error once and doesn't start a session.
- [x] `/mode edit` changes the prompt to `[edit]>`, and a message prints "Not built yet".
- [x] `/plan` switches to plan mode without crashing.
- [x] `/mode banana` prints the error, and the mode doesn't change.
- [x] `/help`, `/exit` and an unknown `/foo` behave as described.

#### Step 3. Tidy the agent ✅
- [x] **New `agent/prompts.go`:** move every system prompt here as a `const`.
  - `routerPrompt` from `router.go`.
  - `chatPrompt` from `chat.go` (already renamed and made a `const`).
  - Placeholder `planPrompt` and `editPrompt`: add them when plan and edit modes are built (Milestone C).
- [x] **Make every prompt name only commands that really exist.** The chat prompt now says `/mode plan` and `/mode edit`, which exist after Step 2.
- [x] **`agent/chat.go`:** if `HandleRequest` returns an error, return it; otherwise print the answer and return `nil`.

**Check:** with the API key temporarily wrong in `.env`, a chat message prints `Error: …` instead of an empty line. ✅ (`Error: API error (401): …`)

#### Step 4. Errors that don't kill the session ✅
- [x] **`llm/llmChat.go`:** replace every `log.Fatal` / `log.Fatalf` with returning an error.
  - `SendRequest` returns `(*http.Request, error)`.
  - The non-200 status check returns `fmt.Errorf("API error (%d): %s", ...)`.
  - `ParseResponse` returns its `Unmarshal` error.
  - `BuildChatRequest` returns `([]byte, error)`, and its three callers check it.
  - `io.ReadAll`'s error is checked too.
- [x] **`config.LoadKeys`:** return errors instead of calling `log.Fatal`, and flip it to `(string, error)` like the rest.

- [x] **`main.go`** checks `LoadKeys`'s error, prints it, and exits with code 1. (It moves to `PersistentPreRunE` in Step 5.)

**Check:** with a wrong API key, or with Wi-Fi off, a message prints an error and the `>` prompt comes back. ✅

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

**Reference:** a working version of this loop exists in commit `1ec8f52`, as `ChatStream` in the old `client.go`. View it with `git show 1ec8f52:internal/llm/client.go` (it was deleted in `d9985ba`, so it's no longer in `HEAD`). It has the scanner loop, the `data: ` handling, the `[DONE]` check and the `strings.Builder`. The main change is to replace its direct `fmt.Print` with the callback or writer.

**Check:** a long chat answer appears word by word, auto mode still prints `(auto → chat)` without the router's answer leaking to the screen, and `go build ./...` passes. Commit.

### Milestone B: tool calling (Phases 2–4) — step 7

Built from scratch (not restored from `1ec8f52`). Streaming stays off (`Stream: false`) for the whole milestone.

**Done when:**
- [x] In chat mode, "what does main.go do?" makes the model call `read_file`, and it answers about the real file.
- [x] A tool error (e.g. a missing file) goes back to the model, and the session keeps going.
- [x] "Create hello.txt with hi in it" asks before writing, writes on yes, and tells the model on no.

#### How tool calling works
1. You send the messages plus a list of **tool definitions** (name, description, JSON schema for the arguments).
2. Instead of text, the model can reply with **`tool_calls`**: "call `read_file` with `{"path":"main.go"}`".
3. **You** run the tool. The model never touches your machine.
4. You append the result as a `role: "tool"` message and send everything again.
5. Repeat until the model replies with plain text. That's the **agent loop**.

```json
// request: what you add
"tools": [{ "type": "function",
  "function": { "name": "read_file", "description": "Read a file in the project",
    "parameters": { "type": "object", "properties": { "path": { "type": "string" } }, "required": ["path"] } } }]

// response: choices[0]
"finish_reason": "tool_calls",
"message": { "role": "assistant", "content": null,
  "tool_calls": [{ "id": "call_abc", "type": "function",
    "function": { "name": "read_file", "arguments": "{\"path\":\"main.go\"}" } }] }

// appended before sending again
{ "role": "assistant", "content": null, "tool_calls": [ ...exactly as received... ] }
{ "role": "tool", "tool_call_id": "call_abc", "content": "package main\n..." }
```

The API is stateless: every request resends the whole conversation. That's normal (every coding agent does it); see "Keeping context small" below.

#### Step B1. Check the model supports tools ✅
- [x] `nvidia/nemotron-3.5-lightning:free` is listed at `https://openrouter.ai/api/v1/models?supported_parameters=tools`.
- [ ] If it handles tools badly in practice (made-up arguments, ignores tools), try `nvidia/nemotron-3-ultra-550b-a55b:free`.

#### Step B2. Extend the `llm` types ✅
Only types; no behaviour changes. All types now live in `llm/types.go`.
- [x] **`FunctionCall`** (named): `Name`, `Arguments` (both `string`).
  - `Arguments` is JSON **inside a string** (`"{\"path\":\"main.go\"}"`), so it must be a `string`. Each tool decodes it a second time (B4).
- [x] **`ToolCall`**: `ID`, `Type`, `Function FunctionCall`.
- [x] **`FunctionDef`** (named): `Name`, `Description` (`string`), `Parameters json.RawMessage`.
- [x] **`ToolDef`**: `Type string`, `Function FunctionDef`.
- [x] **`Message`**: add `ToolCalls []ToolCall` (`json:"tool_calls,omitempty"`) and `ToolCallID string` (`json:"tool_call_id,omitempty"`). No `omitempty` on `Content`.
- [x] **`ChatRequest`**: add `Tools []ToolDef` (`json:"tools,omitempty"`).
- [x] **`ChatResponse`**: add `FinishReason string` (`json:"finish_reason"`) next to `Message` in the choice.
- [ ] Optional: rename the `Choices` type to `Choice` (one element of `[]Choice`).

Things to remember:
- Tags need quotes: `json:"id"`, not `json:id`. `go vet ./...` catches it.
- Only exported (capitalized) fields are encoded or decoded; lowercase ones are silently skipped.
- `omitempty` drops zero values (`""`, `0`, `false`, `nil` slices/maps/pointers), so requests without tools stay exactly as they are today.
- `"content": null` decodes into a `string` as `""` without an error. If a provider later rejects `""` alongside tool calls when you send it back, switch `Content` to `*string`.

**Check:** `go build ./...` and `go vet ./...` pass, and `supercode chat` answers "hi" exactly as before.

#### Step B3. Requests built from messages ✅
- [x] **`BuildRequest(messages []Message, tools []ToolDef)`.** The caller owns the slice, so the conversation can grow inside one turn.
- [x] **`ParseResponse` and `HandleRequest` return the whole `Message`**, not just `Content`, so the loop can see `ToolCalls`. Callers use `answer.Content`.
- [x] **Return an error when `Choices` is empty**, instead of a blank reply.
  - [ ] Drop the `%w` from that error: `err` is `nil` there, so it prints `%!w(<nil>)`.
- [x] **`BuildChatRequest` is a small helper for the two-strings call** that ends with `return BuildRequest(messages, nil)`.

**Check:** `go build ./...` passes, and chat still answers.

#### Step B4. The `tools` package ✅
Files: `tools.go` (Tool + Registry), `path.go` (`safePath`), `readfile.go`, `listfiles.go`, `writefile.go`, `runcommand.go`, `shell_windows.go` / `shell_other.go`, `tools_test.go`.
- [x] **`Tool` type** (`tools.go`): `Name`, `Description`, `Parameters json.RawMessage`, `Run func(args string) (string, error)`, and `NeedsApproval bool`.
  - `Run` is a field holding a function. Each tool stores its own (`Run: readFile`, no `()`), and the registry calls it with `t.Run(args)`.
  - `NeedsApproval` marks tools that change things; the agent asks the user before running them (B7).
- [x] **`Registry`** (`tools.go`): the lookup from the name the model sends to the Go code that runs.
  - `byName map[string]Tool` for lookup, plus an `order []Tool` slice so the tool list sent to the model is always in the same order (map order is random, which would break prompt caching).
  - `NewRegistry(tools ...Tool) *Registry`, `List() []Tool`, `Get(name) (Tool, bool)`, `Run(name, args) (string, error)`.
  - `byName` / `order` are lowercase; a duplicate name is skipped (the first one wins).
- [x] **`tools` doesn't import `llm`.** `agent` converts tools to `[]llm.ToolDef` (`toolDefs` in `loop.go`).
- [x] **`safePath`** (`path.go`): `os.Getwd()` → `filepath.Abs` → `filepath.Rel`; refuses a first part of `..` (split on `filepath.Separator`, so `..notes.txt` is still allowed) and a `Rel` error (another drive). Errors name the path and tell the model to use a path relative to the project root.
- [x] **`read_file`**: decode args → require `path` → `safePath` → `os.Stat` (a directory says "use list_files") → `os.ReadFile` → truncate at 20000 bytes with a note.
- [x] **`list_files`** (AI-written): optional path (defaults to `.`; empty args `""` treated as `{}`), folders end with `/`, `.git` skipped, capped at 500 entries, `(empty directory)` when empty, a file path says "use read_file".
- [x] **`tools_test.go`** (AI-written): `safePath`, the registry, all four tools (`write_file` runs in a temp folder via `t.Chdir(t.TempDir())`), and the approval flags. `go test ./internal/tools -v`.
- [ ] Optional: turn "file not found" into a clean `%q does not exist` with `errors.Is(err, fs.ErrNotExist)` (Windows otherwise adds `GetFileAttributesEx …`).

#### Step B5. The agent loop (`agent/loop.go`) ✅ (AI-written)
```
messages = [system prompt, user input]
for step := 0; step < maxSteps (20); step++ {
    reply := send(messages, toolDefs)
    append reply to messages                 // even when it has tool calls
    if no tool calls: print reply.Content, return
    for each call in reply.ToolCalls:
        print "→ read_file(path=main.go)"
        ask the user first if the tool NeedsApproval
        result, err := registry.Run(call.Function.Name, call.Function.Arguments)
        if err: result = "error: " + err.Error()   // back to the model, not out of the turn
        append {role: "tool", tool_call_id: call.ID, content: result}
}
return error "stopped after 20 steps without a final answer"
```
- [x] `runLoop(systemPrompt, input, reg)`: the loop above, used by every mode handler.
- [x] The assistant message is appended **with** its `tool_calls` before the tool results, and **every** call ID gets a result.
- [x] Unknown tool names and tool errors become `error: …` results, so the model can adjust.
- [x] Progress output: `→ name(key=value, …)` from `describeCall` (values shortened to 60 characters), then the first line of the result or the error.
- [x] Text the model sends alongside tool calls ("Let me look at…") is printed too.

#### Step B6. Hook up the modes ✅
- [x] **Registries are built in `agent.New`:** `chatTools` = `read_file`, `list_files`; `editTools` = those plus `write_file`, `run_command`.
- [x] **Chat** (`chat.go`): `runLoop(chatPrompt, input, a.chatTools)`. The chat prompt now says it can look at files with read-only tools.
- [x] **Edit** (`edit.go`): `runLoop(editPrompt, input, a.editTools)`, with a new `editPrompt` (read before changing, `write_file` replaces whole files, stop if the user declines, short summary at the end). Basic version; Milestone C adds `edit_file` and diffs.
- [x] **Router:** `classify` stays a plain call with no tools.
- [ ] **Plan** stays a stub until Milestone C.

#### Step B7. Tools that change things, with confirmation ✅
- [x] **`write_file`** (`writefile.go`): `path` + `content` (a `*string`, so an empty file is allowed but a missing `content` is an error); `safePath`; refuses a directory; creates missing folders; reports `created` / `overwrote` and the size. `NeedsApproval: true`.
- [x] **`run_command`** (`runcommand.go`): runs in the project root, combined stdout + stderr, `[exit code N]` at the end (a non-zero exit is a normal result, not an error), 60 s timeout via `exec.CommandContext` + `WaitDelay`, output truncated at 20000 bytes, no stdin. `NeedsApproval: true`.
  - The shell is chosen at build time: `shell_windows.go` runs `cmd /C` and passes the command line as is (`SysProcAttr.CmdLine`, because Go's normal argument quoting breaks `cmd.exe`); `shell_other.go` (`//go:build !windows`) runs `sh -c`.
- [x] **Confirmation:** `Agent.Confirm func(question string) bool`, set in `runRepl` and reading from the **same `bufio.Scanner`** as the REPL. `[y/N]`: only `y` / `yes` approve; with no `Confirm` set, everything is declined.
- [x] **On "no",** the model gets "The user declined this action. Do not retry it; ask the user what they want instead."

**Checked live:** chat reads `main.go` and answers about it; a missing file goes back to the model; edit mode creates a file on `y`, doesn't on `n` and the model asks what to do instead; `run_command` runs `go version`; auto mode routes to chat and uses `list_files`.

#### Keeping context small
Every request resends the full conversation; coding agents keep that cheap rather than avoiding it.
- **Truncate tool output** and cap loop iterations (already in B4/B5).
- **Only append to the end of the messages, and keep the system prompt and tool list identical between requests.** That lets providers cache the repeated prefix (prompt caching). On OpenRouter it's automatic for some providers; Anthropic models need `cache_control` markers; many free models don't cache.
- **Compaction** once history spans turns: read `usage.prompt_tokens` from responses and summarise the conversation when it nears the model's limit.
- **Read less:** line ranges for `read_file`, a grep-style `search` tool returning matching lines only.
- **Subagents** (much later): a side task runs in its own conversation and only returns a summary.

### Milestone C: plan and edit modes — step 8 ✅ (AI-written)

**Done when:**
- [x] Plan mode looks at the repo, prints a plan, saves it to `PLAN.md`, and asks before overwriting an existing one.
- [x] Edit mode changes part of a file with `edit_file`, showing a diff before you approve.
- [x] The agent remembers earlier messages in the session, and `/clear` makes it forget.

#### C1. Approval previews (diffs)
- [x] **`Tool.Preview func(args string) (string, error)`**: optional; tools that need approval use it to show what will happen. The agent prints the preview, then asks `Allow <tool>? [y/N]`.
  - If the preview itself fails (e.g. `old_text` not found), the error goes straight back to the model and you're never asked.
- [x] **`diffPreview(before, after)`** (`tools/diff.go`): finds the first and last lines that differ (common prefix / suffix), and shows that one block as `-` / `+` lines with 2 lines of context and a `@@ line N: -x +y @@` header, capped at 40 lines. Not a full diff algorithm: several separate changes in one file show as one bigger block.
- [x] Previews: `edit_file` and `write_file` show a diff (`(new file)` for a new one); `run_command` shows `$ <full command>`.

#### C2. `edit_file` (`tools/editfile.go`)
- [x] Args: `path`, `old_text`, `new_text`. `old_text` must appear **exactly once**: 0 matches → "read the file again and copy the exact text"; 2+ → "include more surrounding lines". Empty `old_text` → "use write_file".
- [x] `planEdit` does all the checks and builds the new contents in memory; `Run` writes it, `Preview` diffs it, so both always agree.
- [x] **CRLF files:** if the file uses Windows line endings, `old_text` / `new_text` are converted to `\r\n` first, because the model always sends `\n`.
- [x] `write_file` refactored: `SaveFile(path, content)` and `FileExists(path)` are exported, so plan mode can save `PLAN.md` through `tools` (and `safePath`).

#### C3. Plan mode (`agent/plan.go`)
- [x] `runLoop(planPrompt, input, a.chatTools)`: read-only tools, so it can look but not touch. `runLoop` now returns the final text as well.
- [x] The final reply is the plan: saved with `tools.SaveFile("PLAN.md", …)`; if `PLAN.md` exists, asks `PLAN.md already exists. Overwrite it? [y/N]` first. An empty plan is an error.
- [x] `planPrompt`: read-only, no saving tool (SuperCode saves the reply), plan format (goal, numbered steps with files, "How to check", "Assumptions").
- [x] `editPrompt` now mentions `edit_file` (preferred for existing files) and reading `PLAN.md` when the user refers to "the plan".

#### C4. Conversation history (`agent/history.go`)
- [x] `Agent.history []llm.Message`: every message except the system prompt (user, assistant, tool calls, tool results). Shared by all modes, so "plan X" → `/edit` → "do the plan" works.
- [x] `runLoop` sends `[system prompt for this mode] + history + new user message`, and saves the history only when the turn finishes; a failed turn (API error, step limit) is forgotten, so history never has a half-finished tool exchange.
- [x] **Size limit:** above `maxHistoryChars` (120000 characters), the oldest **whole turns** are dropped (from one user message to the next), so a tool call is never separated from its result. The latest turn is always kept.
- [x] **`/clear`** (`cli/commands.go`): `a.ClearHistory()`, prints how many of your messages were forgotten. Listed in `/help`.
- [x] `agent/agent_test.go`: history trimming, `/clear`, `describeCall`, tool order.

**Checked live:** chat remembered "teal" across messages and forgot it after `/clear`; plan mode wrote `PLAN.md` and kept it when the overwrite was declined; the 429 rate-limit error printed and the session continued.

#### Later
- [ ] **Stop the agent printing directly:** pass an `io.Writer` or callbacks from the CLI (also needed for Step 6).
- [ ] **Compaction instead of dropping:** summarise old turns with the model rather than deleting them.
- [ ] **A real diff** (e.g. Myers / LCS) when one edit changes several far-apart places.
- [ ] **Let the agent change its own mode** (e.g. a `switch_mode` tool) via `SetMode`.
- [ ] **Unit tests for `runLoop`** with a fake LLM (would need `llm` behind an interface or a function field).
- [ ] **Free-tier rate limit** (20 requests/min): a long tool loop can hit 429. Retry with a short wait, or switch models.

### Multiple API keys and models per task ✅ (AI-written)
**`.env`:** `SUPERCODE_API_KEY1` … `SUPERCODE_API_KEYN` (any number, plain `SUPERCODE_API_KEY` also works; duplicates and empty values are skipped), plus `ORCHESTRATOR_MODEL`, `PLAN_MODEL`, `EDIT_MODEL` (each defaults to the orchestrator model, which defaults to `nvidia/nemotron-3.5-lightning:free`).

- [x] **`config`:** `APIKeys []string` (sorted by number), `OrchestratorModel`, `PlanModel`, `EditModel`, `FallbackModel = "openrouter/free"`. `LoadKeys` returns `([]string, error)`.
- [x] **One entry point, `llm.Complete(task, messages, tools)`** (`llm/pool.go`). Tasks: `TaskRouter`, `TaskChat` → orchestrator model; `TaskPlan` → plan model; `TaskEdit` → edit model.
- [x] **Model chain per task:** its own model → the other two configured models → `openrouter/free`. A model that fails is skipped for a while: 30 min for 404 (doesn't exist), 2 min for 5xx or a provider-side 429; 400 and network errors just move on.
- [x] **Key choice:** the usable key with the **fewest requests in the last minute** (ties: least recently used), so the 4 keys share the load evenly. A key with 20 requests in the last minute is rested before OpenRouter's free limit is hit.
- [x] **Key errors:** 429 from OpenRouter (account limit) → that key rests until the reset time OpenRouter sends (`X-RateLimit-Reset`), and the next key is tried on the same model; 401/402/403 → the key is disabled for the session.
- [x] **All keys resting:** waits up to 90 s for the first one to come back (not for the router, which just falls back to chat), otherwise returns `all models and API keys failed, last error: …`.
- [x] **`APIError`** (`llm/llmChat.go`) keeps the status, OpenRouter's message, whether it came from the upstream provider, and the retry time. A 200 response that carries an `error` object is treated as a failure too. HTTP timeout: 3 minutes.
- [x] **Progress messages** go through `llm.Notify` (set by `agent.New`), e.g. `(key 1 is rate-limited for 2h46m; switching key)`, `(plan model failed: …; trying …)`, `(answered by openrouter/free)`.
- [x] **`/status`** shows the model for each task and every key's state (masked to its last 4 characters).
- [x] Tests: `llm/pool_test.go` (fake OpenRouter server: rotation, 429 → next key, 401 → disabled, broken models → backup and skipped next time, upstream 429, router doesn't wait, per-minute budget, reset parsing), `config/config_test.go` (key collection).

### Current structure (after Milestone C)
```
internal/
  cli/
    root.go       rootCmd: RunE startSession, persistent --mode flag, PersistentPreRunE loads keys
    repl.go       runRepl: prompt, read line, command vs message, call a.Handle, print errors; sets a.Confirm
    commands.go   slash commands: /mode, /chat /plan /edit /auto, /clear, /help, /exit
    chatCli.go    `chat` subcommand: alias that starts the same session
    sayhello.go
  agent/
    agent.go      Agent struct (Mode, Confirm, tool registries, history) + New + Handle (dispatch per message)
    mode.go       Mode type + ParseMode + SetMode
    router.go     classify (auto mode)
    prompts.go    every system prompt (router, chat, plan, edit)
    loop.go       runLoop (the agent loop), runTool (preview + approval), toolDefs, progress labels
    history.go    history, saveHistory (size limit), ClearHistory
    agent_test.go
    chat.go       chat mode: runLoop with read-only tools
    plan.go       plan mode: runLoop with read-only tools, save the reply to PLAN.md
    edit.go       edit mode: runLoop with read, edit_file, write_file and run_command
  llm/
    llmChat.go    BuildRequest, send (one attempt), ParseResponse, APIError
    pool.go       Complete: model per task, key rotation, fallback, Status
    types.go      Message, ToolCall, ToolDef, ChatRequest, ChatResponse
  config/
    config.go     endpoint, models per task, API keys, LoadKeys
  tools/
    tools.go      Tool + Registry
    path.go       safePath
    readfile.go, listfiles.go, editfile.go, writefile.go (+ SaveFile, FileExists), runcommand.go
    diff.go       diffPreview for approval previews
    shell_windows.go, shell_other.go   how run_command starts a shell
    tools_test.go
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
- [x] Chat history
- [x] Tool framework

### Tools
- [x] read_file
- [ ] search
- [x] list_files
- [x] write_file
- [x] execute
- [ ] finish

### Agent
- [x] LLM API integration (OpenRouter, OpenAI-compatible)
- [x] Structured JSON tool calls
- [x] Agent execution loop
- [x] Tool result feedback

### Upgrades
- [x] apply_patch (as `edit_file` find-and-replace)
- [ ] Automatic build verification
- [ ] Repository tree context

### Stretch
- [ ] Git integration
- [ ] Symbol search
- [ ] Multi-file editing

---

## Definition of Done

The project is complete when you can run a single command with a coding task, and the agent autonomously reads the repo, makes a patch-based edit, optionally runs a command to verify, and reports back — all from the terminal.
