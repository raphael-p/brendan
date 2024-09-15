package utils

import (
	"fmt"
	"os"
	"path/filepath"
)

func GetExecDirectory(fallback string) string {
	if os.Args[0] == "brendan" {
		// if executed from compiled binary, use its directory
		ex, err := os.Executable()
		if err != nil {
			panic(fmt.Sprintf("failed to locate executable: %s", err))
		}
		return filepath.Dir(ex)
	} else {
		// otherwise, use fallback directory (relative path)
		return fallback
	}
}
