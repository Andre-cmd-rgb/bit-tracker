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

- three tabs: **landing · diary · chat**. Export, settings, models, and help
  are overlays you can summon from anywhere
- guided **first-run setup** — welcome, pet, palette, and model in four pages;
  runs exactly once, persists to `settings.json`
- landing page with an ASCII pet that speaks tone-aware lines, a today stats
  sidebar, and a mini this-week recap
- configurable pet — eight characters (penguin · robot · cat · bunny · ghost ·
  fox · dragon · owl), ten hats, ten eye expressions, shimmery **shiny mode**
  that cycle-paints the heart glyph through a rainbow, and a pet name
- five colour themes (purple · green · amber · cyan · rose) — everything
  persists to `settings.json`
- in-app model downloader: pick from a curated list of small GGUFs
  (SmolLM2 135M / 360M, Qwen2.5 0.5B / 1.5B, TinyLlama 1.1B, Gemma 3 1B),
  see live progress, and activate with one keystroke — the model is also
  selectable inline from the settings overlay
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
- Markdown + HTML export of one entry, this week, or everything
- weekly recap embedded on the landing page; scrollable list in the diary tab
- runs on Linux, macOS, and Windows — pure-Go SQLite, no CGO

## Stack

- Go (≥ 1.24)
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) + Bubbles + Lip Gloss
- `modernc.org/sqlite` — pure-Go SQLite, no CGO
- optional `llamacpp` build tag for plugging in a direct llama.cpp Go binding

## Local model

The app ships with an offline fallback engine so it runs with or without a
model. The easiest way to get a real model is the in-app downloader:

```
press d anywhere in the app  →  arrow keys to pick a model  →  Enter to download
                                a to activate  →  chat will pick it up
```

Files are saved under the `--models` directory (per-user cache dir by default)
and the choice is persisted to `settings.json`, so the next launch reloads
the same model automatically.

You can also point at any GGUF manually:

```sh
bit-tracker --model /path/to/your-model.gguf
# or
export BIT_TRACKER_MODEL=/path/to/your-model.gguf
bit-tracker
```

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
./bit-tracker [--data DIR] [--exports DIR] [--models DIR] [--model PATH]
```

Defaults (all cross-platform via `os.UserConfigDir` / `os.UserCacheDir`):

- `--data`    → `%AppData%\bit-tracker` (Windows), `~/Library/Application Support/bit-tracker` (macOS), `$XDG_CONFIG_HOME/bit-tracker` or `~/.config/bit-tracker` (Linux)
- `--exports` → `~/bit-tracker-exports`
- `--models`  → per-user cache dir (where GGUFs downloaded in-app are kept)
- `--model`   → explicit override; otherwise the last model activated in the UI is reloaded

## Shortcuts

| Where       | Keys                                                                 |
|-------------|----------------------------------------------------------------------|
| Global      | `1` landing · `2` diary · `3` chat · `Tab` / `Shift+Tab` cycle tabs  |
|             | `s` settings · `e` export · `d` models · `?` help · `q` quit         |
| Landing     | `Enter` jump to diary                                                |
| Diary today | `Enter` / `i` edit · `v` switch to history · `Esc` save & leave      |
|             | `r` rewrite via AI · `+` / `-` mood · `y` +15 study · `p` +15 project · `o` +15 scroll · `x` toggle done |
| Diary list  | `↑/↓` move · `Enter` open · `/` search · `#tag` tag filter · `Esc` back |
| Entry       | `m` export Markdown · `H` export HTML · `Esc` back                   |
| Chat        | `i` type · `Enter` send · `r` rewrite · `f` reflect · `u` wake up · `Ctrl+L` clear |
| Settings    | `↑/↓` field · `←/→` change · `Enter` next · `Esc` close               |
| Models      | `↑/↓` select · `Enter` download · `a` activate · `x` cancel · `Esc` close |
| Setup       | `Enter` next · `Backspace` back · `Esc` skip · per-page letters cycle |

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
