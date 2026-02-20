# go-wimgapi

Pure-Go (no cgo) Windows bindings for `wimgapi.dll`.

## Status
- Initial implementation (v0 scope).
- Platform: Windows only.
- cgo: not used; CI/build should run with `CGO_ENABLED=0`.

## Implemented APIs
- Open/close WIM files
- Get image count and image list metadata
- Load image by index
- Apply image to target directory
- Register progress callback for apply
- Register progress callback for capture
- Decode callback messages with `ProgressDecoder`

## CLI
`cmd/wimctl` provides:
- `wimctl list <path-to-wim>`
- `wimctl apply <path-to-wim> <index> <target-dir>`
- `wimctl capture <source-dir> <path-to-wim>`

## Quick Start
```powershell
go run ./cmd/wimctl list C:\images\install.wim
```

## Roundtrip Test Program
```powershell
go run ./examples/roundtrip-test
```

Integration test wrapper (optional):
```powershell
$env:WIMGAPI_INTEGRATION='1'
go test -run TestRoundtripProgram ./...
```

## Notes
- Some operations may require administrator privileges.
- Runtime behavior can vary by Windows/ADK version.
- If you hit temp-folder errors (for example code `1632`), set a writable WIM temp path (the roundtrip example now does this automatically).
