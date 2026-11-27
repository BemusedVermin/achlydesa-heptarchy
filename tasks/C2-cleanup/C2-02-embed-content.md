---
id: C2-02
title: "One binary: embed the production content"
status: todo
milestone: C2
kind: feature
depends_on: [C2-01]
discovered_by: null
size: M
---

# C2-02 · One binary: embed the production content

## Goal

A built `heptarchy` or `harness` binary carries the production content inside it and runs from any folder with nothing next to it. A `--content` folder still overrides it for test content, a designer's working copy, or watching for changes.

## Context

- `cmd/heptarchy/main.go` and `cmd/harness/main.go`: `defaultContent = "content/data"`, a path relative to the working folder, so today the binaries only work from the repository root. The `--content` flag feeds `content.Load(dir)` and `content.Watch(dir, ...)`.
- `internal/content/load.go`: `loadBundle` runs `load.Instances` with `load.Config{Dir: abs}` over the files under `dir`. The production content imports the generated schemas (`*_go_gen.cue` in `internal/content`, `internal/core/api` and `internal/core/comp`) through the CUE module in `cue.mod`, at the repository root. That is CUE v0.17.1.
- C2-01's load-and-validate test.
- `docs/go-technical-design.md`, under "Content pipeline".

## Deliverables

- A Go file that embeds the production content and everything it imports (`cue.mod` and the generated schemas) as an `fs.FS`. `//go:embed` can't reach a parent folder, so it probably sits at the module root.
- `internal/content`: loading from an `fs.FS`.
- `cmd/heptarchy` and `cmd/harness` use the embedded content unless `--content` names a folder.
- The load-and-validate test loads the embedded copy, so what ships is what's checked.
- `docs/go-technical-design.md` ("Content pipeline") and `docs/content-guide.md` (how to play your edits).

## Spec

1. **Embedded by default** (firm). With no `--content`, both commands load the embedded production content. They read no file from disk.
2. **A folder overrides it** (firm). `--content <dir>` loads that folder instead, exactly as today: `--content content/test` plays the test content. Watching for changes needs a folder, so `play` watches only when `--content` is given.
3. **One loader** (suggestion). `content.LoadFS(fsys fs.FS, dir string) (*Registry, error)`, with `Load(dir)` becoming a thin wrapper over `os.DirFS`. Inside it, feed `cue/load` the files through `load.Config`'s overlay, or its file-system option if v0.17 has one. There should be one code path, not two loaders.
4. **The embed is the module root's job** (suggestion). Content and the core stay unaware of embedding (AGENTS.md rule 1): only `cmd/` and the validate test import the embedded file system.
5. **Determinism** (firm). The embedded and the on-disk copy of the same content give the same registry, so a seed plays the same game either way.

## Acceptance criteria

- [ ] Built and run from outside the repository, the harness plays the production content. In PowerShell:

  ```powershell
  go build -o $env:TEMP\harness.exe ./cmd/harness
  Push-Location $env:TEMP; .\harness.exe run --seeds 2 --steps 50; Pop-Location
  ```

- [ ] `heptarchy.exe`, built the same way, starts `play --seed 1` from `$env:TEMP`. Check it by hand, since `play` is interactive.
- [ ] `go tool task run -- play --seed 1 --content content/test` plays the test content.
- [ ] `harness replay` of a golden log still reproduces its hash.
- [ ] The load-and-validate test loads the embedded copy.
- [ ] The tech design's "Content pipeline" and the content guide are updated.
- [ ] `go tool task check-pr` passes.
- [ ] `go tool task check` passes.

## Out of scope

- A build tag that strips the embed from developer builds. `go run` rebuilds with the current content anyway.
- Release packaging and installers.

## Notes
