package airkv

import (
	"github.com/gin-gonic/gin"
	"github.com/openwrt-dormnet/dormnet/shared/errx"
	"github.com/openwrt-dormnet/dormnet/shared/utils/uci"
)

type ServerConfig struct {
	Prefix    string `yaml:"prefix"`
	StorePath string `yaml:"store_path"`
	Token     string `yaml:"token"`
}

func CreateAirKVDaemon(r *gin.Engine) errx.Exception {
	config := &ServerConfig{}
	config.Prefix = uci.GetString("airkv", "prefix", "/airkv")
	config.StorePath = uci.GetString("airkv", "store_path", "/etc/dormnet/airkv")
	config.Token = uci.GetString("airkv", "token", "")

	kvstore, err := NewAirKVStoreServer(config)
	if err != nil {
		return err
	}
	kvstore.initServer(r)

	return nil
}
