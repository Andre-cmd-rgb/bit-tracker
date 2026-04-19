# bit-tracker

Offline personal life OS in your terminal — goals, mood, journal, todos, and a
warm pixelated AI pet named **Bit**. Everything runs locally: Bit is a
`llama-server` subprocess managed automatically, speaking the OpenAI-compatible
API over `127.0.0.1`.

## Install

```sh
go install github.com/andre-cmd-rgb/bit-tracker@latest
# or, from source:
git clone https://github.com/andre-cmd-rgb/bit-tracker && cd bit-tracker
go build -o bit-tracker
./bit-tracker
```

First launch offers to download a model and the `llama-server` binary
into `~/.local/share/bit-tracker/`. You can skip; the rest of the app still
works. Zero CGO, pure Go throughout.

## Keybindings

| View            | Keys                                                     |
|-----------------|----------------------------------------------------------|
| Global          | `1`–`5` switch views · `q` quit                          |
| Goals           | `n` new · `e` edit · `d` delete · `+`/`-` ±5% · `c` 100% |
|                 | `/` filter · `t` toggle Todos sub-tab                    |
| Todos (in Goals)| `n` new · `Space` toggle · `p` priority · `d` delete     |
| Mood            | `1`–`5` score · `Enter` log · `n` add note               |
| Journal         | `n` new · `Enter` open · `Esc` save · `Ctrl+B` reflect   |
| Bit Chat        | type & `Enter` · `Esc` back                              |

## Config

`~/.config/bit-tracker/config.toml` — auto-written after first run:

```toml
model_name    = "phi-3.5-mini-instruct-Q4_K_M.gguf"
model_path    = "~/.local/share/bit-tracker/models/phi-3.5-mini-instruct-Q4_K_M.gguf"
llama_bin     = "~/.local/share/bit-tracker/bin/llama-server"
first_run_done = true
username      = "you"
```

## Models

| Tier   | RAM  | Model                               | Size  |
|--------|------|-------------------------------------|-------|
| low    | <4GB | SmolLM2 360M Instruct (Q4_K_M)      | ~230MB|
| mid    | 4–8GB| Qwen2.5 0.5B Instruct (Q4_K_M)      | ~390MB|
| high   | >8GB | Phi-3.5 Mini Instruct (Q4_K_M)      | ~2.2GB|

Bit embeds JSON action blocks in its replies; the app parses them and applies:

```
{"action":"add_goal","data":{...}}
{"action":"log_mood","data":{"score":4,"note":"..."}}
{"action":"add_journal","data":{"text":"..."}}
{"action":"add_todo","data":{"text":"","priority":2}}
{"action":"update_goal","data":{"id":1,"progress":75}}
```

## Data

Everything lives in `~/.local/share/bit-tracker/bit-tracker.db` (SQLite/WAL).
Delete that file to reset.
