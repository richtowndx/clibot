package proxy

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProxyManager_NoProxy_ReturnsClientWithEnvProxy(t *testing.T) {
	config := &mockConfigProvider{
		globalEnabled: false,
	}

	pm := NewProxyManager(config)
	client, err := pm.GetHTTPClient("telegram")

	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.NotNil(t, client.Transport)
}

func TestProxyManager_GlobalProxy_ReturnsClientWithProxy(t *testing.T) {
	config := &mockConfigProvider{
		globalEnabled: true,
		globalType:    "http",
		globalURL:     "http://127.0.0.1:8080",
	}

	pm := NewProxyManager(config)
	client, err := pm.GetHTTPClient("telegram")

	assert.NoError(t, err)
	assert.NotNil(t, client)
}

func TestProxyManager_BotLevelProxyOverridesGlobal(t *testing.T) {
	config := &mockConfigProvider{
		globalEnabled: true,
		globalType:    "http",
		globalURL:     "http://127.0.0.1:8080",
		botEnabled:    true,
		botType:       "socks5",
		botURL:        "socks5://127.0.0.1:1080",
	}

	pm := NewProxyManager(config)
	proxyURL := pm.GetProxyURL("telegram")

	assert.Contains(t, proxyURL, "socks5://127.0.0.1:1080")
}

func TestProxyManager_ClientCaching(t *testing.T) {
	config := &mockConfigProvider{
		globalEnabled: true,
		globalType:    "http",
		globalURL:     "http://127.0.0.1:8080",
	}

	pm := NewProxyManager(config)

	// First call
	client1, err := pm.GetHTTPClient("telegram")
	assert.NoError(t, err)
	assert.NotNil(t, client1)

	// Second call should return cached client
	client2, err := pm.GetHTTPClient("telegram")
	assert.NoError(t, err)
	assert.Same(t, client1, client2)
}

func TestProxyManager_ClearCache(t *testing.T) {
	config := &mockConfigProvider{
		globalEnabled: true,
		globalType:    "http",
		globalURL:     "http://127.0.0.1:8080",
	}

	pm := NewProxyManager(config)

	// Create cached client
	client1, err := pm.GetHTTPClient("telegram")
	assert.NoError(t, err)
	assert.NotNil(t, client1)

	// Clear cache
	pm.ClearCache()

	// New client should be different
	client2, err := pm.GetHTTPClient("telegram")
	assert.NoError(t, err)
	assert.NotNil(t, client2)
	assert.NotSame(t, client1, client2)
}

func TestProxyManager_InvalidProxyURL(t *testing.T) {
	config := &mockConfigProvider{
		globalEnabled: true,
		globalType:    "http",
		globalURL:     "://invalid-url",
	}

	pm := NewProxyManager(config)
	_, err := pm.GetHTTPClient("telegram")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid proxy URL")
}

func TestProxyManager_UnsupportedProxyType(t *testing.T) {
	config := &mockConfigProvider{
		globalEnabled: true,
		globalType:    "ftp",
		globalURL:     "ftp://127.0.0.1:2121",
	}

	pm := NewProxyManager(config)
	_, err := pm.GetHTTPClient("telegram")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported proxy type")
}

func TestProxyManager_ProxyWithAuth(t *testing.T) {
	config := &mockConfigProvider{
		globalEnabled: true,
		globalType:    "http",
		globalURL:     "http://127.0.0.1:8080",
		globalUser:    "user",
		globalPass:    "pass",
	}

	pm := NewProxyManager(config)
	client, err := pm.GetHTTPClient("telegram")

	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.NotNil(t, client.Transport)
}

func TestProxyManager_Socks5Proxy(t *testing.T) {
	config := &mockConfigProvider{
		globalEnabled: true,
		globalType:    "socks5",
		globalURL:     "socks5://127.0.0.1:1080",
	}

	pm := NewProxyManager(config)
	client, err := pm.GetHTTPClient("telegram")

	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.NotNil(t, client.Transport)
}

func TestProxyManager_Socks5WithAuth(t *testing.T) {
	config := &mockConfigProvider{
		globalEnabled: true,
		globalType:    "socks5",
		globalURL:     "socks5://127.0.0.1:1080",
		globalUser:    "user",
		globalPass:    "pass",
	}

	pm := NewProxyManager(config)
	client, err := pm.GetHTTPClient("telegram")

	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.NotNil(t, client.Transport)
}
