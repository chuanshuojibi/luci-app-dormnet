package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/openwrt-dormnet/dormnet/internal"
)

func main() {
	code, output := internal.MainFunc()
	if output != nil {
		bytes, _ := json.Marshal(output)
		fmt.Println(string(bytes))
	}
	os.Exit(code)
}
