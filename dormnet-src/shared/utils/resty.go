package utils

import (
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/openwrt-dormnet/dormnet/shared/errx"
	"github.com/openwrt-dormnet/dormnet/shared/log"
	"go.uber.org/zap"
	"golang.org/x/net/html/charset"
)

func NewHttpClient() *resty.Client {
	return resty.New()
}

func NewHttpClientOnIface(iface string) *resty.Client {
	dialer := NewIfaceController(iface).CreateDialer(5 * time.Second)

	transport := &http.Transport{
		DialContext: dialer.DialContext,
	}

	client := resty.New()
	client.SetTransport(transport)

	return client
}

type RequestContext struct {
	client    *resty.Client
	IfaceName string
	Mac       string

	IfaceController IfaceController

	ExtraArgs any
}

func NewRequestContext(iface string, extraArgs any) (*RequestContext, errx.Exception) {
	controller := NewIfaceController(iface)
	up, err := controller.Status()
	if err != nil {
		return nil, errx.NewExceptionWithCause(err, "failed to query status", zap.String("iface", iface))
	}
	if !up.Up || !up.Available {
		err = controller.WaitForUp(time.Second * 5)
		if err != nil {
			return nil, nil
		}
	}
	mac, err := controller.QueryMac()
	if err != nil {
		return nil, errx.NewExceptionWithCause(err, "failed to query mac", zap.String("iface", iface))
	}
	return &RequestContext{
		client:          NewHttpClientOnIface(iface),
		IfaceName:       iface,
		Mac:             mac,
		IfaceController: controller,
		ExtraArgs:       extraArgs,
	}, nil
}

func (r *RequestContext) NewRequest() *resty.Request {
	return r.client.R()
}

func Utf8StringRespBody(resp *resty.Response, fallbackCharset string) (string, errx.Exception) {
	bodyRaw := resp.Body()
	contentType := resp.Header().Get("Content-Type")
	utf8Reader, err := charset.NewReader(strings.NewReader(string(bodyRaw)), contentType)
	if err != nil {
		log.Warn("failed to create charset reader", zap.String("content_type", contentType))
		utf8Reader, err = charset.NewReader(strings.NewReader(string(bodyRaw)), fallbackCharset)
	}
	if err != nil {
		return "", errx.NewExceptionWithError(err, "failed to create charset reader")
	}
	bodyRaw, err = io.ReadAll(utf8Reader)
	if err != nil {
		return "", errx.NewExceptionWithError(err, "failed to read utf8 content")
	}
	return string(bodyRaw), nil
}
