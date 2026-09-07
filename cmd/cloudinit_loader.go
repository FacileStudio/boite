package cmd

import (
	_ "embed"
)

//go:embed cloudinit.yml
var defaultCloudInitYAML string

func defaultCloudInit() string {
	return defaultCloudInitYAML
}
