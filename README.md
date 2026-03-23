# agentic-news

Task 8 adds and verifies the built-in render theme matrix plus the local stylesheet build step needed to preview those themes consistently.

## Supported Themes

Built-in theme IDs:

- `editorial-ai`
- `ai-product-magazine`
- `youth-signal`
- `soft-focus`

## Local Style Build

Install frontend dependencies once:

```bash
npm install
```

Rebuild the local stylesheet bundle after editing `web/tailwind/input.css` or any theme-specific template classes:

```bash
npm run build:styles
```

The renderer copies `web/static/styles.css` into each dated edition’s `assets/` directory, so stale local CSS will propagate into generated theme outputs.

## Theme Run Examples

Generate a sample edition with the `editorial-ai` theme:

```bash
go run ./cmd/daily-builder sample 2026-03-20 --theme editorial-ai
```

Generate a sample edition with the `ai-product-magazine` theme:

```bash
go run ./cmd/daily-builder sample 2026-03-19 --theme ai-product-magazine
```

Other supported theme IDs can be passed the same way:

```bash
go run ./cmd/daily-builder sample 2026-03-19 --theme youth-signal
go run ./cmd/daily-builder sample 2026-03-19 --theme soft-focus
```

## Hugo Render Bridge

The Go pipeline now writes a theme-neutral edition package alongside the themed output. For a run dated `2026-03-20`, the package lives under:

```text
output/_packages/2026/03/20
```

The `sample` command prints both the themed output root and the package root:

```bash
go run ./cmd/daily-builder sample 2026-03-20 --theme editorial-ai
```

If you want to render that package through Hugo, run:

```bash
go run ./cmd/daily-builder render-hugo --date 2026-03-20 --theme editorial-ai
```

`hugo` is an optional local prerequisite for the second step only. The Go renderer and package writer still work without the Hugo binary installed.

## Image Archival

When a rendered pick includes `CoverImageLocal`, the renderer prefers that archived local path instead of the original remote image URL.

That is why original image assets appear under:

```text
/THEME/YYYY/MM/DD/assets/images/...
```

This keeps each dated edition self-contained, preserves historical rendering, and avoids later breakage if the upstream image host changes or disappears.
