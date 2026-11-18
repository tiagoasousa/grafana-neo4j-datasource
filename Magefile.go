//go:build mage
// +build mage

package main

import (
	"os"
	"os/exec"

	// mage:import
	build "github.com/grafana/grafana-plugin-sdk-go/build"
)

// Default configures the default target.
var Default = build.BuildAll

// TestCI runs the backend suite in short mode: DB-backed integration tests
// are skipped because CI has no Neo4j. For the full suite run `go test ./...`
// locally with the dev compose stack up.
func TestCI() error {
	cmd := exec.Command("go", "test", "-short", "./...")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
