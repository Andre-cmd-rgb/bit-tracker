# bit-tracker

A local, terminal-only diary. One plain text body per day, structured metrics
on the side, and a grounded local-AI layer that can rewrite, reflect, or wake
you up when the pattern turns sour. No cloud. No Ollama. No API. No server.

## Philosophy

- One entry per day, auto-opened with the weekday, date, and start time.
- Metrics are deterministic, not vibes — study minutes, scroll minutes,
  project minutes, mood 1–10.
- Good days and bad days are computed from those numbers plus a small set of
  blunt phrase matches. The model never invents the label.
- After two bad days in a row the tone sharpens. After four bad days in a
  rolling week it escalates further. It is not a therapist.
- Everything is a file on your machine.

## Features

- auto-created today entry with `monday, 20 april 2026 / started at 07:14 / today:`
- editable body via `bubbletea` + `bubbles/textarea`
- structured metrics (mood, study, scroll, project, tags, project name/note/done)
- deterministic good day / bad day scoring
- streak detection and tone escalation (normal → reflective → harsh → intervention)
- history browser with keyword + `#tag` search
- local AI abstraction with three modes — `rewrite`, `reflect`, `wake_up` —
  plus a free chat fallback
- the app works with no model at all; the stub engine produces short,
  deterministic summaries built from your own numbers
- Markdown + HTML export of one entry, this week, or everything — raw or
  cleaned
- weekly / monthly / yearly **wrapped** view with stat cards, bars, top tags,
  top themes, project roll-ups, and a blunt one-line takeaway

## Stack

- Go (≥ 1.24)
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) + Bubbles + Lip Gloss
- `modernc.org/sqlite` — pure-Go SQLite, no CGO
- optional `llamacpp` build tag for plugging in a direct llama.cpp Go binding

## Local model

The app ships with an offline fallback engine so it runs with or without a
model. To use real local inference, drop a GGUF file somewhere and point at
it:

```sh
bit-tracker --model ~/.local/share/bit-tracker/models/your-model.gguf
# or
export BIT_TRACKER_MODEL=~/.local/share/bit-tracker/models/your-model.gguf
bit-tracker
```

The default search path is `~/.local/share/bit-tracker/models/model.gguf`.
The model is loaded once and reused for the life of the process.

A direct llama.cpp binding is scaffolded behind the `llamacpp` build tag —
see `internal/ai/llamacpp.go`. The rest of the app never sees the binding's
types, so you can swap implementations without touching the UI.

```sh
go build -tags llamacpp -o bit-tracker ./cmd/bit-tracker
```

When no binding is compiled or the model is missing, `rewrite` collapses the
raw entry, `reflect` summarises your own recent metrics, and `wake_up` reads
them back at you. Nothing in the app breaks if AI is unavailable.

## Build

```sh
make build          # builds ./bit-tracker
make run            # build + run
make test           # run the test suite
make fmt            # gofmt -w .
```

Or directly:

```sh
go build -o bit-tracker ./cmd/bit-tracker
./bit-tracker
```

## Run

```sh
./bit-tracker [--data DIR] [--exports DIR] [--model PATH]
```

Defaults:

- `--data`    → `$XDG_CONFIG_HOME/bit-tracker` (or `~/.config/bit-tracker`)
- `--exports` → `~/bit-tracker-exports`
- `--model`   → `$BIT_TRACKER_MODEL` if set, else `~/.local/share/bit-tracker/models/model.gguf`

## Shortcuts

| Where     | Keys                                                                |
|-----------|---------------------------------------------------------------------|
| Global    | `t` today · `h` history · `c` chat · `e` export · `w` wrapped       |
|           | `Tab` / `Shift+Tab` cycle views · `q` / `Ctrl+C` quit               |
| Today     | `Enter` / `i` start editing · `Esc` save & leave · `Ctrl+S` save    |
|           | `r` rewrite via AI · `+` / `-` mood · `s` +15 study · `p` +15 proj  |
|           | `o` +15 scroll · `x` toggle project-done                            |
| History   | `↑/↓` move · `Enter` open · `/` search · `#tag` tag filter          |
|           | in entry view: `m` export Markdown · `H` export HTML · `Esc` back   |
| Chat      | `i` type prompt · `Enter` send · `Esc` unfocus · `Ctrl+L` clear     |
|           | `r` rewrite today · `f` reflect (14 days) · `u` wake up (7 days)    |
| Export    | `↑/↓` move · `Enter` run · output path shown at the bottom          |
| Wrapped   | `w` week · `m` month · `y` year                                     |

## Export

Every export renders a clean, printable document with:

- weekday + date + start time
- body (raw or cleaned)
- metric list
- tags, project name, project note, completion state
- optional AI text (rewrite or reflection) as a separate section

Formats: Markdown and HTML. Range exports paginate entries with `---` / `<hr>`
separators. Files go to `--exports` (default `~/bit-tracker-exports`).

## Wrapped

`w` switches to the recap view. Week / month / year re-aggregate on demand.
Each period shows:

- days journaled, average mood, good/bad/neutral days
- total time: study vs project vs scroll, with bars
- longest good + longest bad streak
- meaningful study days (≥ 90 min)
- project roll-up (touched + completed)
- top tags and top themes (phrase matches in body text)
- blunt one-line takeaway

No broken charts, no trophies. Just the numbers you wrote down, read back.

## Data

One SQLite file: `$data/bit-tracker.db`. Tables: `entries` (one row per date)
and `ai_messages` (optional interaction log). Delete the file to reset.

## Non-goals

No cloud sync. No auth. No plugins. No vector DB. No embeddings. No voice.
No telemetry. No agent loops. No multi-user. No web UI. No Electron. No
Ollama. No remote API. No external model server. No giant config system.

## Limitations

- The bundled fallback is not a language model. Turn on `-tags llamacpp` and
  wire in a binding if you want real generations.
- Phrase detection is intentionally shallow — it is a guardrail, not an NLP
  stack.
- Single local user, single machine.
