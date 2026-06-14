package vad

import (
	"math"
	"sync"
	"time"
)

// Status VAD 检测结果。
type Status int

const (
	Silence Status = iota
	Speech
	SpeechEnd // 检测到语音结束
)

// State VAD 上下文：当前状态 + 连续帧计数。
type State struct {
	Status       Status
	SilenceFrames int // 连续静默帧数
	SpeechFrames  int // 连续语音帧数
}

// energyVAD 能量阈值端点检测，纯 Go 无外部依赖。
type energyVAD struct {
	mu         sync.Mutex
	state      State
	frameMS    int     // 每帧毫秒数
	silenceMS  int     // 静默阈值
	threshold  float64 // 能量阈值
	enabled    bool
}

// NewEnergyVAD 建一个 VAD。
// mode 0-3，越大阈值越高（越不容易触发语音）。
// frameMS 每帧毫秒数，silenceMS 连续静默多久算结束。
func NewEnergyVAD(mode, frameMS, silenceMS int) *energyVAD {
	if frameMS <= 0 {
		frameMS = 30
	}
	if silenceMS <= 0 {
		silenceMS = 800
	}
	// mode 映射到能量阈值
	thresholds := []float64{0.01, 0.02, 0.05, 0.1}
	if mode < 0 {
		mode = 0
	}
	if mode >= len(thresholds) {
		mode = len(thresholds) - 1
	}
	return &energyVAD{
		frameMS:   frameMS,
		silenceMS: silenceMS,
		threshold: thresholds[mode],
		enabled:   true,
	}
}

// Process 喂一帧 PCM int16 样本，返回当前状态。
func (v *energyVAD) Process(pcm []int16) State {
	v.mu.Lock()
	defer v.mu.Unlock()

	if !v.enabled || len(pcm) == 0 {
		return v.state
	}

	energy := computeEnergy(pcm)
	isSpeech := energy > v.threshold

	silenceFramesNeeded := v.silenceMS / v.frameMS
	if silenceFramesNeeded < 1 {
		silenceFramesNeeded = 1
	}

	if isSpeech {
		v.state.SilenceFrames = 0
		v.state.SpeechFrames++
		v.state.Status = Speech
	} else {
		v.state.SpeechFrames = 0
		v.state.SilenceFrames++

		if v.state.Status == Speech && v.state.SilenceFrames >= silenceFramesNeeded {
			v.state.Status = SpeechEnd
		} else if v.state.Status == SpeechEnd {
			v.state.Status = Silence
		}
	}

	return v.state
}

// computeEnergy PCM 归一化能量 0-1。
func computeEnergy(pcm []int16) float64 {
	if len(pcm) == 0 {
		return 0
	}
	var sum float64
	for _, v := range pcm {
		sum += float64(v) * float64(v)
	}
	mean := sum / float64(len(pcm))
	return math.Sqrt(mean) / 32768.0 // 归一化到 0-1
}

// Event 是 VAD 发出的时间事件（给外部驱动使用）。
type Event struct {
	Type Status
	Time time.Time
}
