package frame

import (
	"bytes"
	"encoding/base64"
	"image"
	_ "image/jpeg"
	"image/png"
	"strings"
	"sync"
	"time"

	"golang.org/x/image/draw"
)

// Sampler 自适应帧采样器 —— 帧差法降频 + 等比缩放 + JPEG 质量压缩。
type Sampler struct {
	mu           sync.Mutex
	lastFrame    *processedFrame // 上一帧缓存
	interval     time.Duration   // 最小采样间隔
	quality      int             // JPEG 质量 1-100
	maxWidth     int             // 缩放后最大宽度
	lastSampleAt time.Time       // 上次采样时间
}

type processedFrame struct {
	data    []byte // JPEG bytes
	hash    uint64 // 简易感知哈希
	width   int
	height  int
}

// NewSampler 创建一个自适应帧采样器。
func NewSampler(intervalMs, quality, maxWidth int) *Sampler {
	if intervalMs <= 0 {
		intervalMs = 1000
	}
	if quality <= 0 || quality > 100 {
		quality = 30
	}
	if maxWidth <= 0 {
		maxWidth = 640
	}
	return &Sampler{
		interval: time.Duration(intervalMs) * time.Millisecond,
		quality:  quality,
		maxWidth: maxWidth,
	}
}

// ShouldSample 判断当前帧是否值得采样。
// frameBase64 是前端传来的 Base64 JPEG/PNG 帧。
// 返回值：(是否采样, 同一帧的 Base64 数据, 错误)
func (s *Sampler) ShouldSample(frameBase64 string) (bool, string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()

	// 1. 时间间隔过滤
	if !s.lastSampleAt.IsZero() && now.Sub(s.lastSampleAt) < s.interval {
		return false, "", nil
	}

	// 2. 解码 Base64
	frameBytes, err := base64.StdEncoding.DecodeString(stripDataURI(frameBase64))
	if err != nil {
		return false, "", err
	}

	// 3. 缩放 + 重新编码为低质量 JPEG
	compressed, pHash, w, h, err := s.process(frameBytes)
	if err != nil {
		return false, "", err
	}

	// 4. 帧差过滤（与上一帧的感知哈希对比）
	if s.lastFrame != nil && s.lastFrame.width == w && s.lastFrame.height == h {
		if hammingDistance(s.lastFrame.hash, pHash) < 3 {
			return false, "", nil
		}
	}

	s.lastFrame = &processedFrame{data: compressed, hash: pHash, width: w, height: h}
	s.lastSampleAt = now

	output := base64.StdEncoding.EncodeToString(compressed)
	return true, output, nil
}

// process 缩放图片并按 JPEG 重新编码。
func (s *Sampler) process(raw []byte) ([]byte, uint64, int, int, error) {
	img, _, err := image.Decode(bytes.NewReader(raw))
	if err != nil {
		return nil, 0, 0, 0, err
	}

	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()

	// 等比缩放
	if w > s.maxWidth {
		ratio := float64(s.maxWidth) / float64(w)
		h = int(float64(h) * ratio)
		w = s.maxWidth
	}

	if w != bounds.Dx() {
		resized := image.NewRGBA(image.Rect(0, 0, w, h))
		draw.CatmullRom.Scale(resized, resized.Bounds(), img, bounds, draw.Over, nil)
		img = resized
	}

	var buf bytes.Buffer
	if err := encodeFrame(&buf, img); err != nil {
		return nil, 0, 0, 0, err
	}

	data := buf.Bytes()
	pHash := perceptualHash(img)

	return data, pHash, w, h, nil
}

// perceptualHash 算简易 64 位感知哈希。
func perceptualHash(img image.Image) uint64 {
	// 缩放到 8x8 灰度，取均值二值化
	small := image.NewGray(image.Rect(0, 0, 8, 8))
	draw.CatmullRom.Scale(small, small.Bounds(), img, img.Bounds(), draw.Over, nil)

	var sum int
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			sum += int(small.GrayAt(x, y).Y)
		}
	}
	avg := byte(sum / 64)

	var hash uint64
	idx := 0
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			if small.GrayAt(x, y).Y > avg {
				hash |= 1 << idx
			}
			idx++
		}
	}
	return hash
}

// hammingDistance 算两个哈希的汉明距离。
func hammingDistance(a, b uint64) int {
	xor := a ^ b
	dist := 0
	for xor != 0 {
		dist++
		xor &= xor - 1
	}
	return dist
}

// encodeFrame 将图片编码为 PNG。GPT-4o Vision 支持 PNG，且 Go 标准库 PNG 编码器无 cgo 依赖。
func encodeFrame(buf *bytes.Buffer, img image.Image) error {
	encoder := png.Encoder{CompressionLevel: png.BestSpeed}
	return encoder.Encode(buf, img)
}

// stripDataURI 去掉 "data:image/jpeg;base64," 前缀。
func stripDataURI(s string) string {
	if idx := strings.Index(s, ","); idx != -1 {
		return s[idx+1:]
	}
	return s
}
