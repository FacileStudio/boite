package cmd

import (
	_ "embed"
)

//go:embed cloud-init.default.yml
var defaultCloudInitYAML string

func defaultCloudInit() string {
	return defaultCloudInitYAML
}
