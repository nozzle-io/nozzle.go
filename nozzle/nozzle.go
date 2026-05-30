// Package nozzle provides Go bindings for the nozzle GPU texture sharing library.
//
// nozzle enables local inter-process GPU texture sharing via Metal/IOSurface (macOS),
// D3D11 (Windows), and DMA-BUF (Linux).
//
// Before using this package, build the nozzle static library:
//
//	make
package nozzle

/*
#cgo CFLAGS: -I${SRCDIR}/../deps/nozzle/include
#cgo LDFLAGS: -L${SRCDIR}/../.build -lnozzle -lstdc++

#cgo darwin LDFLAGS: -framework Metal -framework IOSurface -framework Foundation -framework Accelerate -framework OpenGL -lobjc
#cgo windows LDFLAGS: -ld3d11 -ldxgi -lopengl32 -lbcrypt
#cgo linux LDFLAGS: -ldrm -lgbm -lEGL -lGL

#include <stdlib.h>
#include <nozzle/nozzle_c.h>
*/
import "C"
import (
	"errors"
	"fmt"
	"runtime"
	"unsafe"
)

// ErrorCode represents a nozzle error.
type ErrorCode int

const (
	ErrorUnknown            ErrorCode = C.NOZZLE_ERROR_UNKNOWN
	ErrorInvalidArgument    ErrorCode = C.NOZZLE_ERROR_INVALID_ARGUMENT
	ErrorUnsupportedBackend ErrorCode = C.NOZZLE_ERROR_UNSUPPORTED_BACKEND
	ErrorUnsupportedFormat  ErrorCode = C.NOZZLE_ERROR_UNSUPPORTED_FORMAT
	ErrorDeviceMismatch     ErrorCode = C.NOZZLE_ERROR_DEVICE_MISMATCH
	ErrorResourceCreation   ErrorCode = C.NOZZLE_ERROR_RESOURCE_CREATION_FAILED
	ErrorSharedHandle       ErrorCode = C.NOZZLE_ERROR_SHARED_HANDLE_FAILED
	ErrorSenderNotFound     ErrorCode = C.NOZZLE_ERROR_SENDER_NOT_FOUND
	ErrorSenderClosed       ErrorCode = C.NOZZLE_ERROR_SENDER_CLOSED
	ErrorTimeout            ErrorCode = C.NOZZLE_ERROR_TIMEOUT
	ErrorBackend            ErrorCode = C.NOZZLE_ERROR_BACKEND_ERROR
	ErrorCommandFailed      ErrorCode = C.NOZZLE_ERROR_COMMAND_FAILED
)

func (e ErrorCode) Error() string {
	switch e {
	case ErrorUnknown:
		return "unknown error"
	case ErrorInvalidArgument:
		return "invalid argument"
	case ErrorUnsupportedBackend:
		return "unsupported backend"
	case ErrorUnsupportedFormat:
		return "unsupported format"
	case ErrorDeviceMismatch:
		return "device mismatch"
	case ErrorResourceCreation:
		return "resource creation failed"
	case ErrorSharedHandle:
		return "shared handle failed"
	case ErrorSenderNotFound:
		return "sender not found"
	case ErrorSenderClosed:
		return "sender closed"
	case ErrorTimeout:
		return "timeout"
	case ErrorBackend:
		return "backend error"
	case ErrorCommandFailed:
		return "command execution failed"
	default:
		return fmt.Sprintf("nozzle error (%d)", int(e))
	}
}

func checkCode(code C.NozzleErrorCode) error {
	if code == C.NOZZLE_OK {
		return nil
	}
	return ErrorCode(code)
}

// BackendType represents a GPU backend.
type BackendType C.NozzleBackendType

const (
	BackendUnknown BackendType = C.NOZZLE_BACKEND_UNKNOWN
	BackendD3D11   BackendType = C.NOZZLE_BACKEND_D3D11
	BackendMetal   BackendType = C.NOZZLE_BACKEND_METAL
	BackendOpenGL  BackendType = C.NOZZLE_BACKEND_OPENGL
	BackendDMABuf  BackendType = C.NOZZLE_BACKEND_DMA_BUF
)

func (b BackendType) String() string {
	switch b {
	case BackendD3D11:
		return "d3d11"
	case BackendMetal:
		return "metal"
	case BackendOpenGL:
		return "opengl"
	case BackendDMABuf:
		return "dma_buf"
	default:
		return "unknown"
	}
}

// TextureFormat represents a pixel format.
type TextureFormat C.NozzleTextureFormat

const (
	FormatUnknown      TextureFormat = C.NOZZLE_FORMAT_UNKNOWN
	FormatR8UNorm      TextureFormat = C.NOZZLE_FORMAT_R8_UNORM
	FormatRG8UNorm     TextureFormat = C.NOZZLE_FORMAT_RG8_UNORM
	FormatRGBA8UNorm   TextureFormat = C.NOZZLE_FORMAT_RGBA8_UNORM
	FormatBGRA8UNorm   TextureFormat = C.NOZZLE_FORMAT_BGRA8_UNORM
	FormatRGBA8SRGB    TextureFormat = C.NOZZLE_FORMAT_RGBA8_SRGB
	FormatBGRA8SRGB    TextureFormat = C.NOZZLE_FORMAT_BGRA8_SRGB
	FormatR16UNorm     TextureFormat = C.NOZZLE_FORMAT_R16_UNORM
	FormatRG16UNorm    TextureFormat = C.NOZZLE_FORMAT_RG16_UNORM
	FormatRGBA16UNorm  TextureFormat = C.NOZZLE_FORMAT_RGBA16_UNORM
	FormatR16Float     TextureFormat = C.NOZZLE_FORMAT_R16_FLOAT
	FormatRG16Float    TextureFormat = C.NOZZLE_FORMAT_RG16_FLOAT
	FormatRGBA16Float  TextureFormat = C.NOZZLE_FORMAT_RGBA16_FLOAT
	FormatR32Float     TextureFormat = C.NOZZLE_FORMAT_R32_FLOAT
	FormatRG32Float    TextureFormat = C.NOZZLE_FORMAT_RG32_FLOAT
	FormatRGBA32Float  TextureFormat = C.NOZZLE_FORMAT_RGBA32_FLOAT
	FormatR32Uint      TextureFormat = C.NOZZLE_FORMAT_R32_UINT
	FormatRGBA32Uint   TextureFormat = C.NOZZLE_FORMAT_RGBA32_UINT
	FormatDepth32Float TextureFormat = C.NOZZLE_FORMAT_DEPTH32_FLOAT
)

// BytesPerPixel returns the number of bytes per pixel, or 0 for unknown formats.
func (f TextureFormat) BytesPerPixel() int {
	switch f {
	case FormatR8UNorm:
		return 1
	case FormatRG8UNorm, FormatR16UNorm, FormatR16Float:
		return 2
	case FormatRGBA8UNorm, FormatBGRA8UNorm, FormatRGBA8SRGB, FormatBGRA8SRGB,
		FormatRG16UNorm, FormatRG16Float, FormatR32Float, FormatR32Uint, FormatDepth32Float:
		return 4
	case FormatRGBA16UNorm, FormatRGBA16Float, FormatRG32Float:
		return 8
	case FormatRGBA32Float, FormatRGBA32Uint:
		return 16
	default:
		return 0
	}
}

// ReceiveMode controls how frames are acquired.
type ReceiveMode C.NozzleReceiveMode

const (
	ReceiveLatestOnly           ReceiveMode = C.NOZZLE_RECEIVE_LATEST_ONLY
	ReceiveSequentialBestEffort ReceiveMode = C.NOZZLE_RECEIVE_SEQUENTIAL_BEST_EFFORT
)

// FrameStatus indicates the result of frame acquisition.
type FrameStatus C.NozzleFrameStatus

const (
	FrameNew          FrameStatus = C.NOZZLE_FRAME_NEW
	FrameNoNew        FrameStatus = C.NOZZLE_FRAME_NO_NEW
	FrameDropped      FrameStatus = C.NOZZLE_FRAME_DROPPED
	FrameSenderClosed FrameStatus = C.NOZZLE_FRAME_SENDER_CLOSED
	FrameError        FrameStatus = C.NOZZLE_FRAME_ERROR
)

// TextureOrigin controls pixel data row ordering.
type TextureOrigin C.NozzleTextureOrigin

const (
	OriginTopLeft    TextureOrigin = C.NOZZLE_ORIGIN_TOP_LEFT
	OriginBottomLeft TextureOrigin = C.NOZZLE_ORIGIN_BOTTOM_LEFT
)

// FormatSource indicates how a texture format was determined.
type FormatSource C.NozzleFormatSource

const (
	FormatSourceUnknown        FormatSource = C.NOZZLE_FORMAT_SOURCE_UNKNOWN
	FormatSourceRequested      FormatSource = C.NOZZLE_FORMAT_SOURCE_REQUESTED
	FormatSourceCallerHint     FormatSource = C.NOZZLE_FORMAT_SOURCE_CALLER_HINT
	FormatSourceNativeObserved FormatSource = C.NOZZLE_FORMAT_SOURCE_NATIVE_OBSERVED
)

// NativeFormatKind identifies the type of a native GPU format value.
type NativeFormatKind C.NozzleNativeFormatKind

const (
	NativeKindUnknown          NativeFormatKind = C.NOZZLE_NATIVE_KIND_UNKNOWN
	NativeKindMTLPixelFormat   NativeFormatKind = C.NOZZLE_NATIVE_KIND_MTL_PIXEL_FORMAT
	NativeKindDXGIFormat       NativeFormatKind = C.NOZZLE_NATIVE_KIND_DXGI_FORMAT
	NativeKindDRMFourCC        NativeFormatKind = C.NOZZLE_NATIVE_KIND_DRM_FOURCC
	NativeKindGLInternalFormat NativeFormatKind = C.NOZZLE_NATIVE_KIND_GL_INTERNAL_FORMAT
)

// SenderDesc configures a new sender.
type SenderDesc struct {
	Name                string
	ApplicationName     string
	RingBufferSize      uint32
	AllowFormatFallback bool
}

// ReceiverDesc configures a new receiver.
type ReceiverDesc struct {
	Name            string
	ApplicationName string
	ReceiveMode     ReceiveMode
}

// SenderInfo contains sender metadata.
type SenderInfo struct {
	Name            string
	ApplicationName string
	ID              string
	Backend         BackendType
}

// ConnectedSenderInfo contains detailed info about the connected sender.
type ConnectedSenderInfo struct {
	Name            string
	ApplicationName string
	ID              string
	Backend         BackendType
	Width           uint32
	Height          uint32
	Format          TextureFormat
	SemanticFormat  TextureFormat
	EstimatedFPS    float64
	FrameCounter    uint64
	LastUpdateTime  uint64
}

// FrameInfo contains frame metadata.
type FrameInfo struct {
	FrameIndex        uint64
	Timestamp         uint64
	Width             uint32
	Height            uint32
	Format            TextureFormat
	SemanticFormat    TextureFormat
	DroppedFrameCount uint32
}

// MappedPixels provides access to frame pixel data.
type MappedPixels struct {
	Data              []byte
	RowStrideBytes    int
	Width             int
	Height            int
	Format            TextureFormat
	Origin            TextureOrigin
	frame             *C.NozzleFrame
	mapping           *C.NozzlePixelMapping
	owner             *WritableFrame
	writable          bool
	locked            bool
	osthreadLocked    bool
	absRowStrideBytes int
}

// Row returns a byte slice for the given row, or an error if out of bounds.
func (m *MappedPixels) Row(y int) ([]byte, error) {
	if y < 0 || y >= m.Height {
		return nil, fmt.Errorf("row %d out of bounds (height %d)", y, m.Height)
	}
	rowStride := m.RowStrideBytes
	absStride := m.absRowStrideBytes
	if absStride == 0 {
		absStride = absInt(rowStride)
	}
	if absStride == 0 {
		return nil, fmt.Errorf("invalid row stride %d", rowStride)
	}
	start := y * absStride
	if rowStride < 0 {
		start = (m.Height - 1 - y) * absStride
	}
	if start < 0 || start+absStride > len(m.Data) {
		return nil, fmt.Errorf("row %d outside mapped pixel data", y)
	}
	return m.Data[start : start+absStride], nil
}

// Unmap releases the pixel mapping and discards checked unlock errors.
//
// Read-only mappings returned by LockPixels are Go-owned copies; Unmap is a
// no-op for them. Writable mappings are native views owned by an opaque core
// mapping handle.
func (m *MappedPixels) Unmap() {
	_ = m.UnmapChecked()
}

// UnmapChecked releases a writable pixel mapping and reports backend unlock
// failures. It invalidates the MappedPixels even when unlock reports an error;
// retrying after an error is unsupported.
func (m *MappedPixels) UnmapChecked() error {
	if m == nil {
		return ErrorInvalidArgument
	}
	if !m.locked {
		if m.writable {
			return ErrorInvalidArgument
		}
		return nil
	}

	var err error
	if m.writable {
		if m.mapping != nil {
			err = checkCode(C.nozzle_pixel_mapping_unlock_checked(&m.mapping))
		} else {
			err = checkCode(C.nozzle_frame_unlock_writable_pixels_checked(m.frame))
		}
	} else {
		if m.mapping != nil {
			C.nozzle_pixel_mapping_unlock(&m.mapping)
		} else {
			C.nozzle_frame_unlock_pixels(m.frame)
		}
	}

	osthreadLocked := m.osthreadLocked
	if m.owner != nil && m.owner.activeMapping == m {
		m.owner.activeMapping = nil
	}
	m.owner = nil
	m.Data = nil
	m.frame = nil
	m.mapping = nil
	m.locked = false
	m.osthreadLocked = false
	if osthreadLocked {
		runtime.UnlockOSThread()
	}
	return err
}

// Sender sends GPU textures to a named receiver.
type Sender struct {
	raw *C.NozzleSender
}

// NewSender creates a new sender.
func NewSender(desc SenderDesc) (*Sender, error) {
	cName := C.CString(desc.Name)
	cAppName := C.CString(desc.ApplicationName)
	defer C.free(unsafe.Pointer(cName))
	defer C.free(unsafe.Pointer(cAppName))

	cDesc := C.NozzleSenderDesc{
		name:                 cName,
		application_name:     cAppName,
		ring_buffer_size:     C.uint32_t(desc.RingBufferSize),
		fallback_flags:       C.uint32_t(3),
		fallback_flags_valid: 1,
	}

	var raw *C.NozzleSender
	if err := checkCode(C.nozzle_sender_create(&cDesc, &raw)); err != nil {
		return nil, err
	}
	return &Sender{raw: raw}, nil
}

// Close destroys the sender.
func (s *Sender) Close() {
	if s.raw != nil {
		C.nozzle_sender_destroy(s.raw)
		s.raw = nil
	}
}

// AcquireWritableFrame gets a frame to write pixel data into.
func (s *Sender) AcquireWritableFrame(width, height uint32, format TextureFormat) (*WritableFrame, error) {
	var raw *C.NozzleFrame
	if err := checkCode(C.nozzle_sender_acquire_writable_frame(s.raw, C.uint32_t(width), C.uint32_t(height), C.NozzleTextureFormat(format), &raw)); err != nil {
		return nil, err
	}
	return &WritableFrame{raw: raw}, nil
}

// CommitFrame publishes a writable frame.
func (s *Sender) CommitFrame(f *WritableFrame) error {
	return checkCode(C.nozzle_sender_commit_frame(s.raw, f.raw))
}

// PublishGLTexture publishes an OpenGL texture.
func (s *Sender) PublishGLTexture(glName, glTarget, width, height uint32, format TextureFormat) error {
	return checkCode(C.nozzle_sender_publish_gl_texture(s.raw, C.uint32_t(glName), C.uint32_t(glTarget), C.uint32_t(width), C.uint32_t(height), C.NozzleTextureFormat(format)))
}

// Info returns sender metadata.
func (s *Sender) Info() (SenderInfo, error) {
	var raw C.NozzleSenderInfo
	if err := checkCode(C.nozzle_sender_get_info(s.raw, &raw)); err != nil {
		return SenderInfo{}, err
	}
	return SenderInfo{
		Name:            C.GoString(raw.name),
		ApplicationName: C.GoString(raw.application_name),
		ID:              C.GoString(raw.id),
		Backend:         BackendType(raw.backend),
	}, nil
}

// Receiver receives GPU textures from a named sender.
type Receiver struct {
	raw *C.NozzleReceiver
}

// NewReceiver creates a new receiver.
func NewReceiver(desc ReceiverDesc) (*Receiver, error) {
	cName := C.CString(desc.Name)
	cAppName := C.CString(desc.ApplicationName)
	defer C.free(unsafe.Pointer(cName))
	defer C.free(unsafe.Pointer(cAppName))

	cDesc := C.NozzleReceiverDesc{
		name:             cName,
		application_name: cAppName,
		receive_mode:     C.NozzleReceiveMode(desc.ReceiveMode),
	}

	var raw *C.NozzleReceiver
	if err := checkCode(C.nozzle_receiver_create(&cDesc, &raw)); err != nil {
		return nil, err
	}
	return &Receiver{raw: raw}, nil
}

// Close destroys the receiver.
func (r *Receiver) Close() {
	if r.raw != nil {
		C.nozzle_receiver_destroy(r.raw)
		r.raw = nil
	}
}

// AcquireFrame waits for and returns the next available frame.
func (r *Receiver) AcquireFrame(timeoutMs uint64) (*Frame, error) {
	cDesc := C.NozzleAcquireDesc{timeout_ms: C.uint64_t(timeoutMs)}
	var raw *C.NozzleFrame
	if err := checkCode(C.nozzle_receiver_acquire_frame(r.raw, &cDesc, &raw)); err != nil {
		return nil, err
	}
	return &Frame{raw: raw}, nil
}

// ConnectedInfo returns info about the connected sender.
func (r *Receiver) ConnectedInfo() (ConnectedSenderInfo, error) {
	var raw C.NozzleConnectedSenderInfo
	if err := checkCode(C.nozzle_receiver_get_connected_info(r.raw, &raw)); err != nil {
		return ConnectedSenderInfo{}, err
	}
	return ConnectedSenderInfo{
		Name:            C.GoString(raw.name),
		ApplicationName: C.GoString(raw.application_name),
		ID:              C.GoString(raw.id),
		Backend:         BackendType(raw.backend),
		Width:           uint32(raw.width),
		Height:          uint32(raw.height),
		Format:          TextureFormat(raw.format),
		SemanticFormat:  TextureFormat(raw.semantic_format),
		EstimatedFPS:    float64(raw.estimated_fps),
		FrameCounter:    uint64(raw.frame_counter),
		LastUpdateTime:  uint64(raw.last_update_time_ns),
	}, nil
}

// IsConnected returns whether a sender is connected.
func (r *Receiver) IsConnected() bool {
	_, err := r.ConnectedInfo()
	return err == nil
}

// Frame represents a received GPU texture.
type Frame struct {
	raw *C.NozzleFrame
}

// Release frees the frame.
func (f *Frame) Release() {
	if f.raw != nil {
		C.nozzle_frame_release(f.raw)
		f.raw = nil
	}
}

// Info returns frame metadata.
func (f *Frame) Info() (FrameInfo, error) {
	var raw C.NozzleFrameInfo
	if err := checkCode(C.nozzle_frame_get_info(f.raw, &raw)); err != nil {
		return FrameInfo{}, err
	}
	return FrameInfo{
		FrameIndex:        uint64(raw.frame_index),
		Timestamp:         uint64(raw.timestamp_ns),
		Width:             uint32(raw.width),
		Height:            uint32(raw.height),
		Format:            TextureFormat(raw.format),
		SemanticFormat:    TextureFormat(raw.semantic_format),
		DroppedFrameCount: uint32(raw.dropped_frame_count),
	}, nil
}

// LockPixels copies the frame pixel data for reading.
//
// The core C API performs lock/copy/unlock inside a single call, so read-only
// callers do not need to pin the goroutine to an OS thread. The returned
// MappedPixels owns a Go copy; Unmap/UnmapChecked are no-ops for read-only
// mappings.
func (f *Frame) LockPixels(origin TextureOrigin) (*MappedPixels, error) {
	info, err := f.Info()
	if err != nil {
		return nil, err
	}
	resolved, err := f.ResolvedFormat()
	if err != nil {
		return nil, err
	}
	rowStride, total, err := compactPixelCopySize(
		int(info.Width), int(info.Height), int(resolved.BytesPerPixel))
	if err != nil {
		return nil, err
	}

	data := make([]byte, total)
	var mapped C.NozzleMappedPixels
	if err := checkCode(C.nozzle_frame_copy_pixels_with_origin(
		f.raw,
		C.NozzleTextureOrigin(origin),
		unsafe.Pointer(&data[0]),
		C.uint64_t(len(data)),
		&mapped,
	)); err != nil {
		return nil, err
	}

	stride := int(mapped.row_stride_bytes)
	if stride != rowStride {
		rowStride = stride
	}
	absStride := absInt(rowStride)
	return &MappedPixels{
		Data:              data,
		RowStrideBytes:    rowStride,
		Width:             int(mapped.width),
		Height:            int(mapped.height),
		Format:            TextureFormat(mapped.format),
		Origin:            TextureOrigin(mapped.origin),
		frame:             nil,
		writable:          false,
		locked:            false,
		osthreadLocked:    false,
		absRowStrideBytes: absStride,
	}, nil
}

// LockWritablePixels maps the frame pixel data for writing.
func (f *Frame) LockWritablePixels(origin TextureOrigin) (*MappedPixels, error) {
	var mapped C.NozzleMappedPixels
	var mapping *C.NozzlePixelMapping
	if err := checkCode(C.nozzle_frame_lock_writable_pixels_mapping_with_origin(
		f.raw, C.NozzleTextureOrigin(origin), &mapping, &mapped)); err != nil {
		return nil, err
	}

	data, err := nativeMappedPixelsSlice(mapped)
	if err != nil {
		C.nozzle_pixel_mapping_unlock(&mapping)
		return nil, err
	}

	stride := int(mapped.row_stride_bytes)
	return &MappedPixels{
		Data:              data,
		RowStrideBytes:    stride,
		Width:             int(mapped.width),
		Height:            int(mapped.height),
		Format:            TextureFormat(mapped.format),
		Origin:            TextureOrigin(mapped.origin),
		frame:             f.raw,
		mapping:           mapping,
		writable:          true,
		locked:            true,
		osthreadLocked:    false,
		absRowStrideBytes: absInt(stride),
	}, nil
}

// CopyToGLTexture copies frame data to an OpenGL texture.
func (f *Frame) CopyToGLTexture(glName, glTarget, width, height uint32, format TextureFormat) error {
	return checkCode(C.nozzle_frame_copy_to_gl_texture(f.raw, C.uint32_t(glName), C.uint32_t(glTarget), C.uint32_t(width), C.uint32_t(height), C.NozzleTextureFormat(format)))
}

// WritableFrame represents a frame that can receive pixel data.
type WritableFrame struct {
	raw           *C.NozzleFrame
	activeMapping *MappedPixels
}

// Info returns frame metadata.
func (f *WritableFrame) Info() (FrameInfo, error) {
	var raw C.NozzleFrameInfo
	if err := checkCode(C.nozzle_frame_get_info(f.raw, &raw)); err != nil {
		return FrameInfo{}, err
	}
	return FrameInfo{
		FrameIndex:        uint64(raw.frame_index),
		Timestamp:         uint64(raw.timestamp_ns),
		Width:             uint32(raw.width),
		Height:            uint32(raw.height),
		Format:            TextureFormat(raw.format),
		SemanticFormat:    TextureFormat(raw.semantic_format),
		DroppedFrameCount: uint32(raw.dropped_frame_count),
	}, nil
}

// LockWritablePixels maps the frame pixel data for writing.
//
// The returned Data slice is a native mapped view, not a Go-owned copy. The
// native mapping lifetime is owned by an opaque core mapping handle.
func (f *WritableFrame) LockWritablePixels(origin TextureOrigin) (*MappedPixels, error) {
	var mapped C.NozzleMappedPixels
	var mapping *C.NozzlePixelMapping
	if err := checkCode(C.nozzle_frame_lock_writable_pixels_mapping_with_origin(
		f.raw, C.NozzleTextureOrigin(origin), &mapping, &mapped)); err != nil {
		return nil, err
	}

	data, err := nativeMappedPixelsSlice(mapped)
	if err != nil {
		C.nozzle_pixel_mapping_unlock(&mapping)
		return nil, err
	}

	stride := int(mapped.row_stride_bytes)
	pixels := &MappedPixels{
		Data:              data,
		RowStrideBytes:    stride,
		Width:             int(mapped.width),
		Height:            int(mapped.height),
		Format:            TextureFormat(mapped.format),
		Origin:            TextureOrigin(mapped.origin),
		frame:             f.raw,
		mapping:           mapping,
		owner:             f,
		writable:          true,
		locked:            true,
		osthreadLocked:    false,
		absRowStrideBytes: absInt(stride),
	}
	f.activeMapping = pixels
	return pixels, nil
}

// UnmapWritablePixelsChecked releases the writable pixel mapping and reports
// backend unlock failures. If this frame created a writable MappedPixels object,
// the mapping state is shared so frame-level and mapping-level unmap both
// invalidate the mapping and release the OS-thread pin exactly once.
func (f *WritableFrame) UnmapWritablePixelsChecked() error {
	if f.activeMapping != nil {
		return f.activeMapping.UnmapChecked()
	}
	return checkCode(C.nozzle_frame_unlock_writable_pixels_checked(f.raw))
}

// UnmapWritablePixels releases the writable pixel mapping and discards checked
// unlock errors.
func (f *WritableFrame) UnmapWritablePixels() {
	_ = f.UnmapWritablePixelsChecked()
}

func absInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

func compactPixelCopySize(width, height, bytesPerPixel int) (int, int, error) {
	if width <= 0 || height <= 0 {
		return 0, 0, fmt.Errorf("invalid pixel dimensions %dx%d", width, height)
	}
	if bytesPerPixel <= 0 {
		return 0, 0, fmt.Errorf("invalid bytes per pixel %d", bytesPerPixel)
	}
	rowStride := width * bytesPerPixel
	if rowStride/bytesPerPixel != width {
		return 0, 0, fmt.Errorf("pixel row size overflow")
	}
	total := rowStride * height
	if total/height != rowStride {
		return 0, 0, fmt.Errorf("pixel data size overflow")
	}
	return rowStride, total, nil
}

func mappedSize(mapped C.NozzleMappedPixels) (int, int, error) {
	stride := int(mapped.row_stride_bytes)
	absStride := absInt(stride)
	height := int(mapped.height)
	if mapped.data == nil {
		return 0, 0, fmt.Errorf("mapped pixel data is nil")
	}
	if absStride == 0 && height > 0 {
		return 0, 0, fmt.Errorf("invalid row stride %d", stride)
	}
	if height < 0 {
		return 0, 0, fmt.Errorf("invalid mapped height %d", height)
	}
	total := absStride * height
	if height != 0 && total/height != absStride {
		return 0, 0, fmt.Errorf("mapped pixel data size overflow")
	}
	return absStride, total, nil
}

func nativeMappedPixelsSlice(mapped C.NozzleMappedPixels) ([]byte, error) {
	absStride, total, err := mappedSize(mapped)
	if err != nil {
		return nil, err
	}
	base := mapped.data
	height := int(mapped.height)
	if int(mapped.row_stride_bytes) < 0 && height > 0 {
		base = unsafe.Add(base, -((height - 1) * absStride))
	}
	return unsafe.Slice((*byte)(base), total), nil
}

// EnumerateSenders returns the count of active senders.
func EnumerateSenders() (uint32, error) {
	var array C.NozzleSenderInfoArray
	if rc := C.nozzle_enumerate_senders(&array); rc != C.NOZZLE_OK {
		return 0, ErrorCode(rc)
	}
	count := uint32(array.count)
	C.nozzle_free_sender_info_array(&array)
	return count, nil
}

// IsGPUAvailable checks if a GPU backend is available by attempting sender creation.
func IsGPUAvailable() bool {
	name := C.CString("nozzle-go-gpu-check")
	appName := C.CString("nozzle-go")
	defer C.free(unsafe.Pointer(name))
	defer C.free(unsafe.Pointer(appName))

	cDesc := C.NozzleSenderDesc{
		name:             name,
		application_name: appName,
		ring_buffer_size: 1,
	}
	var raw *C.NozzleSender
	rc := C.nozzle_sender_create(&cDesc, &raw)
	if rc == C.NOZZLE_OK && raw != nil {
		C.nozzle_sender_destroy(raw)
		return true
	}
	return false
}

// ResolvedTextureFormat contains detailed format resolution info.
type ResolvedTextureFormat struct {
	StorageFormat  TextureFormat
	SemanticFormat TextureFormat
	FormatSource   FormatSource
	NativeBackend  BackendType
	NativeKind     NativeFormatKind
	NativeValue    uint32
	ChannelOrder   uint32
	ComponentType  uint32
	ComponentBits  uint8
	ChannelCount   uint8
	BytesPerPixel  uint8
}

// Device represents a GPU device handle.
type Device struct {
	raw *C.NozzleDevice
}

// GetDefaultDevice returns the default GPU device.
func GetDefaultDevice() (*Device, error) {
	var raw *C.NozzleDevice
	if err := checkCode(C.nozzle_device_get_default(&raw)); err != nil {
		return nil, err
	}
	return &Device{raw: raw}, nil
}

// Close destroys the device.
func (d *Device) Close() {
	if d.raw != nil {
		C.nozzle_device_destroy(d.raw)
		d.raw = nil
	}
}

// Texture represents a wrapped native GPU texture.
type Texture struct {
	raw *C.NozzleTexture
}

// TextureWrapDesc configures native texture wrapping.
type TextureWrapDesc struct {
	NativeTexture unsafe.Pointer
	Width         uint32
	Height        uint32
	Format        TextureFormat
	Backend       BackendType
}

// WrapTexture wraps a native GPU texture.
func WrapTexture(desc TextureWrapDesc) (*Texture, error) {
	cDesc := C.NozzleTextureWrapDesc{
		native_texture: desc.NativeTexture,
		width:          C.uint32_t(desc.Width),
		height:         C.uint32_t(desc.Height),
		format:         C.NozzleTextureFormat(desc.Format),
		backend:        C.NozzleBackendType(desc.Backend),
	}
	var raw *C.NozzleTexture
	if err := checkCode(C.nozzle_texture_wrap(&cDesc, &raw)); err != nil {
		return nil, err
	}
	return &Texture{raw: raw}, nil
}

// Close destroys the texture.
func (t *Texture) Close() {
	if t.raw != nil {
		C.nozzle_texture_destroy(t.raw)
		t.raw = nil
	}
}

// PublishTexture publishes a wrapped texture.
func (s *Sender) PublishTexture(tex *Texture) error {
	return checkCode(C.nozzle_sender_publish_texture(s.raw, tex.raw))
}

// PublishNativeTexture publishes a native GPU texture directly.
func (s *Sender) PublishNativeTexture(nativeTexture unsafe.Pointer, width, height uint32, format TextureFormat) error {
	return checkCode(C.nozzle_sender_publish_native_texture(s.raw, nativeTexture, C.uint32_t(width), C.uint32_t(height), C.NozzleTextureFormat(format)))
}

// PublishNativeTextureEx publishes a native GPU texture with explicit storage and semantic formats.
func (s *Sender) PublishNativeTextureEx(nativeTexture unsafe.Pointer, width, height uint32, storageFormat, semanticFormat TextureFormat) error {
	return checkCode(C.nozzle_sender_publish_native_texture_ex(s.raw, nativeTexture, C.uint32_t(width), C.uint32_t(height), C.NozzleTextureFormat(storageFormat), C.NozzleTextureFormat(semanticFormat)))
}

// ResolvedFormat returns detailed format resolution info for the frame.
func (f *Frame) ResolvedFormat() (ResolvedTextureFormat, error) {
	var raw C.NozzleResolvedTextureFormat
	if err := checkCode(C.nozzle_frame_get_resolved_format(f.raw, &raw)); err != nil {
		return ResolvedTextureFormat{}, err
	}
	return ResolvedTextureFormat{
		StorageFormat:  TextureFormat(raw.storage_format),
		SemanticFormat: TextureFormat(raw.semantic_format),
		FormatSource:   FormatSource(raw.format_source),
		NativeBackend:  BackendType(raw.native_backend),
		NativeKind:     NativeFormatKind(raw.native_kind),
		NativeValue:    uint32(raw.native_value),
		ChannelOrder:   uint32(raw.channel_order),
		ComponentType:  uint32(raw.component_type),
		ComponentBits:  uint8(raw.component_bits),
		ChannelCount:   uint8(raw.channel_count),
		BytesPerPixel:  uint8(raw.bytes_per_pixel),
	}, nil
}

// CopyToNativeTexture copies frame data to a native GPU texture.
func (f *Frame) CopyToNativeTexture(nativeTexture unsafe.Pointer, width, height uint32, format TextureFormat) error {
	return checkCode(C.nozzle_frame_copy_to_native_texture(f.raw, nativeTexture, C.uint32_t(width), C.uint32_t(height), C.NozzleTextureFormat(format)))
}

// SwizzleChannels rearranges channel order in pixel data.
func SwizzleChannels(src, dst []byte, width, height int64, srcRowBytes, dstRowBytes int64, format TextureFormat, permuteMap [4]uint8) error {
	return checkCode(C.nozzle_swizzle_channels(
		unsafe.Pointer(&src[0]),
		unsafe.Pointer(&dst[0]),
		C.uint32_t(width), C.uint32_t(height),
		C.int64_t(srcRowBytes), C.int64_t(dstRowBytes),
		C.NozzleTextureFormat(format),
		(*C.uint8_t)(&permuteMap[0]),
	))
}

// WidenUint16ToUint32 converts 16-bit pixel data to 32-bit.
func WidenUint16ToUint32(src, dst []byte, width, height int64, srcRowBytes, dstRowBytes int64, channels uint32) error {
	return checkCode(C.nozzle_widen_uint16_to_uint32(
		unsafe.Pointer(&src[0]),
		unsafe.Pointer(&dst[0]),
		C.uint32_t(width), C.uint32_t(height),
		C.int64_t(srcRowBytes), C.int64_t(dstRowBytes),
		C.uint32_t(channels),
	))
}

// ConvertUint32ToFloat32 converts 32-bit unsigned integer pixel data to float.
func ConvertUint32ToFloat32(src, dst []byte, width, height int64, srcRowBytes, dstRowBytes int64, channels uint32) error {
	return checkCode(C.nozzle_convert_uint32_to_float32(
		unsafe.Pointer(&src[0]),
		unsafe.Pointer(&dst[0]),
		C.uint32_t(width), C.uint32_t(height),
		C.int64_t(srcRowBytes), C.int64_t(dstRowBytes),
		C.uint32_t(channels),
	))
}

// ErrNoGPU is returned when GPU-dependent operations fail on headless systems.
var ErrNoGPU = errors.New("no GPU available")
