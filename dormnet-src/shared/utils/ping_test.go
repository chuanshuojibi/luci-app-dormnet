package utils

import "testing"

func TestPingOnce(t *testing.T) {
	pinger := NewPingerDefault()
	_, err := pinger.PingOnce("www.baidu.com")
	for err != nil {
		t.Error(err)
		err = err.Cause()
	}
}
