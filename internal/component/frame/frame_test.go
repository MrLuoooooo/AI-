package frame

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/png"
	"testing"
	"time"
)

// makeTestFrame 生成一个简单 PNG 测试帧，返回 Base64 编码字符串。
func makeTestFrame(w, h int) string {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	var buf bytes.Buffer
	png.Encode(&buf, img)
	raw := "data:image/png;base64," + base64.StdEncoding.EncodeToString(buf.Bytes())
	return raw
}

func TestNewSampler_Defaults(t *testing.T) {
	s := NewSampler(0, 0, 0)
	if s.interval != time.Second {
		t.Errorf("interval = %v, want 1s", s.interval)
	}
	if s.quality != 30 {
		t.Errorf("quality = %d, want 30", s.quality)
	}
	if s.maxWidth != 640 {
		t.Errorf("maxWidth = %d, want 640", s.maxWidth)
	}
}

func TestNewSampler_Custom(t *testing.T) {
	s := NewSampler(500, 50, 1280)
	if s.interval != 500*time.Millisecond {
		t.Errorf("interval = %v", s.interval)
	}
	if s.quality != 50 {
		t.Errorf("quality = %d", s.quality)
	}
	if s.maxWidth != 1280 {
		t.Errorf("maxWidth = %d", s.maxWidth)
	}
}

func TestShouldSample_FirstFrame(t *testing.T) {
	s := NewSampler(1000, 30, 640)
	ok, data, err := s.ShouldSample(makeTestFrame(320, 240))
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Error("first frame should be sampled")
	}
	if data == "" {
		t.Error("returned data should not be empty")
	}
}

func TestShouldSample_IntervalFilter(t *testing.T) {
	s := NewSampler(100000, 30, 640) // 100s interval
	ok, _, err := s.ShouldSample(makeTestFrame(320, 240))
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("first frame should pass")
	}
	// Second frame within interval -> rejected
	ok, _, err = s.ShouldSample(makeTestFrame(320, 240))
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Error("second frame should be filtered by interval")
	}
}

func TestShouldSample_SameFrame(t *testing.T) {
	s := NewSampler(0, 30, 640) // 0 interval (passes time check)
	frame := makeTestFrame(320, 240)
	ok, _, err := s.ShouldSample(frame)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("first frame should pass")
	}
	// Same frame -> hash match -> rejected
	ok, _, err = s.ShouldSample(frame)
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Error("identical frame should be filtered")
	}
}

func TestShouldSample_InvalidBase64(t *testing.T) {
	s := NewSampler(0, 30, 640)
	_, _, err := s.ShouldSample("not-valid-base64!!!")
	if err == nil {
		t.Error("expected error for invalid base64")
	}
}

func TestStripDataURI(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"data:image/png;base64,abc123", "abc123"},
		{"data:image/jpeg;base64,def456", "def456"},
		{"plain-string", "plain-string"},
	}
	for _, tt := range tests {
		got := stripDataURI(tt.input)
		if got != tt.want {
			t.Errorf("stripDataURI(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestHammingDistance_Same(t *testing.T) {
	if d := hammingDistance(0xDEADBEEF, 0xDEADBEEF); d != 0 {
		t.Errorf("distance = %d, want 0", d)
	}
}

func TestHammingDistance_Different(t *testing.T) {
	d := hammingDistance(0, ^uint64(0))
	if d != 64 {
		t.Errorf("distance = %d, want 64", d)
	}
}

func TestPerceptualHash(t *testing.T) {
	// Create a non-uniform image (upper half white, lower half black)
	img := image.NewGray(image.Rect(0, 0, 64, 64))
	for y := 0; y < 32; y++ {
		for x := 0; x < 64; x++ {
			img.Pix[y*img.Stride+x] = 255
		}
	}
	h := perceptualHash(img)
	if h == 0 {
		t.Error("hash should not be 0 for non-uniform image")
	}
}

func TestEncodeFrame(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 320, 240))
	var buf bytes.Buffer
	err := encodeFrame(&buf, img)
	if err != nil {
		t.Fatal(err)
	}
	if buf.Len() == 0 {
		t.Error("encoded frame is empty")
	}
}
