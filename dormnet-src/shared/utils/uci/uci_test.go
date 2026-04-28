package uci

import (
	"strings"
	"testing"

	"github.com/digineo/go-uci"
)

func TestUciSections(t *testing.T) {
	config := uci.NewTree("./bin/config")
	err := config.LoadConfig("dormnet", true)
	if err != nil {
		t.Fatal(err)
	}
	tests, ok := config.GetSections("dormnet", "test")
	t.Log(ok)
	t.Log(strings.Join(tests, ", "))
}
