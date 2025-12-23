//go:build docker
// +build docker

package cmd

var (
	ConfigDirectory         = "/data"
	DefaultStateDirectory   = "/data"
	DefaultRuntimeDirectory = "/tmp"
	DefaultLibexecDirectory = "/usr/lib/mailchat"
)
