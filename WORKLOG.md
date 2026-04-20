# WORKLOG

## Objective
Windows support + in-app GGUF model download + landing page with ASCII pet
(`bit`) + three-tab layout (landing/diary/chat) + settings overlay + themable
colours + character customisation.

## Current Status
All feature work complete. `go build`, `go vet`, `go test` green on
linux/amd64, windows/amd64, darwin/arm64.

## Files Changed
- cmd/bit-tracker/main.go (settings.Store, model dir, flag reshuffle)
- internal/settings/{settings.go,settings_test.go}
- internal/pet/{pet.go,quote.go,pet_test.go}
- internal/ai/catalog.go
- internal/download/{download.go,download_test.go}
- internal/ui/theme.go (palettes, Popup, Overlay, PetFrame)
- internal/app/app.go (3 tabs, overlays, settings, download wiring)
- internal/app/view_landing.go · view_diary.go · view_chat.go · overlays.go
- README.md

## Decisions
- Default engine stays `ai.stub`; real inference gated behind `llamacpp` tag.
- Downloaded GGUFs live in the per-user cache dir (`--models` flag overrides).
- `settings.json` tracks pet config, theme, and active model filename.
- `s/e/d/?` are global overlay hotkeys; study bump is `y/Y` to free `s`.

## Next Steps
- Commit + push to `claude/build-bit-tracker-fXO93`.

## Blockers
None.

## Resume Prompt
All feature work done; only commit + push remain. Branch: `claude/build-bit-tracker-fXO93`.
