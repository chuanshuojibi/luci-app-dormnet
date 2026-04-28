package airkv

import (
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/openwrt-dormnet/dormnet/shared/errx"
	"github.com/openwrt-dormnet/dormnet/shared/utils"
	"go.uber.org/zap"
)

type Client interface {
	Get(key string) (*Store, errx.Exception)
	Set(key string, value *Store) errx.Exception
}

type AbsClient struct {
	server     string
	token      string
	httpClient *resty.Client
}

func (c AbsClient) newRequest() *resty.Request {
	return c.httpClient.NewRequest()
}

func (c AbsClient) Get(key string) (*Store, errx.Exception) {
	var result Resp

	r := c.newRequest()
	r.SetHeader("Authorization", fmt.Sprintf("Bearer %s", c.token))
	r.SetResult(&result)
	_, err := r.Get(fmt.Sprintf("%s/%s", c.server, key))
	if err != nil {
		return nil, errx.NewExceptionWithError(err, "failed to create request", zap.String("key", key))
	}

	if !result.Success {
		return nil, errx.NewException(result.Message, zap.String("key", key), zap.String("detail", result.Message))
	}
	return &Store{
		Content:   result.Content,
		Timestamp: result.Timestamp,
	}, nil
}

func (c AbsClient) Set(key string, value *Store) errx.Exception {
	var result Resp

	r := c.newRequest()
	r.SetHeader("Authorization", fmt.Sprintf("Bearer %s", c.token))
	r.SetResult(&result)
	r.SetBody(*value)
	_, err := r.Post(fmt.Sprintf("%s/%s", c.server, key))
	if err != nil {
		return errx.NewExceptionWithError(err, "failed to create request", zap.String("key", key))
	}

	if !result.Success {
		return errx.NewException(result.Message, zap.String("key", key), zap.String("detail", result.Message))
	}
	return nil
}

func NewAirKVClient(server string, token string) Client {
	return &AbsClient{
		server:     server,
		token:      token,
		httpClient: utils.NewHttpClient(),
	}
}
