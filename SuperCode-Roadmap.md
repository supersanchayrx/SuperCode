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
| 7 | Milestone B: tool calling ← **you are here** (B4) | The big one. |
| 8 | Milestone C: plan and edit modes | Built on the tools from B. |

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
- [ ] In chat mode, "what does main.go do?" makes the model call `read_file`, and it answers about the real file.
- [ ] A tool error (e.g. a missing file) goes back to the model, and the session keeps going.
- [ ] "Create hello.txt with hi in it" asks before writing, writes on yes, and tells the model on no.

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

#### Step B4. The `tools` package, one tool at a time ← you are here
Files: `tools.go` (Tool + Registry), `path.go` (`safePath`), `readfile.go`, `listfiles.go`, `tools_test.go`.
- [x] **`Tool` type** (`tools.go`): `Name`, `Description`, `Parameters json.RawMessage`, and `Run func(args string) (string, error)`.
  - `Run` is a field holding a function. Each tool stores its own (`Run: readFile`, no `()`), and the registry calls it with `t.Run(args)`.
- [x] **`Registry`** (`tools.go`): the lookup from the name the model sends to the Go code that runs.
  - `byName map[string]Tool` for lookup, plus an `order []Tool` slice so the tool list sent to the model is always in the same order (map order is random, which would break prompt caching).
  - `NewRegistry(tools ...Tool) *Registry`: builds the map with `make` and fills both fields.
  - `List() []Tool`: the "menu" the agent turns into `[]llm.ToolDef`.
  - `Run(name, args string) (string, error)`: finds the tool by name and calls its `Run`; `unknown tool "x"` error otherwise.
  - Each mode gets its own registry (B6), e.g. chat: `NewRegistry(ReadFile(), ListFiles())`.
  - [x] `byName` / `order` are lowercase, so only `NewRegistry` can fill them.
  - [x] A tool whose name is already registered is skipped (the first one wins).
- [x] **`safePath(p string) (string, error)`** (`path.go`): see below. Splits `rel` on `filepath.Separator` and refuses a first part of `..` (so `..notes.txt` is still allowed). Error messages name the path and tell the model to use a path relative to the project root.
- [x] **`tools_test.go`** (AI-written): table-driven `TestSafePath`, plus registry tests (`Run`, unknown tool, list order + duplicate skip, empty registry). Run with `go test ./internal/tools -v`.
- [ ] **`readfile.go`** ← next: `ReadFile() Tool` + `readFile(args string) (string, error)`.
- [ ] **`listfiles.go`**: `ListFiles() Tool` + `listFiles(args string) (string, error)`.
- [ ] Add tests for both tools to `tools_test.go`.
- [ ] **`tools` must not import `llm`.** `agent` converts tools to `[]llm.ToolDef` (a small loop). Same layering rule as before.
- [ ] **`read_file` first**, then **`list_files`**.
  - Each tool unmarshals its `Arguments` string into its own small args struct.
  - Reject paths outside the working directory with `safePath`: `os.Getwd()` → `filepath.Abs(p)` → `filepath.Rel(root, abs)`; refuse `..` or anything starting with `..` + separator, and treat a `Rel` error (another drive on Windows) as outside.
  - Truncate big output (e.g. 20 KB + "…truncated") so one file doesn't blow the context.
  - Return errors; never `log.Fatal`.

**Check:** calling `read_file` from a throwaway test with `{"path":"main.go"}` returns the file; `{"path":"../x"}` returns an error.

#### Step B5. The agent loop (`agent/loop.go`)
```
messages = [system prompt, user input]
for i := 0; i < maxIterations; i++ {
    reply := send(messages, toolDefs)
    append reply to messages                 // even when it has tool calls
    if no tool calls: print reply.Content, return
    for each call in reply.ToolCalls:
        print "→ read_file(main.go)"
        result, err := registry.Run(call.Function.Name, call.Function.Arguments)
        if err: result = "error: " + err.Error()   // back to the model, not out of the turn
        append {role: "tool", tool_call_id: call.ID, content: result}
}
return error "stopped after N steps"
```
- [ ] Write the loop with a max number of iterations.
- [ ] Append the assistant message **with** its `tool_calls` before the tool results; otherwise the API rejects the tool messages.
- [ ] Answer **every** call ID (the model can request several tools at once).
- [ ] An unknown tool name becomes an error result, not a crash.
- [ ] Print a short progress line per tool call.

**Check:** chat mode, "what does main.go do?" prints `→ read_file(main.go)` and then a real answer.

#### Step B6. Hook up the modes
- [ ] **Chat** goes through the loop with the read-only registry (`read_file`, `list_files`).
- [ ] **Router:** `classify` stays a plain call with no tools.
- [ ] **Plan / edit** stay stubs until Milestone C, but will call the same loop with a different registry:
  - plan: read-only, plus writing `PLAN.md`;
  - edit: read, write and `run_command`.
- [ ] Update the chat prompt to say it can read files.

#### Step B7. Tools that change things, with confirmation
- [ ] **`write_file`**, then **`run_command`** (with a timeout via `exec.CommandContext`).
- [ ] **Ask before running them.** The CLI gives the agent a function, e.g. a `Confirm func(question string) bool` field on `Agent`, set in `startSession`.
  - It must read from the **same `bufio.Scanner`** the REPL uses. A second reader on `os.Stdin` can swallow the next input line.
- [ ] **On "no",** send a "user declined" tool result so the model can change course.

**Check:** the third "Done when" item above.

#### Later (after B works)
- [ ] **Stop the agent printing directly:** pass an `io.Writer` or callbacks from the CLI (also needed for Step 6).
- [ ] **Let the agent change its own mode** (e.g. a `switch_mode` tool) via `SetMode`. The REPL reads the mode for the prompt, so it stays in sync.
- [ ] **Conversation history across turns:** keep `messages` on the `Agent` instead of rebuilding it each turn (Milestone C).

#### Keeping context small
Every request resends the full conversation; coding agents keep that cheap rather than avoiding it.
- **Truncate tool output** and cap loop iterations (already in B4/B5).
- **Only append to the end of the messages, and keep the system prompt and tool list identical between requests.** That lets providers cache the repeated prefix (prompt caching). On OpenRouter it's automatic for some providers; Anthropic models need `cache_control` markers; many free models don't cache.
- **Compaction** once history spans turns: read `usage.prompt_tokens` from responses and summarise the conversation when it nears the model's limit.
- **Read less:** line ranges for `read_file`, a grep-style `search` tool returning matching lines only.
- **Subagents** (much later): a side task runs in its own conversation and only returns a summary.

### Milestone C: plan and edit modes — step 8
- [ ] **Plan mode:** call the model with the planner prompt (read-only tools, so it can look at the repo), write the result to `PLAN.md`, and ask before overwriting an existing one.
- [ ] **Edit mode:** an `edit_file` tool that does find-and-replace (see Phase 3), shows what will change, and asks before writing.
- [ ] **Conversation history across messages** on the `Agent` (a `[]Message`), plus a `/clear` command to reset it.

### Target structure after Milestone A
```
internal/
  cli/
    root.go       rootCmd: RunE startSession, persistent --mode flag, PersistentPreRunE loads keys
    repl.go       runRepl: prompt, read line, command vs message, call a.Handle, print errors
    commands.go   slash commands: /mode, /chat /plan /edit /auto, /help, /exit
    chatCli.go    `chat` subcommand: alias that starts the same session
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
  tools/          built from scratch in Milestone B (Tool, Registry, read_file, list_files, write_file, run_command)
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
