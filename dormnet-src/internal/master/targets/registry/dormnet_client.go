package registry

import (
	"reflect"
	"strconv"

	"github.com/openwrt-dormnet/dormnet/shared/errx"
	"github.com/openwrt-dormnet/dormnet/shared/utils"
	"github.com/openwrt-dormnet/dormnet/shared/utils/uci"
	"go.uber.org/zap"
)

type DormnetClientConstants struct {
	DormId   string `json:"id"`
	DormName string `json:"name"`
}

type DormnetClient interface {
	Constants() DormnetClientConstants
	DefaultExtraArgs() any

	init() errx.Exception

	AccountId() string

	Process(ifaces []*DormnetClientBindIface) errx.Exception
	CheckInternet(ctx *utils.RequestContext) (bool, errx.Exception)
	NewRequestContext(iface *DormnetClientBindIface) (*utils.RequestContext, errx.Exception)
}

type AbsDormnetClient struct {
	id         string
	Username   string
	Password   string
	LoginIface string
}

type SampleDormnetClient struct {
	AbsDormnetClient
}

func (c *AbsDormnetClient) init() errx.Exception {
	return nil
}

func (c *AbsDormnetClient) Constants() DormnetClientConstants {
	panic(errx.NotImplementedError())
}

func (c *AbsDormnetClient) DefaultExtraArgs() any {
	return nil
}

func ParseExtraArgs(defaultVal any, bindIfaceSectionName string, out *DormnetClientBindIface) errx.Exception {
	bindIface := DormnetClientBindIface{
		Iface:     "",
		ExtraArgs: defaultVal,
	}

	iface, ok := uci.LookupString(bindIfaceSectionName, "iface")
	if !ok {
		return errx.NewException("No iface set", zap.String("section bind iface", bindIfaceSectionName))
	}
	bindIface.Iface = iface

	infos, err := FormatExtraArgInfo(defaultVal)
	if err != nil {
		return errx.NewExceptionWithError(err, "failed to format extra args")
	}

	for _, info := range infos {
		switch info.Type {
		case ExtraArgTypeFlag:
			if value, ok := uci.LookupBool(bindIfaceSectionName, info.Id); ok {
				info.field.SetBool(value)
			}
		case ExtraArgTypeValue:
			if value, ok := uci.LookupString(bindIfaceSectionName, info.Id); ok {
				info.field.SetString(value)
			}
		case ExtraArgTypeListValue:
			if value, ok := uci.LookupString(bindIfaceSectionName, info.Id); ok {
				typ := info.field.Type()
				var realValue any
				switch typ.Kind() {
				case reflect.String:
					realValue = value
				case reflect.Int:
					var err error
					if realValue, err = strconv.Atoi(value); err != nil {
						return errx.NewExceptionWithError(err, "failed to convert string to int",
							zap.String("type", typ.String()),
							zap.String("value", value))
					}
				default:
					panic(errx.NewException("ExtraArgTypeListValue only support type of int or string",
						zap.String("type", typ.String())))
				}
				refValue := reflect.ValueOf(realValue)
				if refValue.CanConvert(typ) {
					info.field.Set(refValue.Convert(typ))
				} else {
					return errx.NewException("cannot convert extra arg into ListValue item type",
						zap.String("type", typ.String()),
						zap.String("value", value))
				}
			}
		default:
			panic(errx.NewException("unknown extra arg type", zap.String("type", info.Type.String())))
		}
	}
	*out = bindIface
	return nil
}

func (c *AbsDormnetClient) AccountId() string {
	return c.id
}

func (c *AbsDormnetClient) CheckInternet(ctx *utils.RequestContext) (bool, errx.Exception) {
	checkFunc := func() (bool, errx.Exception) {
		pinger := utils.NewPinger(ctx.IfaceName)
		_, err := pinger.PingOnce("223.5.5.5")
		if err != nil {
			return false, nil
		} else {
			return true, nil
		}
	}
	var err errx.Exception
	for i := 0; i < 3; i++ {
		var ok bool
		ok, err = checkFunc()
		if err == nil && ok {
			return true, nil
		}
	}
	return false, err
}

func (c *AbsDormnetClient) NewRequestContext(iface *DormnetClientBindIface) (*utils.RequestContext, errx.Exception) {
	return utils.NewRequestContext(iface.Iface, iface.ExtraArgs)
}
