# SuperCode 🦹

A CLI-based AI coding agent written in Go. Give it a task, and it autonomously explores your repository, edits code using patches, runs terminal commands, and completes the job — all from your terminal.

Built with Go. Powered by any OpenAI-compatible LLM API.

---

## Features

- **Autonomous coding agent** — reads files, writes patches, runs commands, and verifies changes without hand-holding
- **Agentic tool-calling loop** — the model decides what to do, calls tools, reads results, and keeps going until the task is done
- **Streaming responses** — real-time token output so you can follow the agent's thinking as it works
- **Any LLM provider** — works with OpenRouter, OpenAI, Anthropic, or any OpenAI-compatible API out of the box
- **Patch-based editing** — no full file rewrites; the agent finds the exact text to change and replaces it surgically
- **Safety first** — confirmation prompts before running commands or editing files
- **Built-in tools:**
  - `read_file` — read file contents
  - `list_files` — explore directory structure
  - `write_file` — create or overwrite files
  - `run_command` — execute shell commands
  - `search` — search for patterns across files
  - `apply_patch` — apply patch-based edits to files

---

## Quick Start

### Prerequisites

- [Go 1.21+](https://go.dev/dl/)
- An API key from [OpenRouter](https://openrouter.ai) (free tier available), [OpenAI](https://platform.openai.com), or any compatible provider

### Install

```bash
git clone https://github.com/supersanchayrx/SuperCode.git
cd SuperCode
go build -o supercode .
```

### Set up your API key

**Linux / macOS:**
```bash
export SUPERCODE_API_KEY="sk-or-v1-your-key-here"
```

**Windows (CMD):**
```cmd
set SUPERCODE_API_KEY=sk-or-v1-your-key-here
```

**Windows (PowerShell):**
```powershell
$env:SUPERCODE_API_KEY = "sk-or-v1-your-key-here"
```

### Run it

```bash
# Give it a coding task
supercode run "add a function that reverses a string to utils.go"

# Specify a working directory
supercode run -d ./myproject "fix the failing tests"

# Use a specific model
supercode run -m "qwen/qwen3-coder:free" "refactor the error handling in main.go"
```

---

## Configuration

SuperCode is configured via environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `SUPERCODE_API_KEY` | Your LLM provider API key | *(required)* |
| `SUPERCODE_MODEL` | Model to use | `nvidia/nemotron-3-nano-30b-a3b:free` |
| `SUPERCODE_BASE_URL` | API base URL | `https://openrouter.ai/api/v1` |

### Supported Providers

| Provider | Base URL | Notes |
|----------|----------|-------|
| [OpenRouter](https://openrouter.ai) | `https://openrouter.ai/api/v1` | Free models available |
| [OpenAI](https://platform.openai.com) | `https://api.openai.com/v1` | GPT-4o, GPT-4, etc. |
| [Ollama](https://ollama.com) (local) | `http://localhost:11434/v1` | Run models locally |

---

## How It Works

SuperCode runs an **agentic loop**:

```
1. You give it a task
2. It sends the task + tool definitions to the LLM
3. The model decides what to do:
   ├── Call a tool (read_file, run_command, etc.)
   │   → Execute the tool
   │   → Feed the result back to the model
   │   → Go to step 3
   │
   └── Return a final answer
       → Print it and stop
```

The agent keeps looping — exploring, editing, and verifying — until it's satisfied the task is complete, or it hits the iteration limit.

---

## Project Structure

```
supercode/
  main.go                  # Entry point
  internal/
    cli/                   # Cobra CLI commands
      root.go              # Root command
      run.go               # Run subcommand
    agent/                 # The agentic reasoning loop
    llm/                   # LLM API client (chat + streaming)
      client.go            # HTTP client, message types
    tools/                 # Tool definitions and execution
    config/                # Configuration from env vars
      config.go
```

---

*Built from scratch as a learning project. Inspired by tools like [OpenCode](https://github.com/opencode-ai/opencode), [Aider](https://github.com/paul-gauthier/aider), and [Claude Code](https://claude.ai).*