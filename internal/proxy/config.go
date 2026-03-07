package proxy

// ConfigProvider provides proxy configuration to ProxyManager
// This avoids circular dependency between core and proxy packages
type ConfigProvider interface {
	GetGlobalProxyEnabled() bool
	GetGlobalProxyType() string
	GetGlobalProxyURL() string
	GetGlobalProxyUsername() string
	GetGlobalProxyPassword() string
	GetBotProxyEnabled(botType string) bool
	GetBotProxyType(botType string) string
	GetBotProxyURL(botType string) string
	GetBotProxyUsername(botType string) string
	GetBotProxyPassword(botType string) string
}
