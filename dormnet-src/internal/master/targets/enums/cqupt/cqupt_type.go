//go:generate go tool go-enum --marshal --template ../../../../../templates/go_enum_values.tmpl

package cqupt

import "github.com/openwrt-dormnet/dormnet/internal/constants"

// ENUM(PC, Mobile)
type CquptClientType string

func (ct CquptClientType) AccountPrefix() string {
	switch ct {
	case CquptClientTypeMobile:
		return "1"
	case CquptClientTypePC:
		return "0"
	}
	return ""
}

func (ct CquptClientType) UserAgent() string {
	switch ct {
	case CquptClientTypeMobile:
		return constants.UserAgentMobile
	case CquptClientTypePC:
		return constants.UserAgentPC
	}
	return ""
}

func (ct CquptClientType) Callback() string {
	switch ct {
	case CquptClientTypeMobile:
		return "dr1005"
	case CquptClientTypePC:
		return "dr1003"
	}
	return ""
}
