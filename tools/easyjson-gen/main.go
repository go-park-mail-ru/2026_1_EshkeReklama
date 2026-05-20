package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: easyjson-gen <dir>")
		os.Exit(2)
	}

	dir := os.Args[1]
	entries, err := os.ReadDir(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read dir %s: %v\n", dir, err)
		os.Exit(1)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if !strings.HasSuffix(name, ".go") {
			continue
		}
		if strings.HasSuffix(name, "_easyjson.go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		if slices.Contains([]string{"generate.go", "image.go", "partner_dictionary.go"}, name) {
			continue
		}

		cmd := exec.Command("go", "run", "github.com/mailru/easyjson/easyjson@v0.7.6", "-all", name)
		cmd.Dir = dir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "generate %s: %v\n", filepath.Join(dir, name), err)
			os.Exit(1)
		}
	}
}
