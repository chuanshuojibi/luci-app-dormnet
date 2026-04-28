//go:generate go tool go-enum

package cqupt

// ENUM(PortalUnreachable, GatewayUnreachable, NotLogin, LoginButNotAvailable, Login)
type CquptLoginStatus string
