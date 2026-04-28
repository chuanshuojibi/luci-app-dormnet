package registry

import (
	"encoding/json"
	"testing"
)

type CustomDormnetClientExtraArgs struct {
	Test string `json:"test"`
}

func GetType() any {
	return &CustomDormnetClientExtraArgs{
		Test: "test",
	}
}

func TestJson(t *testing.T) {
	val := GetType()
	str, err := json.Marshal(val)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(str))
}
