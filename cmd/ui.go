package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/FacileStudio/boite/cmd/qemu"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
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

func printSuccess(msg string) {
	fmt.Printf("%s %s\n", successIcon, msg)
}

func printError(msg string) {
	fmt.Fprintf(os.Stderr, "%s %s\n", errorIcon, msg)
}

func printInfo(msg string) {
	fmt.Printf("%s %s\n", infoIcon, msg)
}

func printSessionEnd(msg string) {
	fmt.Printf("\n%s\n", lipgloss.NewStyle().Foreground(subtleColor).Render(msg))
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
	valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#F3F4F6"))
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

func renderInstanceTableQEMU(instances []*qemu.Instance) string {
	if len(instances) == 0 {
		emptyCard := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(borderColor).
			Padding(1, 2).
			Render(fmt.Sprintf("%s\n%s",
				lipgloss.NewStyle().Foreground(subtleColor).Render("No sandbox instances found."),
				lipgloss.NewStyle().Foreground(accentColor).Render("Run 'boite create <name>' to launch one."),
			))
		return emptyCard
	}

	t := table.New().
		Border(lipgloss.RoundedBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(borderColor)).
		Headers("NAME", "STATUS", "SSH PORT", "CREATED")

	for _, inst := range instances {
		var stateStyled string
		switch inst.Status {
		case "running":
			stateStyled = lipgloss.NewStyle().Foreground(successColor).Bold(true).Render(inst.Status)
		case "stopped":
			stateStyled = lipgloss.NewStyle().Foreground(subtleColor).Render(inst.Status)
		default:
			stateStyled = lipgloss.NewStyle().Foreground(warnColor).Render(inst.Status)
		}

		t.Row(
			lipgloss.NewStyle().Foreground(primaryColor).Bold(true).Render(inst.Name),
			stateStyled,
			lipgloss.NewStyle().Foreground(accentColor).Render(fmt.Sprintf("%d", inst.SSHPort)),
			lipgloss.NewStyle().Foreground(subtleColor).Render(inst.CreatedAt.Format("2006-01-02 15:04")),
		)
	}

	return t.Render()
}
