package qemu

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/term"
	"github.com/clipperhouse/displaywidth"
)

const progressBarWidth = 40

var barModel = progress.New(
	progress.WithSolidFill("#7D56F4"),
	progress.WithWidth(progressBarWidth),
)

var progressFrames = []string{
	"▰▱▱▱▱▱▱", "▰▰▱▱▱▱▱", "▰▰▰▱▱▱▱", "▰▰▰▰▱▱▱",
	"▰▰▰▰▰▱▱", "▰▰▰▰▰▰▱", "▰▰▰▰▰▰▰", "▰▱▱▱▱▱▱",
}

var progressAccent = lipgloss.Color("#00D2FF")

// renderLive redraws the single live line in place. Spinner mode renders
// "<spinner> <msg>", determinate mode renders "<bar> <msg>". The message is
// truncated to the terminal width so the line never wraps: a wrapped line
// becomes two screen rows and the carriage-return redraw can only clean the
// last row, leaving stale wrapped fragments stacked on screen.
func renderLive() {
	p := pstate.Load()
	if p == nil || !p.active {
		return
	}
	w := termWidth()
	if p.frac >= 0.0 {
		bar := progressBarText(p.frac)
		avail := w - displaywidth.String(bar) - 1
		fmt.Fprintf(os.Stderr, "\r\033[K%s %s", bar, truncateTo(p.msg, avail))
		return
	}
	spin := spinnerFrame(p.spin)
	avail := w - displaywidth.String(spin) - 1
	fmt.Fprintf(os.Stderr, "\r\033[K%s %s", spin, truncateTo(p.msg, avail))
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
	p := pstate.Load()
	if p != nil && p.color {
		return glyph(f, progressAccent)
	}
	return f
}

func glyph(s string, col lipgloss.Color) string {
	p := pstate.Load()
	if p == nil || !p.color {
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
