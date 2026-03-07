package core

import "github.com/keepmind9/clibot/internal/proxy"

// CoreConfigAdapter wraps core.Config to implement proxy.ConfigProvider
// This allows ProxyManager to access configuration without creating circular dependency
type CoreConfigAdapter struct {
	config *Config
}

func NewCoreConfigAdapter(config *Config) *CoreConfigAdapter {
	return &CoreConfigAdapter{config: config}
}

func (a *CoreConfigAdapter) GetGlobalProxyEnabled() bool {
	return a.config.Proxy.Enabled
}

func (a *CoreConfigAdapter) GetGlobalProxyType() string {
	return a.config.Proxy.Type
}

func (a *CoreConfigAdapter) GetGlobalProxyURL() string {
	return a.config.Proxy.URL
}

func (a *CoreConfigAdapter) GetGlobalProxyUsername() string {
	return a.config.Proxy.Username
}

func (a *CoreConfigAdapter) GetGlobalProxyPassword() string {
	return a.config.Proxy.Password
}

func (a *CoreConfigAdapter) GetBotProxyEnabled(botType string) bool {
	if botConfig, exists := a.config.Bots[botType]; exists && botConfig.Proxy != nil {
		return botConfig.Proxy.Enabled
	}
	return false
}

func (a *CoreConfigAdapter) GetBotProxyType(botType string) string {
	if botConfig, exists := a.config.Bots[botType]; exists && botConfig.Proxy != nil {
		return botConfig.Proxy.Type
	}
	return ""
}

func (a *CoreConfigAdapter) GetBotProxyURL(botType string) string {
	if botConfig, exists := a.config.Bots[botType]; exists && botConfig.Proxy != nil {
		return botConfig.Proxy.URL
	}
	return ""
}

func (a *CoreConfigAdapter) GetBotProxyUsername(botType string) string {
	if botConfig, exists := a.config.Bots[botType]; exists && botConfig.Proxy != nil {
		return botConfig.Proxy.Username
	}
	return ""
}

func (a *CoreConfigAdapter) GetBotProxyPassword(botType string) string {
	if botConfig, exists := a.config.Bots[botType]; exists && botConfig.Proxy != nil {
		return botConfig.Proxy.Password
	}
	return ""
}

var _ proxy.ConfigProvider = (*CoreConfigAdapter)(nil)
