package qemu

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProgressBarFillsLeftToRight(t *testing.T) {
	zero := CountChar(progressBarText(0.0), "█")
	low := CountChar(progressBarText(0.25), "█")
	high := CountChar(progressBarText(0.5), "█")
	full := CountChar(progressBarText(1.0), "█")

	assert.True(t, zero == 0)
	assert.True(t, low > 0 && low < high && high < full)
	assert.True(t, CountChar(progressBarText(1.0), "░") == 0)
}

func TestProgressBarCapsAtFullWidth(t *testing.T) {
	full := CountChar(progressBarText(1.0), "█")
	for _, f := range []float64{1.0, 1.25, 2.0} {
		assert.Equal(t, CountChar(progressBarText(f), "█"), full)
	}
}

func TestSpinnerFramesStableWidth(t *testing.T) {
	assert.Equal(t, len(progressFrames), 8)
	for i := range len(progressFrames) {
		assert.Equal(t, len(progressFrames[i]), len(progressFrames[0]))
		assert.Equal(t, spinnerFrame(i), progressFrames[i%len(progressFrames)])
	}
}

func CountChar(s, needle string) int {
	split := strings.Split(s, needle)
	return len(split) - 1
}
