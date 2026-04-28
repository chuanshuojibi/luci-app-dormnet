package uci

import (
	"strconv"
	"sync"

	"github.com/digineo/go-uci"
)

type UciContext interface {
	GetString(section string, option string, defVal string) string
	LookupString(section string, option string) (string, bool)

	GetInt(section string, option string, defVal int) int
	LookupInt(section string, option string) (int, bool)

	GetBool(section string, option string, defVal bool) bool
	LookupBool(section string, option string) (bool, bool)

	GetList(section string, option string) []string

	GetSections(typ string) []string

	CreateSection(typ string, section string) bool
	Set(section string, option string, value ...string) bool
	Add(section string, option string, value string) bool
}

type AbsUciContext struct {
	mutex  *sync.Mutex
	config string
}

func NewUciConfig(config string) UciContext {
	defer uci.LoadConfig(config, false)
	return &AbsUciContext{
		mutex:  &sync.Mutex{},
		config: config,
	}
}

func (c *AbsUciContext) GetString(section string, option string, defVal string) string {
	value, ok := c.LookupString(section, option)
	if !ok {
		return defVal
	} else {
		return value
	}
}

func (c *AbsUciContext) LookupString(section string, option string) (string, bool) {
	value, exist := uci.Get(c.config, section, option)
	if !exist || len(value) < 1 {
		return "", false
	}
	return value[0], true
}

func (c *AbsUciContext) GetInt(section string, option string, defVal int) int {
	value, ok := c.LookupInt(section, option)
	if !ok {
		return defVal
	} else {
		return value
	}
}

func (c *AbsUciContext) LookupInt(section string, option string) (int, bool) {
	value, ok := c.LookupString(section, option)
	if !ok || len(value) < 1 {
		return 0, false
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, false
	}
	return parsed, true
}

func (c *AbsUciContext) GetBool(section string, option string, defVal bool) bool {
	value, ok := c.LookupBool(section, option)
	if !ok {
		return defVal
	} else {
		return value
	}
}

func (c *AbsUciContext) LookupBool(section string, option string) (bool, bool) {
	return uci.GetBool(c.config, section, option)
}

func (c *AbsUciContext) GetList(section string, option string) []string {
	value, ok := uci.Get(c.config, section, option)
	if !ok {
		return []string{}
	} else {
		return value
	}
}

func (c *AbsUciContext) GetSections(section string) []string {
	value, ok := uci.GetSections(c.config, section)
	if !ok {
		return []string{}
	} else {
		return value
	}
}

func (c *AbsUciContext) CreateSection(typ string, name string) bool {
	err := uci.AddSection(c.config, name, typ)
	return err == nil
}

func (c *AbsUciContext) Set(section string, option string, value ...string) bool {
	return uci.Set(c.config, section, option, value...)
}

func (c *AbsUciContext) Add(section string, option string, value string) bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	existValue, ok := uci.Get(c.config, section, option)
	if !ok {
		return false
	}
	existValue = append(existValue, value)
	return c.Set(section, option, existValue...)
}
