package utils

import (
	"testing"
)

func TestPeerOutput(t *testing.T) {
	cli, err := NewShellCli()
	if err != nil {
		t.Fatal(err)
	}
	defer cli.Close()

	code, stdout, err := cli.Run("easytier-cli", "--output", "json", "peer")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("code: %d", code)
	t.Logf("stdout: %s", stdout)
}
