package vad

import (
	"testing"
)

func TestNewEnergyVAD_Defaults(t *testing.T) {
	v := NewEnergyVAD(0, 0, 0)
	if v == nil {
		t.Fatal("vad is nil")
	}
	if v.frameMS != 30 {
		t.Errorf("frameMS = %d, want 30", v.frameMS)
	}
	if v.silenceMS != 800 {
		t.Errorf("silenceMS = %d, want 800", v.silenceMS)
	}
	if !v.enabled {
		t.Error("should be enabled by default")
	}
}

func TestNewEnergyVAD_ModeMapping(t *testing.T) {
	tests := []struct {
		mode      int
		threshold float64
	}{
		{0, 0.01},
		{1, 0.02},
		{2, 0.05},
		{3, 0.1},
	}
	for _, tt := range tests {
		v := NewEnergyVAD(tt.mode, 30, 800)
		if v.threshold != tt.threshold {
			t.Errorf("mode=%d threshold = %f, want %f", tt.mode, v.threshold, tt.threshold)
		}
	}
}

func TestNewEnergyVAD_ModeClamp(t *testing.T) {
	// Negative mode -> 0
	v := NewEnergyVAD(-1, 30, 800)
	if v.threshold != 0.01 {
		t.Errorf("mode=-1 threshold = %f, want 0.01", v.threshold)
	}
	// Too large mode -> last
	v = NewEnergyVAD(100, 30, 800)
	if v.threshold != 0.1 {
		t.Errorf("mode=100 threshold = %f, want 0.1", v.threshold)
	}
}

func TestComputeEnergy_Empty(t *testing.T) {
	if e := computeEnergy(nil); e != 0 {
		t.Errorf("empty energy = %f, want 0", e)
	}
	if e := computeEnergy([]int16{}); e != 0 {
		t.Errorf("empty energy = %f, want 0", e)
	}
}

func TestComputeEnergy_Silence(t *testing.T) {
	e := computeEnergy(make([]int16, 100))
	if e != 0 {
		t.Errorf("silence energy = %f, want 0", e)
	}
}

func TestComputeEnergy_Loud(t *testing.T) {
	pcm := make([]int16, 100)
	for i := range pcm {
		pcm[i] = 16384 // half of max amplitude
	}
	e := computeEnergy(pcm)
	if e <= 0 {
		t.Errorf("loud energy = %f, want > 0", e)
	}
	if e > 1.0 {
		t.Errorf("energy = %f, want <= 1.0", e)
	}
}

func TestProcess_SilenceOnly(t *testing.T) {
	v := NewEnergyVAD(0, 30, 800)
	// Feed silence frames
	for i := 0; i < 50; i++ {
		st := v.Process(make([]int16, 100))
		if st.Status != Silence {
			t.Errorf("frame %d: status = %v, want Silence", i, st.Status)
		}
	}
}

func TestProcess_SpeechDetection(t *testing.T) {
	v := NewEnergyVAD(0, 30, 800)
	// Loud frame
	pcm := make([]int16, 100)
	for i := range pcm {
		pcm[i] = 16384
	}
	st := v.Process(pcm)
	if st.Status != Speech {
		t.Errorf("loud frame status = %v, want Speech", st.Status)
	}
}

func TestProcess_SpeechToEnd(t *testing.T) {
	v := NewEnergyVAD(0, 30, 300) // 300ms silence threshold, 30ms per frame

	// Start with speech
	pcm := make([]int16, 100)
	for i := range pcm {
		pcm[i] = 16384
	}
	st := v.Process(pcm)
	if st.Status != Speech {
		t.Fatalf("speech status = %v", st.Status)
	}

	// Feed silence frames — need enough to trigger SpeechEnd
	silenceNeeded := v.silenceMS / v.frameMS // 300/30 = 10
	for i := 0; i < silenceNeeded; i++ {
		st = v.Process(make([]int16, 100))
	}
	if st.Status != SpeechEnd {
		t.Errorf("after %d silence frames: status = %v", silenceNeeded, st.Status)
	}

	// One more silence -> Silence
	st = v.Process(make([]int16, 100))
	if st.Status != Silence {
		t.Errorf("after SpeechEnd: status = %v, want Silence", st.Status)
	}
}

func TestProcess_Disabled(t *testing.T) {
	v := NewEnergyVAD(0, 30, 800)
	v.enabled = false
	pcm := make([]int16, 100)
	for i := range pcm {
		pcm[i] = 32767
	}
	st := v.Process(pcm)
	if st.Status != Silence {
		t.Errorf("disabled VAD should stay Silence, got %v", st.Status)
	}
}

func TestEvent_Type(t *testing.T) {
	ev := Event{Type: Speech}
	if ev.Type != Speech {
		t.Errorf("Event.Type = %v", ev.Type)
	}
}
