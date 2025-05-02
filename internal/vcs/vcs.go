package vcs

import (
	"fmt"
	"runtime/debug"
)

const apiVer = "1.0.0"

func Version(env string) string {
	var (
		revision string
		modified bool
	)

	switch env {
	case "production", "staging":
		return apiVer
	default:
		bi, ok := debug.ReadBuildInfo()
		if ok {
			for _, s := range bi.Settings {
				switch s.Key {
				case "vcs.revision":
					revision = s.Value
				case "vcs.modified":
					if s.Value == "true" {
						modified = true
					}
				}
			}
		}
	}
	fmt.Println(revision, modified)
	if modified {
		return fmt.Sprintf("%s-dirty", revision)
	}

	return revision
}
