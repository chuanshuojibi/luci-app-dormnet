package airkv

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/openwrt-dormnet/dormnet/shared/errx"
	"github.com/openwrt-dormnet/dormnet/shared/log"
	"github.com/openwrt-dormnet/dormnet/shared/utils"
	"go.uber.org/zap"
	"golang.org/x/sys/unix"
)

type StoreServer struct {
	config *ServerConfig

	keyMutex utils.KeyMutex
}

func NewAirKVStoreServer(config *ServerConfig) (*StoreServer, errx.Exception) {
	if err := os.MkdirAll(config.StorePath, 0755); err != nil {
		return nil, errx.NewExceptionWithError(err, "failed to create store directory")
	}

	return &StoreServer{
		config:   config,
		keyMutex: utils.NewAirKVMutex(),
	}, nil
}

func (kv *StoreServer) initServer(r *gin.Engine) {
	r.Use(kv.checkAuthFunc())
	handleFunc := kv.handleRequestFunc()
	httpPath := fmt.Sprintf("%s/:id", kv.config.Prefix)
	r.GET(httpPath, handleFunc)
	r.POST(httpPath, handleFunc)
}

func (kv *StoreServer) checkAuthFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, kv.config.Prefix) && kv.config.Token != "" {
			auth := c.GetHeader("Authorization")
			prefix := "Bearer "
			if !strings.HasPrefix(auth, prefix) {
				c.AbortWithStatus(http.StatusUnauthorized)
				return
			}
			auth = auth[len(prefix):]
			if auth != kv.config.Token {
				c.AbortWithStatus(http.StatusUnauthorized)
				return
			}
		}

		c.Next()
	}
}

type Store struct {
	Content   string `json:"content"`
	Timestamp int64  `json:"timestamp"`
}

type Resp struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Detail  string `json:"detail"`
	Store
}

func (kv *StoreServer) handleRequestFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.Param("id")
		if strings.HasSuffix(key, "/") {
			key = key[:len(key)-1]
		}
		switch c.Request.Method {
		case "GET":
			kv.handleGet(c, key)
		case "POST":
			kv.handleSet(c, key)
		}
	}
}

func (kv *StoreServer) handleFailedMessage(c *gin.Context, key string, msg string) {
	log.Warn(msg, zap.String("key", key))
	c.JSON(http.StatusInternalServerError, &Resp{
		Success: false,
		Message: msg,
	})
}

func (kv *StoreServer) handleFailedMessageWithError(c *gin.Context, key string, err errx.Exception) {
	detail := err.Cause()
	if detail == nil {
		kv.handleFailedMessage(c, key, err.Error())
		return
	}
	log.Warn("failed to handel response", zap.String("key", key), zap.Any("error", detail))
	c.JSON(http.StatusInternalServerError, &Resp{
		Success: false,
		Message: err.Error(),
		Detail:  detail.Error(),
	})
}

func (kv *StoreServer) handleGet(c *gin.Context, key string) {
	err := kv.runInLock(key, false, func(file *os.File) errx.Exception {
		json, err := kv.readContent(file)
		if err != nil {
			return err
		}
		log.Debug("response success", zap.String("key", key), zap.Time("last_modify", time.UnixMilli(json.Timestamp)))
		c.JSON(http.StatusOK, &Resp{
			Success: true,
			Store:   *json,
		})
		return nil
	})
	if err != nil {
		kv.handleFailedMessageWithError(c, key, err)
	}
}

func (kv *StoreServer) handleSet(c *gin.Context, key string) {
	var body Store
	if err := c.ShouldBind(&body); err != nil {
		kv.handleFailedMessageWithError(c, key, errx.NewExceptionWithError(err, "failed to read content form request"))
		return
	}

	err := kv.runInLock(key, true, func(file *os.File) errx.Exception {
		rawJson, err := kv.readContent(file)
		if err != nil {
			return err
		}
		if rawJson.Timestamp != body.Timestamp {
			return errx.NewException("value changed, should not be updated this time")
		}

		body.Timestamp = time.Now().UnixMilli()
		newValue, _ := json.Marshal(body)
		if _, err := file.Write(newValue); err != nil {
			return errx.NewExceptionWithError(err, "failed to save content of file")
		}

		log.Debug("update success", zap.String("key", key), zap.Time("last_modify", time.UnixMilli(rawJson.Timestamp)))
		c.JSON(http.StatusOK, &Resp{
			Success: true,
		})
		return nil
	})
	if err != nil {
		kv.handleFailedMessageWithError(c, key, err)
	}
}

func (kv *StoreServer) runInLock(key string, createWhenNotExit bool, block func(file *os.File) errx.Exception) errx.Exception {
	filePath := filepath.Join(kv.config.StorePath, fmt.Sprintf("%s.airkv", key))
	flags := os.O_RDWR
	if createWhenNotExit {
		flags |= os.O_CREATE
	}
	file, err := os.OpenFile(filePath, flags, 0644)
	if err != nil {
		if os.IsNotExist(err) {
			return errx.NewExceptionWithError(err, "key not found")
		}
		return errx.NewExceptionWithError(err, "failed to open key")
	}
	defer file.Close()
	if err := unix.Flock(int(file.Fd()), unix.LOCK_EX); err != nil {
		return errx.NewExceptionWithError(err, "failed to lock file")
	}
	defer unix.Flock((int)(file.Fd()), unix.LOCK_UN)

	kv.keyMutex.Lock(key)
	defer kv.keyMutex.Unlock(key)

	return block(file)
}

func (kv *StoreServer) readContent(filePath *os.File) (*Store, errx.Exception) {
	data, err := io.ReadAll(filePath)
	if err != nil {
		return nil, errx.NewExceptionWithError(err, "failed to read content of file")
	}

	rawJson := new(Store)
	err = json.Unmarshal(data, &rawJson)
	if err != nil {
		return nil, errx.NewExceptionWithError(err, "failed to parse content of file")
	}

	return rawJson, nil
}
