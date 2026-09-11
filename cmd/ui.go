package cmd

import (
	"fmt"
	"os"
	"strings"

	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"
	"github.com/FacileStudio/boite/cmd/qemu"
)

var (
	primaryColor = lipgloss.Color("#7D56F4")
	accentColor  = lipgloss.Color("#00D2FF")
	successColor = lipgloss.Color("#10B981")
	errColor     = lipgloss.Color("#EF4444")
	warnColor    = lipgloss.Color("#F59E0B")
	subtleColor  = lipgloss.Color("#6B7280")
	borderColor  = lipgloss.Color("#4B5563")

	successIcon = lipgloss.NewStyle().Foreground(successColor).Bold(true).Render("✓")
	errorIcon   = lipgloss.NewStyle().Foreground(errColor).Bold(true).Render("✗")
	infoIcon    = lipgloss.NewStyle().Foreground(accentColor).Bold(true).Render("▸")
)

// The printers route through lipgloss's colorprofile writers, which strip ANSI
// entirely when the target stream is not a TTY or NO_COLOR is set. Piped output
// (logs, scripts) therefore carries no escape sequences.

func printSuccess(msg string) {
	lipgloss.Println(successIcon, msg)
}

func printError(msg string) {
	lipgloss.Fprintln(os.Stderr, errorIcon, msg)
}

func printInfo(msg string) {
	lipgloss.Println(infoIcon, msg)
}

func printSessionEnd(msg string) {
	fmt.Println()
	lipgloss.Println(lipgloss.NewStyle().Foreground(subtleColor).Render(msg))
}

func styleBanner(ascii string, version string) string {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(0, 1).
		Render(ascii + " v" + version)
}

func renderSandboxCard(name, workspace string, noMount bool, sshPort int) string {
	titleStyle := lipgloss.NewStyle().Foreground(primaryColor).Bold(true)
	labelStyle := lipgloss.NewStyle().Foreground(accentColor)
	valueStyle := lipgloss.NewStyle().Foreground(subtleColor)
	cmdStyle := lipgloss.NewStyle().Foreground(successColor).Bold(true)

	lines := []string{
		titleStyle.Render("Sandbox Ready"),
		"",
		fmt.Sprintf("%s  %s", labelStyle.Render("VM:       "), valueStyle.Render(name)),
		fmt.Sprintf("%s  %s", labelStyle.Render("User:     "), valueStyle.Render("boite (/home/boite)")),
		fmt.Sprintf("%s  %s", labelStyle.Render("SSH Port: "), valueStyle.Render(fmt.Sprintf("%d", sshPort))),
	}

	if !noMount {
		wsVal := fmt.Sprintf("/workspace (%s)", workspace)
		lines = append(lines, fmt.Sprintf("%s  %s", labelStyle.Render("Workspace:"), valueStyle.Render(wsVal)))
	}

	lines = append(lines,
		"",
		fmt.Sprintf("%s  %s", labelStyle.Render("Connect:  "), cmdStyle.Render("boite run "+name)),
	)

	content := strings.Join(lines, "\n")
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(1, 2).
		Render(content)
}

// renderEmptyTable renders the placeholder shown when no instance exists.
func renderEmptyTable() string {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(1, 2).
		Render(fmt.Sprintf("%s\n%s",
			lipgloss.NewStyle().Foreground(subtleColor).Render("No sandbox instances found."),
			lipgloss.NewStyle().Foreground(accentColor).Render("Run 'boite create <name>' to launch one."),
		))
}

func renderInstanceTableQEMU(instances []*qemu.Instance) string {
	if len(instances) == 0 {
		return renderEmptyTable()
	}

	header := lipgloss.NewStyle().Foreground(primaryColor).Bold(true)
	t := table.New().
		Border(lipgloss.RoundedBorder()).
		BorderStyle(header).
		Headers("NAME", "STATUS", "SSH PORT", "CREATED").
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == table.HeaderRow {
				return header
			}
			switch col {
			case 0:
				return lipgloss.NewStyle().Foreground(primaryColor).Bold(true)
			case 1:
				return statusCellStyle(instances[row].Status)
			case 2:
				return lipgloss.NewStyle().Foreground(accentColor)
			default:
				return lipgloss.NewStyle().Foreground(subtleColor)
			}
		})

	for _, inst := range instances {
		t.Row(inst.Name, inst.Status, fmt.Sprintf("%d", inst.SSHPort), inst.CreatedAt.Format("2006-01-02 15:04"))
	}
	return t.Render()
}
