# Golang Binding Plan for Windows WIMGAPI

## 1. Goals and Scope
- Goal: build a maintainable, testable Go package that wraps `wimgapi.dll` with type-safe, Go-style APIs.
- Target platform: Windows only (`windows/amd64` first, then `windows/arm64`).
- MVP capabilities:
  - Open and close WIM files.
  - List images and read basic metadata (index, name, description, flags).
  - Load an image and apply it to a target directory.
  - Provide basic progress callback and error mapping.

## 2. Out of Scope for v0
- Full coverage of all WIMGAPI functions.
- Cross-platform emulation.
- GUI tools.
- Full DISM integration (extension points only).

## 3. Binding Strategy
- Primary approach: `golang.org/x/sys/windows` with dynamic loading:
  - `windows.NewLazySystemDLL("wimgapi.dll")`
  - `NewProc(...)`
- Hard requirement: no cgo in any package, build tag, or test path.
- Principle: pure Go (`syscall`/`x/sys/windows`) only, to keep builds simple and portable across Windows runners.

## 4. Package Layout
Suggested module path: `github.com/<org>/go-wimgapi`

```text
/wimgapi
  errors.go           // Win32/HRESULT to Go error
  types.go            // constants, enums, structs, handle wrappers
  dll.go              // lazy load dll + proc binding + version checks
  file.go             // WIMCreateFile/WIMCloseHandle wrappers
  image.go            // WIMLoadImage/WIMGetImageInformation wrappers
  apply.go            // WIMApplyImage wrapper + options
  callback.go         // callback registration and event dispatch
  mount.go            // mount/unmount (phase 2)
  capture.go          // capture/create (phase 2)
/internal/syscallx
  utf16.go
  handle.go
  unsafeconv.go
/cmd/wimctl
  main.go             // minimal CLI for integration validation
```

## 5. Public API Shape (Go-first)
- `Open(path string, opts OpenOptions) (*File, error)`
  - wraps `WIMCreateFile`.
- `(*File) Close() error`
- `(*File) Images() ([]ImageInfo, error)`
  - combines count + load + metadata parsing.
- `(*File) LoadImage(index int) (*Image, error)`
- `(*Image) Apply(target string, opts ApplyOptions) error`
  - wraps `WIMApplyImage`.
- Callback model:
  - `type ProgressFunc func(evt ProgressEvent) (cancel bool)`
  - translate WIMGAPI messages to typed Go events.

## 6. Error and Resource Model
- All exported resource types must have explicit `Close()`.
- Optional finalizer can be used only for leak warning, not normal cleanup.
- Unified error type:
  - `type Error struct { Op string; Code uint32; Msg string }`
- Include Win32 message text via `FormatMessage`.
- Add helper predicates for common failures:
  - access denied
  - invalid path
  - image index out of range

## 7. Callback and Concurrency Rules
- Keep native callback minimal and non-blocking.
- Marshal callback events into Go through internal channel.
- Document thread-safety clearly:
  - `File` and `Image` are not safe for concurrent use unless stated.
- Support cancellation by callback return value and context integration where possible.
- Implement callback bridging with `syscall.NewCallback` patterns only (no cgo trampolines).

## 8. Build and Compatibility
- Guard all implementation files with `//go:build windows`.
- Detect missing `wimgapi.dll` at runtime and return actionable errors.
- Document environment dependencies:
  - ADK / OS version differences
  - admin rights for some mount/apply scenarios
- Enforce no-cgo in CI:
  - run builds/tests with `CGO_ENABLED=0`
  - add a lint/check step that fails on any `import "C"` usage.

## 9. Test Plan
- Unit tests:
  - option translation
  - UTF-16 conversion
  - error conversion
  - input validation
- Integration tests (Windows runner):
  - open/list/load/apply using small fixture WIM.
- Optional E2E:
  - run `cmd/wimctl` against temp target directory and verify output tree.
- CI proposal:
  - GitHub Actions `windows-latest`
  - separate `unit` and `integration` jobs.

## 10. Milestones
- M0 (1-2 days): scaffolding
  - module init, layout, dll loader, base errors.
- M1 (3-5 days): file and image read path
  - Open/Close/Images/LoadImage + tests.
- M2 (3-5 days): apply support
  - Apply + callback + CLI demo.
- M3 (4-7 days, optional): mount/capture
  - mount, commit, unmount; docs for privilege constraints.
- M4 (2-3 days): release prep
  - README, examples, CI polish, tag `v0.1.0`.

## 11. Risks and Mitigations
- Risk: behavior differences across Windows versions.
  - Mitigation: integration matrix (Win10/Win11/Server).
- Risk: callback boundary instability.
  - Mitigation: minimal callback body, strict memory ownership rules, and stress tests for callback lifecycle.
- Risk: flaky high-privilege tests.
  - Mitigation: isolate privileged tests and do not block core unit tests.

## 12. Documentation and Release Rules
- README must include:
  - supported scope
  - install and quick start
  - required privileges
  - known limitations
- Add `examples/`:
  - `list-images`
  - `apply-image`
- Versioning:
  - iterate in `v0.x`, maintain clear changelog for API changes.

## 13. First Iteration Backlog
1. Initialize Go module and directory skeleton.
2. Implement `dll.go` with lazy proc binding.
3. Implement callback plumbing using `syscall.NewCallback` and validate no-cgo build.
4. Implement `errors.go` with Win32 message conversion.
5. Implement `file.go` with `Open` and `Close`.
6. Implement `image.go` with minimal listing and load flow.
7. Add `cmd/wimctl list <wim>` for smoke validation.
8. Add unit tests and baseline Windows CI workflow.
