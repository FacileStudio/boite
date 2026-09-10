package qemu

import (
	"fmt"
	"os"
	"sync/atomic"
	"time"
)

// Progress output goes to stderr so generated stdout stays clean. On a TTY the
// helpers animate a single live line: a spinner and message while a step runs,
// or a determinate progress bar plus percent when the step reports a fraction.
// When stderr is not a TTY they degrade to plain lines with no animation.
//
// Only one line is ever animated, redrawn in place with a carriage return and
// an erase-to-end-of-line. Completed steps are printed as permanent lines and
// left alone. Because nothing ever moves the cursor off the current line and
// nothing is erased across lines, the output cannot scroll stray frames into
// the terminal scrollback the way a multi-line cursor-up redraw can.
//
// All progress state lives in one pointer-sized slot swapped atomically. The
// caller's goroutine owns every field except spin, which the animation
// goroutine advances on each frame, so a full snapshot is exchanged per
// update and the animation loop never reads a partially-updated step.
type progressState struct {
	mode   int
	active bool
	msg    string
	frac   float64
	spin   int
	color  bool
}

var pstate atomic.Pointer[progressState]

var (
	progressHalt    = make(chan bool, 1)
	progressStopped = make(chan bool, 1)
)

const (
	progressModePlain    = 1
	progressModeAnimated = 2
)

// ProgressStart enables animated progress. Call it before the first
// ProgressPhase/ProgressTick; without it, progress calls degrade to plain
// lines (or no-ops when called before any start). It must be paired with
// ProgressStop once progress output is no longer needed.
func ProgressStart() {
	mode := progressModePlain
	if isTty(os.Stderr) {
		mode = progressModeAnimated
	}
	pstate.Store(&progressState{
		mode:   mode,
		active: false,
		frac:   -1.0,
		spin:   0,
		color:  os.Getenv("NO_COLOR") == "",
	})
	if mode == progressModeAnimated {
		go animateProgress()
	}
}

// animateProgress redraws the live line at roughly 12fps until ProgressStop
// signals on progressHalt. Spin lives outside the caller's snapshot so the
// two writers never fight over the same field.
func animateProgress() {
	for {
		select {
		case <-progressHalt:
			progressStopped <- true
			return
		default:
			next := *pstate.Load()
			next.spin += 1
			pstate.Store(&next)
			renderLive()
			time.Sleep(80 * time.Millisecond)
		}
	}
}

// ProgressStop stops the animation and clears the live line if one is shown.
// Settled lines were already printed and stay.
func ProgressStop() {
	p := pstate.Load()
	if p == nil || p.mode == 0 {
		return
	}
	if p.mode == progressModeAnimated {
		progressHalt <- true
		<-progressStopped
		if p.active {
			clearLine()
			p.active = false
		}
	}
	p.mode = 0
	pstate.Store(p)
}

// ProgressPhase begins a new step, settling the previous step as a permanent
// line.
func ProgressPhase(msg string) {
	p := pstate.Load()
	if p == nil || p.mode == 0 {
		return
	}
	if p.mode == progressModePlain {
		fmt.Fprintf(os.Stderr, "▸ %s\n", msg)
		return
	}
	if p.active {
		clearLine()
		fmt.Fprintf(os.Stderr, "▸ %s\n", p.msg)
	}
	p.active = true
	p.msg = msg
	p.frac = -1.0
	pstate.Store(p)
}

// ProgressTick updates the active step's message and determinate progress. A
// non-negative frac lights the solid bar; a negative frac keeps the spinner.
func ProgressTick(msg string, frac float64) {
	p := pstate.Load()
	if p == nil || p.mode == 0 {
		return
	}
	p.active = true
	p.msg = msg
	p.frac = frac
	pstate.Store(p)
}

// ProgressDone settles the active step as a completed line.
func ProgressDone(msg string) {
	p := pstate.Load()
	if p == nil || p.mode == 0 {
		return
	}
	if p.mode == progressModePlain {
		fmt.Fprintf(os.Stderr, "✓ %s\n", msg)
		return
	}
	if p.active {
		clearLine()
	}
	p.active = false
	fmt.Fprintf(os.Stderr, "✓ %s\n", msg)
	p.frac = -1.0
	pstate.Store(p)
}

// ProgressFail settles the active step as a failed line.
func ProgressFail(msg string) {
	p := pstate.Load()
	if p == nil || p.mode == 0 {
		return
	}
	if p.mode == progressModePlain {
		fmt.Fprintf(os.Stderr, "✗ %s\n", msg)
		return
	}
	if p.active {
		clearLine()
	}
	p.active = false
	fmt.Fprintf(os.Stderr, "✗ %s\n", msg)
	p.frac = -1.0
	pstate.Store(p)
}

// ProgressWarn settles the active step as a warning line: the step finished,
// but not cleanly. Used when a guest boot completes with non-fatal noise that
// should not fail the whole create.
func ProgressWarn(msg string) {
	p := pstate.Load()
	if p == nil || p.mode == 0 {
		return
	}
	if p.mode == progressModePlain {
		fmt.Fprintf(os.Stderr, "! %s\n", msg)
		return
	}
	if p.active {
		clearLine()
	}
	p.active = false
	fmt.Fprintf(os.Stderr, "! %s\n", msg)
	p.frac = -1.0
	pstate.Store(p)
}
