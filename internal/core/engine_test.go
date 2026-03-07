package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEngine_ProxyManagerInitialized(t *testing.T) {
	config := &Config{
		Sessions: []SessionConfig{},
	}
	engine := NewEngine(config)

	assert.NotNil(t, engine.proxyMgr)
}

func TestEngine_GetProxyClient(t *testing.T) {
	config := &Config{
		Sessions: []SessionConfig{},
		Proxy: ProxyConfig{
			Enabled: false,
		},
	}
	engine := NewEngine(config)

	client, err := engine.proxyMgr.GetHTTPClient("telegram")
	assert.NoError(t, err)
	assert.NotNil(t, client)
}
