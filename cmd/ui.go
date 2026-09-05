package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

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
	fmt.Fprintf(os.Stderr, "%s %s\n", infoIcon, msg)
}

func styleBanner(banner, version string) string {
	b := lipgloss.NewStyle().Foreground(primaryColor).Bold(true).Render(banner)
	v := lipgloss.NewStyle().Foreground(subtleColor).Render(version)
	return fmt.Sprintf("%s\n%s\n", b, v)
}

func renderSandboxCard(name, workspace string, noMount bool) string {
	titleStyle := lipgloss.NewStyle().Foreground(primaryColor).Bold(true)
	labelStyle := lipgloss.NewStyle().Foreground(accentColor)
	valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#F3F4F6"))
	cmdStyle := lipgloss.NewStyle().Foreground(successColor).Bold(true)

	lines := []string{
		titleStyle.Render("Sandbox Ready"),
		"",
		fmt.Sprintf("%s  %s", labelStyle.Render("VM:       "), valueStyle.Render(name)),
		fmt.Sprintf("%s  %s", labelStyle.Render("User:     "), valueStyle.Render("boite (/home/boite)")),
	}

	if !noMount {
		wsVal := fmt.Sprintf("/workspace (%s)", workspace)
		lines = append(lines, fmt.Sprintf("%s  %s", labelStyle.Render("Workspace:"), valueStyle.Render(wsVal)))
	}

	lines = append(lines,
		"",
		fmt.Sprintf("%s  %s", labelStyle.Render("Connect:  "), cmdStyle.Render("boite shell "+name)),
	)

	content := strings.Join(lines, "\n")
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(1, 2).
		Render(content)
}

type multipassListResponse struct {
	List []struct {
		Name    string   `json:"name"`
		State   string   `json:"state"`
		IPv4    []string `json:"ipv4"`
		Release string   `json:"release"`
	} `json:"list"`
}

func renderInstanceTable(rawJSON []byte) (string, bool) {
	var resp multipassListResponse
	if err := json.Unmarshal(rawJSON, &resp); err != nil {
		return "", false
	}

	if len(resp.List) == 0 {
		emptyCard := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(borderColor).
			Padding(1, 2).
			Render(fmt.Sprintf("%s\n%s",
				lipgloss.NewStyle().Foreground(subtleColor).Render("No sandbox instances found."),
				lipgloss.NewStyle().Foreground(accentColor).Render("Run 'boite create <name>' to launch one."),
			))
		return emptyCard, true
	}

	t := table.New().
		Border(lipgloss.RoundedBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(borderColor)).
		Headers("NAME", "STATE", "IPV4", "RELEASE")

	for _, item := range resp.List {
		var stateStyled string
		switch strings.ToLower(item.State) {
		case "running":
			stateStyled = lipgloss.NewStyle().Foreground(successColor).Bold(true).Render(item.State)
		case "stopped":
			stateStyled = lipgloss.NewStyle().Foreground(subtleColor).Render(item.State)
		case "deleted":
			stateStyled = lipgloss.NewStyle().Foreground(errColor).Render(item.State)
		default:
			stateStyled = lipgloss.NewStyle().Foreground(warnColor).Render(item.State)
		}

		ip := "-"
		if len(item.IPv4) > 0 {
			ip = strings.Join(item.IPv4, ", ")
		}

		t.Row(
			lipgloss.NewStyle().Foreground(primaryColor).Bold(true).Render(item.Name),
			stateStyled,
			lipgloss.NewStyle().Foreground(accentColor).Render(ip),
			lipgloss.NewStyle().Foreground(subtleColor).Render(item.Release),
		)
	}

	return t.Render(), true
}
