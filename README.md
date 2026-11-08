# nozzle.go

> This codebase is currently in its AI-slob prototyping phase: the code runs on momentum, vibes, and plausible intent.
> Proper debugging will be introduced once demand graduates from hypothetical to measurable.

Go bindings for [nozzle](https://github.com/nozzle-io/nozzle) — cross-platform GPU texture sharing between local processes.

## Disclaimer / Notice

This library is currently a work in progress and contains many incomplete features and unverified implementations.
Although it may appear usable at first glance, it may not function correctly.

## Build Requirements

- Go 1.21+
- C++17 compiler (clang / MSVC)
- macOS 12+, Windows 10+, or Linux

The nozzle C library is built from source via a git submodule. A `Makefile` compiles the static library before `go build` links against it.

## Build

```bash
make
go build ./nozzle/
```

### Run Tests

```bash
make
go test ./nozzle/ -v
```

## Usage

### Sender

```go
package main

import (
    "fmt"
    "github.com/nozzle-io/nozzle.go/nozzle"
)

func main() {
    sender, err := nozzle.NewSender(nozzle.SenderDesc{
        Name:            "go-sender",
        ApplicationName: "MyApp",
        RingBufferSize:  3,
    })
    if err != nil {
        panic(err)
    }
    defer sender.Close()

    frame, err := sender.AcquireWritableFrame(1920, 1080, nozzle.FormatRGBA8UNorm)
    if err != nil {
        panic(err)
    }

    pixels, err := frame.LockWritablePixels(nozzle.OriginTopLeft)
    if err != nil {
        panic(err)
    }
    pixelsLocked := true
    defer func() {
        if pixelsLocked {
            pixels.Unmap()
        }
    }()

    for y := 0; y < pixels.Height; y++ {
        row, _ := pixels.Row(y)
        for i := range row {
            row[i] = 0xFF
        }
    }

    if err := pixels.UnmapChecked(); err != nil {
        panic(err)
    }
    pixelsLocked = false

    if err := sender.CommitFrame(frame); err != nil {
        panic(err)
    }
    _ = fmt.Sprintf("done")
}
```

### Receiver

```go
receiver, err := nozzle.NewReceiver(nozzle.ReceiverDesc{
    Name:            "go-sender",
    ApplicationName: "MyViewer",
    ReceiveMode:     nozzle.ReceiveLatestOnly,
})
if err != nil {
    panic(err)
}
defer receiver.Close()

frame, err := receiver.AcquireFrame(5000)
if err != nil {
    panic(err)
}
defer frame.Release()

info, err := frame.Info()
if err != nil {
    panic(err)
}
fmt.Printf("%dx%d frame #%d\n", info.Width, info.Height, info.FrameIndex)
```

### Metadata Snapshots

`sender.Info()` returns owned sender metadata. `receiver.ConnectionSnapshot()`
returns `ConnectionSnapshot{Connected, Info}`, copying state and metadata
together under the native receiver lock. Inline C strings are copied within
their array bounds; Go strings preserve the original bytes, including invalid
UTF-8, rather than applying lossy decoding. Results survive later calls and
handle closure.

`Connected` is the receiver's last-observed state, not fresh liveness or a
reconnect attempt. After acquisition observes a disconnect, it is false while
`Info` retains last-known identity. `ConnectedInfo()` preserves its existing
public fields and returns that info even while disconnected. `IsConnected()`
reads the snapshot flag and returns false on retrieval error. A sender can exit
after the receiver last observed it; acquisition errors observe closure.

This does not change handle lifetime or concurrency rules: `Close` must not race
any other use of the same sender/receiver, and a snapshot call does not retain
the native handle against concurrent closure.

Pure string tests cover signed/unsigned bytes, NUL boundaries, invalid UTF-8,
and copying. Required macOS Metal tests cover same-process closure and retained
metadata, not cross-process transfer, GPU pixels, native wide-counter values,
or Windows/Linux runtime behavior.

### Discovery

```go
count, err := nozzle.EnumerateSenders()
if err != nil {
    panic(err)
}
fmt.Printf("found %d senders\n", count)
```

### GPU Check

```go
if nozzle.IsGPUAvailable() {
    fmt.Println("GPU available")
}
```

## Error Handling

All fallible operations return `error`. Nozzle errors implement the `error` interface with human-readable messages:

```go
frame, err := sender.AcquireWritableFrame(0, 0, nozzle.FormatUnknown)
if err != nil {
    switch err {
    case nozzle.ErrorInvalidArgument:
        // handle bad args
    case nozzle.ErrorUnsupportedFormat:
        // handle bad format
    default:
        // handle other errors
    }
}
```

## Pixel Mapping Semantics

Read-only `LockPixels()` returns a Go-owned copy. Core performs the
lock/copy/unlock sequence inside one C API call, so the Go binding does not need
to pin the goroutine to an OS thread for read-only copies. No unmap is required
for read-only copies, and `Unmap()` / `UnmapChecked()` are compatibility no-ops.

Writable `LockWritablePixels()` returns a native mapped memory view, not a copy.
The current goroutine stays pinned to its OS thread until `UnmapChecked()` or
`Unmap()` is called. Do not pass writable mappings to another goroutine. New
code should use `UnmapChecked()` and only call `CommitFrame()` after it succeeds;
legacy `Unmap()` discards unlock errors.

## Texture Formats

| Format | Bytes/Pixel |
|--------|-------------|
| `FormatR8UNorm` | 1 |
| `FormatRG8UNorm` | 2 |
| `FormatRGBA8UNorm` / `FormatBGRA8UNorm` | 4 |
| `FormatRGBA8SRGB` / `FormatBGRA8SRGB` | 4 |
| `FormatR16UNorm` | 2 |
| `FormatRG16UNorm` | 4 |
| `FormatRGBA16UNorm` | 8 |
| `FormatR16Float` | 2 |
| `FormatRG16Float` | 4 |
| `FormatRGBA16Float` | 8 |
| `FormatR32Float` | 4 |
| `FormatRG32Float` | 8 |
| `FormatRGBA32Float` | 16 |
| `FormatR32Uint` | 4 |
| `FormatRGBA32Uint` | 16 |
| `FormatDepth32Float` | 4 |

## Platform Notes

- **macOS**: Links Metal, IOSurface, Foundation, Accelerate, OpenGL frameworks automatically via cgo
- **Windows**: Links d3d11, dxgi, opengl32, bcrypt automatically via cgo
- **Linux**: Links drm, gbm, EGL, GL automatically via cgo

## License

MIT
