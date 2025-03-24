// Fake podman CLI binary
package main

import (
	"os"
	"path"
	"path/filepath"
)

func main() {
	print("podman")
	for _, arg := range os.Args[1:] {
		print(" " + arg)
	}
	println()
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	outputTxt := path.Join(filepath.Dir(ex), "output.txt")
	_, err = os.Stat(outputTxt)
	if err == nil {
		data, _ := os.ReadFile(outputTxt)
		_, _ = os.Stdout.Write(data)
	}
	os.Exit(0)
}
