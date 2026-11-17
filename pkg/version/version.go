// Package version holds the version metadata of gogit.
package version

// Version is the version of gogit. It is injected at build time via:
//
//	-ldflags "-X github.com/linuxsuren/gogit/pkg/version.Version=<version>"
var Version = "dev"
