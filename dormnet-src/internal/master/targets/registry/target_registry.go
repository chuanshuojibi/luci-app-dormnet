package registry

import (
	"strings"

	"github.com/openwrt-dormnet/dormnet/shared/errx"
	"github.com/openwrt-dormnet/dormnet/shared/log"
	"go.uber.org/zap"
)

type TargetConstructor func(args *AbsDormnetClient) DormnetClient
type TargetRegistry struct {
	constant    DormnetClientConstants
	constructor TargetConstructor
}

var registry = map[string]TargetRegistry{}

func ListAll() []DormnetClientConstants {
	targets := make([]DormnetClientConstants, 0, len(registry))
	for k := range registry {
		targets = append(targets, registry[k].constant)
	}
	return targets
}

func Register(factory TargetConstructor) {
	client := factory(&AbsDormnetClient{})
	constant := client.Constants()
	id := strings.ToLower(constant.DormId)
	if HasClient(id) {
		log.Warn("duplicate target registry", zap.String("target", id))
	}
	registry[id] = TargetRegistry{
		constant:    constant,
		constructor: factory,
	}
}

func CreateClient(typ string, id string, username string, password string, loginIface string) (DormnetClient, errx.Exception) {
	client := registry[strings.ToLower(typ)].constructor(&AbsDormnetClient{
		id:         id,
		Username:   username,
		Password:   password,
		LoginIface: loginIface,
	})
	err := client.init()
	if err != nil {
		return nil, errx.NewExceptionWithCause(err, "failed to create dormnet client")
	}
	return client, nil
}

func HasClient(typ string) bool {
	_, exist := registry[strings.ToLower(typ)]
	return exist
}
