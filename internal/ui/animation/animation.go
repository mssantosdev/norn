package animation

import (
	"math"
	"time"
)

// Easing functions for smooth animations.
type Easing func(t float64) float64

// Linear interpolation.
func Linear(t float64) float64 {
	return t
}

// EaseOut decelerates toward the end.
func EaseOut(t float64) float64 {
	return 1 - math.Pow(1-t, 3)
}

// EaseInOut accelerates then decelerates.
func EaseInOut(t float64) float64 {
	if t < 0.5 {
		return 4 * t * t * t
	}
	return 1 - math.Pow(-2*t+2, 3)/2
}

// EaseOutBounce provides a bouncing deceleration.
func EaseOutBounce(t float64) float64 {
	const n1 = 7.5625
	const d1 = 2.75
	if t < 1/d1 {
		return n1 * t * t
	} else if t < 2/d1 {
		t -= 1.5 / d1
		return n1*t*t + 0.75
	} else if t < 2.5/d1 {
		t -= 2.25 / d1
		return n1*t*t + 0.9375
	}
	t -= 2.625 / d1
	return n1*t*t + 0.984375
}

// Timing constants for animation durations.
const (
	DurationMicro    = 80 * time.Millisecond
	DurationFast     = 120 * time.Millisecond
	DurationNormal   = 200 * time.Millisecond
	DurationSlow     = 300 * time.Millisecond
	DurationStagger  = 20 * time.Millisecond
	DurationListItem = 30 * time.Millisecond
)

// Frame represents a single animation frame with interpolated values.
type Frame struct {
	Progress float64 // 0.0 to 1.0
	Value    float64 // Interpolated value
}

// AnimateFloat generates frames for a float animation.
func AnimateFloat(from, to float64, duration time.Duration, easing Easing) []float64 {
	if duration <= 0 {
		return []float64{to}
	}
	frames := int(duration / (16 * time.Millisecond)) // ~60fps
	if frames < 1 {
		frames = 1
	}
	result := make([]float64, frames)
	for i := 0; i < frames; i++ {
		t := float64(i) / float64(frames-1)
		if easing != nil {
			t = easing(t)
		}
		result[i] = from + (to-from)*t
	}
	return result
}

// AnimateInt generates frames for an integer animation.
func AnimateInt(from, to int, duration time.Duration, easing Easing) []int {
	floatFrames := AnimateFloat(float64(from), float64(to), duration, easing)
	result := make([]int, len(floatFrames))
	for i, v := range floatFrames {
		result[i] = int(math.Round(v))
	}
	return result
}

// Stagger returns delays for staggered animations.
func Stagger(count int, delay time.Duration) []time.Duration {
	delays := make([]time.Duration, count)
	for i := 0; i < count; i++ {
		delays[i] = time.Duration(i) * delay
	}
	return delays
}

// ShakeFrames generates horizontal shake offsets for error feedback.
// Returns pixel offsets: left, right, left, rest.
func ShakeFrames(amplitude float64) []float64 {
	return []float64{-amplitude, amplitude, -amplitude, 0}
}

// FadeFrames generates opacity values for fade-in animation.
func FadeFrames(duration time.Duration) []float64 {
	return AnimateFloat(0, 1, duration, EaseOut)
}

// SlideUpFrames generates vertical offsets for slide-up entrance.
func SlideUpFrames(distance float64, duration time.Duration) []float64 {
	return AnimateFloat(distance, 0, duration, EaseOut)
}

// ProgressBarFrames generates progress bar fill characters for a given width.
func ProgressBarFrames(width int, duration time.Duration) []int {
	return AnimateInt(0, width, duration, EaseOut)
}

// CheckmarkFrames returns the sequence for a checkmark morph animation.
func CheckmarkFrames() []string {
	return []string{"◐", "◑", "◒", "◓", "✓"}
}

// SpinnerMorphFrames returns frames for spinner-to-checkmark morph.
func SpinnerMorphFrames() []string {
	return []string{"◐", "◑", "✓"}
}

// PulseFrames generates opacity values for a pulse effect (2 cycles).
func PulseFrames(duration time.Duration) []float64 {
	frames := int(duration / (16 * time.Millisecond))
	result := make([]float64, frames)
	for i := 0; i < frames; i++ {
		t := float64(i) / float64(frames)
		// Two full sine cycles
		result[i] = 0.5 + 0.5*math.Sin(t*4*math.Pi)
	}
	return result
}
