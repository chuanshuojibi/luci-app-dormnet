//go:generate go tool go-enum --marshal --template ../../../../../templates/go_enum_values.tmpl

package cqupt

// ENUM(telecom, cmcc, unicom)
type CquptClientOperator string
