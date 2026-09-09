package qemu

import (
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/term"
	"github.com/clipperhouse/displaywidth"
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
var (
	progressMode    int
	progressActive  bool
	progressMsg     string
	progressFrac    float64
	progressSpin    int
	progressHalt    chan bool
	progressStopped chan bool
	useColor        bool

	barModel = progress.New(
		progress.WithSolidFill("#7D56F4"),
		progress.WithWidth(progressBarWidth),
	)
)

const (
	progressModePlain    = 1
	progressModeAnimated = 2
	progressBarWidth     = 40
)

var progressFrames = []string{
	"▰▱▱▱▱▱▱", "▰▰▱▱▱▱▱", "▰▰▰▱▱▱▱", "▰▰▰▰▱▱▱",
	"▰▰▰▰▰▱▱", "▰▰▰▰▰▰▱", "▰▰▰▰▰▰▰", "▰▱▱▱▱▱▱",
}

var progressAccent = lipgloss.Color("#00D2FF")

// ProgressStart enables animated progress. Call it before the first
// ProgressPhase/ProgressTick; without it, progress calls degrade to plain
// lines (or no-ops when called before any start). It must be paired with
// ProgressStop once progress output is no longer needed.
func ProgressStart() {
	progressMode = progressModePlain
	if isTty(os.Stderr) {
		progressMode = progressModeAnimated
	}
	useColor = os.Getenv("NO_COLOR") == ""
	progressActive = false
	progressFrac = -1.0
	progressSpin = 0
	if progressMode == progressModeAnimated {
		progressHalt = make(chan bool, 1)
		progressStopped = make(chan bool, 1)
		go func() {
			for {
				select {
				case <-progressHalt:
					progressStopped <- true
					return
				default:
					progressSpin += 1
					renderLive()
					time.Sleep(80 * time.Millisecond)
				}
			}
		}()
	}
}

// ProgressStop stops the animation and clears the live line if one is shown.
// Settled lines were already printed and stay.
func ProgressStop() {
	if progressMode == 0 {
		return
	}
	if progressMode == progressModeAnimated {
		if progressHalt != nil {
			progressHalt <- true
			<-progressStopped
		}
		if progressActive {
			clearLine()
			progressActive = false
		}
	}
	progressMode = 0
}

// ProgressPhase begins a new step, settling the previous step as a permanent
// line.
func ProgressPhase(msg string) {
	if progressMode == 0 {
		return
	}
	if progressMode == progressModePlain {
		fmt.Fprintf(os.Stderr, "▸ %s\n", msg)
		return
	}
	if progressActive {
		clearLine()
		fmt.Fprintf(os.Stderr, "▸ %s\n", progressMsg)
	}
	progressActive = true
	progressMsg = msg
	progressFrac = -1.0
}

// ProgressTick updates the active step's message and determinate progress. A
// non-negative frac lights the solid bar; a negative frac keeps the spinner.
func ProgressTick(msg string, frac float64) {
	if progressMode == 0 {
		return
	}
	progressActive = true
	progressMsg = msg
	progressFrac = frac
}

// ProgressDone settles the active step as a completed line.
func ProgressDone(msg string) {
	if progressMode == 0 {
		return
	}
	if progressMode == progressModePlain {
		fmt.Fprintf(os.Stderr, "✓ %s\n", msg)
		return
	}
	if progressActive {
		clearLine()
	}
	progressActive = false
	fmt.Fprintf(os.Stderr, "✓ %s\n", msg)
	progressFrac = -1.0
}

// ProgressFail settles the active step as a failed line.
func ProgressFail(msg string) {
	if progressMode == 0 {
		return
	}
	if progressMode == progressModePlain {
		fmt.Fprintf(os.Stderr, "✗ %s\n", msg)
		return
	}
	if progressActive {
		clearLine()
	}
	progressActive = false
	fmt.Fprintf(os.Stderr, "✗ %s\n", msg)
	progressFrac = -1.0
}

// ProgressWarn settles the active step as a warning line: the step finished,
// but not cleanly. Used when a guest boot completes with non-fatal noise that
// should not fail the whole create.
func ProgressWarn(msg string) {
	if progressMode == 0 {
		return
	}
	if progressMode == progressModePlain {
		fmt.Fprintf(os.Stderr, "! %s\n", msg)
		return
	}
	if progressActive {
		clearLine()
	}
	progressActive = false
	fmt.Fprintf(os.Stderr, "! %s\n", msg)
	progressFrac = -1.0
}

// renderLive redraws the single live line in place. Spinner mode renders
// "<spinner> <msg>", determinate mode renders "<bar> <msg>". The message is
// truncated to the terminal width so the line never wraps: a wrapped line
// becomes two screen rows and the carriage-return redraw can only clean the
// last row, leaving stale wrapped fragments stacked on screen.
func renderLive() {
	if !progressActive {
		return
	}
	w := termWidth()
	if progressFrac >= 0.0 {
		bar := progressBarText(progressFrac)
		avail := w - displaywidth.String(bar) - 1
		fmt.Fprintf(os.Stderr, "\r\033[K%s %s", bar, truncateTo(progressMsg, avail))
		return
	}
	spin := spinnerFrame(progressSpin)
	avail := w - displaywidth.String(spin) - 1
	fmt.Fprintf(os.Stderr, "\r\033[K%s %s", spin, truncateTo(progressMsg, avail))
}

func clearLine() {
	fmt.Fprintf(os.Stderr, "\r\033[K")
}

// termWidth returns the visible width of the terminal, defaulting to 80 when
// it cannot be measured.
func termWidth() int {
	w, _, err := term.GetSize(2)
	if err != nil || w < 20 {
		return 80
	}
	return w
}

// truncateTo shortens s (by grapheme width) to at most maxCols display columns,
// appending an ellipsis when it has to cut.
func truncateTo(s string, maxCols int) string {
	if maxCols < 0 {
		return ""
	}
	if displaywidth.String(s) <= maxCols {
		return s
	}
	ell := "…"
	ellW := 1
	if maxCols <= ellW {
		return ""
	}
	cols := 0
	out := ""
	g := displaywidth.StringGraphemes(s)
	for g.Next() {
		v := g.Value()
		w := g.Width()
		if cols+w > maxCols-ellW {
			break
		}
		out = out + v
		cols += w
	}
	return out + ell
}

// progressBarText renders the fraction using the charmbracelet bubbles
// progress bar.
func progressBarText(frac float64) string {
	return barModel.ViewAs(frac)
}

func spinnerFrame(spin int) string {
	f := progressFrames[spin%len(progressFrames)]
	if useColor {
		return glyph(f, progressAccent)
	}
	return f
}

func glyph(s string, col lipgloss.Color) string {
	if !useColor {
		return s
	}
	return lipgloss.NewStyle().Foreground(col).Render(s)
}
func isTty(f *os.File) bool {
	info, err := f.Stat()
	if err != nil {
		return false
	}
	return (info.Mode() & os.ModeCharDevice) != 0
}
