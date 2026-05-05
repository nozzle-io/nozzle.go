package nozzle

import (
	"testing"
)

func TestTextureFormatBytesPerPixel(t *testing.T) {
	tests := []struct {
		format TextureFormat
		expect int
	}{
		{FormatR8UNorm, 1},
		{FormatRG8UNorm, 2},
		{FormatR16UNorm, 2},
		{FormatR16Float, 2},
		{FormatRGBA8UNorm, 4},
		{FormatBGRA8UNorm, 4},
		{FormatRGBA8SRGB, 4},
		{FormatBGRA8SRGB, 4},
		{FormatRG16UNorm, 4},
		{FormatRG16Float, 4},
		{FormatR32Float, 4},
		{FormatR32Uint, 4},
		{FormatDepth32Float, 4},
		{FormatRGBA16UNorm, 8},
		{FormatRGBA16Float, 8},
		{FormatRG32Float, 8},
		{FormatRGBA32Float, 16},
		{FormatRGBA32Uint, 16},
	}
	for _, tt := range tests {
		got := tt.format.BytesPerPixel()
		if got != tt.expect {
			t.Errorf("format %d BytesPerPixel = %d, want %d", int(tt.format), got, tt.expect)
		}
	}
	if FormatUnknown.BytesPerPixel() != 0 {
		t.Errorf("FormatUnknown.BytesPerPixel() = %d, want 0", FormatUnknown.BytesPerPixel())
	}
}

func TestErrorCodeValues(t *testing.T) {
	if int(ErrorUnknown) != 1 {
		t.Errorf("ErrorUnknown = %d, want 1", int(ErrorUnknown))
	}
	if int(ErrorInvalidArgument) != 2 {
		t.Errorf("ErrorInvalidArgument = %d, want 2", int(ErrorInvalidArgument))
	}
	if int(ErrorTimeout) != 10 {
		t.Errorf("ErrorTimeout = %d, want 10", int(ErrorTimeout))
	}
	if int(ErrorBackend) != 11 {
		t.Errorf("ErrorBackend = %d, want 11", int(ErrorBackend))
	}
}

func TestErrorCodeErrorMessages(t *testing.T) {
	codes := []ErrorCode{
		ErrorUnknown, ErrorInvalidArgument, ErrorUnsupportedBackend,
		ErrorUnsupportedFormat, ErrorDeviceMismatch, ErrorResourceCreation,
		ErrorSharedHandle, ErrorSenderNotFound, ErrorSenderClosed,
		ErrorTimeout, ErrorBackend,
	}
	for _, code := range codes {
		msg := code.Error()
		if len(msg) == 0 {
			t.Errorf("ErrorCode(%d).Error() returned empty string", int(code))
		}
	}
}

func TestBackendTypeValues(t *testing.T) {
	if int(BackendUnknown) != 0 {
		t.Fail()
	}
	if int(BackendD3D11) != 1 {
		t.Fail()
	}
	if int(BackendMetal) != 2 {
		t.Fail()
	}
	if int(BackendOpenGL) != 3 {
		t.Fail()
	}
	if int(BackendDMABuf) != 4 {
		t.Fail()
	}
}

func TestBackendTypeString(t *testing.T) {
	if BackendMetal.String() != "metal" {
		t.Errorf("BackendMetal.String() = %q", BackendMetal.String())
	}
	if BackendUnknown.String() != "unknown" {
		t.Errorf("BackendUnknown.String() = %q", BackendUnknown.String())
	}
}

func TestReceiveModeValues(t *testing.T) {
	if int(ReceiveLatestOnly) != 0 {
		t.Fail()
	}
	if int(ReceiveSequentialBestEffort) != 1 {
		t.Fail()
	}
}

func TestTextureOriginValues(t *testing.T) {
	if int(OriginTopLeft) != 0 {
		t.Fail()
	}
	if int(OriginBottomLeft) != 1 {
		t.Fail()
	}
}

func TestFrameStatusValues(t *testing.T) {
	if int(FrameNew) != 0 {
		t.Errorf("FrameNew = %d, want 0", int(FrameNew))
	}
	if int(FrameNoNew) != 1 {
		t.Errorf("FrameNoNew = %d, want 1", int(FrameNoNew))
	}
	if int(FrameDropped) != 2 {
		t.Errorf("FrameDropped = %d, want 2", int(FrameDropped))
	}
	if int(FrameSenderClosed) != 3 {
		t.Errorf("FrameSenderClosed = %d, want 3", int(FrameSenderClosed))
	}
	if int(FrameError) != 4 {
		t.Errorf("FrameError = %d, want 4", int(FrameError))
	}
}

func TestFormatSourceValues(t *testing.T) {
	if int(FormatSourceUnknown) != 0 {
		t.Errorf("FormatSourceUnknown = %d, want 0", int(FormatSourceUnknown))
	}
	if int(FormatSourceRequested) != 1 {
		t.Errorf("FormatSourceRequested = %d, want 1", int(FormatSourceRequested))
	}
	if int(FormatSourceCallerHint) != 2 {
		t.Errorf("FormatSourceCallerHint = %d, want 2", int(FormatSourceCallerHint))
	}
	if int(FormatSourceNativeObserved) != 3 {
		t.Errorf("FormatSourceNativeObserved = %d, want 3", int(FormatSourceNativeObserved))
	}
}

func TestNativeFormatKindValues(t *testing.T) {
	if int(NativeKindUnknown) != 0 {
		t.Errorf("NativeKindUnknown = %d, want 0", int(NativeKindUnknown))
	}
	if int(NativeKindMTLPixelFormat) != 1 {
		t.Errorf("NativeKindMTLPixelFormat = %d, want 1", int(NativeKindMTLPixelFormat))
	}
	if int(NativeKindDXGIFormat) != 2 {
		t.Errorf("NativeKindDXGIFormat = %d, want 2", int(NativeKindDXGIFormat))
	}
	if int(NativeKindDRMFourCC) != 3 {
		t.Errorf("NativeKindDRMFourCC = %d, want 3", int(NativeKindDRMFourCC))
	}
	if int(NativeKindGLInternalFormat) != 4 {
		t.Errorf("NativeKindGLInternalFormat = %d, want 4", int(NativeKindGLInternalFormat))
	}
}

func TestMappedPixelsRowBounds(t *testing.T) {
	data := make([]byte, 32)
	for i := range data {
		data[i] = byte(i)
	}
	mp := MappedPixels{
		Data:           data,
		RowStrideBytes: 8,
		Width:          8,
		Height:         4,
	}

	row0, err := mp.Row(0)
	if err != nil || len(row0) != 8 {
		t.Fatalf("Row(0) = %d, %v", len(row0), err)
	}
	row3, err := mp.Row(3)
	if err != nil || len(row3) != 8 {
		t.Fatalf("Row(3) = %d, %v", len(row3), err)
	}

	_, err = mp.Row(4)
	if err == nil {
		t.Error("Row(4) should fail")
	}
}

func TestSenderCreateAndDestroy(t *testing.T) {
	s, err := NewSender(SenderDesc{
		Name:            "go-test-sender",
		ApplicationName: "go-test",
		RingBufferSize:  3,
	})
	if err != nil {
		t.Skipf("no GPU: %v", err)
	}
	defer s.Close()

	info, err := s.Info()
	if err != nil {
		t.Fatalf("Info() failed: %v", err)
	}
	if info.Name != "go-test-sender" {
		t.Errorf("Name = %q, want %q", info.Name, "go-test-sender")
	}
	if info.ApplicationName != "go-test" {
		t.Errorf("ApplicationName = %q, want %q", info.ApplicationName, "go-test")
	}
}

func TestSenderEmptyNameFails(t *testing.T) {
	_, err := NewSender(SenderDesc{
		Name:            "",
		ApplicationName: "test",
	})
	if err != ErrorInvalidArgument {
		t.Errorf("expected ErrorInvalidArgument, got %v", err)
	}
}

func TestReceiverCreateAndDestroy(t *testing.T) {
	r, err := NewReceiver(ReceiverDesc{
		Name:            "go-test-nonexistent",
		ApplicationName: "go-test",
		ReceiveMode:     ReceiveLatestOnly,
	})
	if err != nil {
		t.Skipf("no GPU: %v", err)
	}
	defer r.Close()

	if r.IsConnected() {
		t.Error("should not be connected")
	}
}

func TestEnumerateSenders(t *testing.T) {
	count, err := EnumerateSenders()
	if err != nil {
		t.Skipf("enumerate failed: %v", err)
	}
	t.Logf("found %d senders", count)
}

func TestSenderWritableFrame(t *testing.T) {
	s, err := NewSender(SenderDesc{
		Name:            "go-test-frame",
		ApplicationName: "go-test",
	})
	if err != nil {
		t.Skipf("no GPU: %v", err)
	}
	defer s.Close()

	frame, err := s.AcquireWritableFrame(64, 64, FormatRGBA8UNorm)
	if err != nil {
		t.Skipf("acquire frame failed: %v", err)
	}

	info, err := frame.Info()
	if err != nil {
		t.Fatalf("frame Info() failed: %v", err)
	}
	if info.Width != 64 || info.Height != 64 {
		t.Errorf("frame size = %dx%d, want 64x64", info.Width, info.Height)
	}
	if info.Format.BytesPerPixel() != 4 {
		t.Errorf("format bpp = %d, want 4", info.Format.BytesPerPixel())
	}

	pixels, err := frame.LockWritablePixels(OriginTopLeft)
	if err != nil {
		t.Skipf("lock pixels failed: %v", err)
	}
	defer frame.UnmapWritablePixels()

	if pixels.Width != 64 || pixels.Height != 64 {
		t.Errorf("pixels size = %dx%d, want 64x64", pixels.Width, pixels.Height)
	}
	for i := range pixels.Data {
		pixels.Data[i] = 0xFF
	}

	if err := s.CommitFrame(frame); err != nil {
		t.Fatalf("CommitFrame failed: %v", err)
	}
}

func TestIsGPUAvailable(t *testing.T) {
	t.Logf("GPU available: %v", IsGPUAvailable())
}
