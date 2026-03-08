# Network Proxy Configuration

This guide explains how to configure network proxies for accessing IM platforms in restricted network environments.

## Quick Start

### Using Environment Variables (Simplest)

```bash
export HTTP_PROXY="http://127.0.0.1:7890"
export HTTPS_PROXY="http://127.0.0.1:7890"
clibot serve
```

### Using Configuration File

Add to your `config.yaml`:

```yaml
proxy:
  enabled: true
  url: "http://127.0.0.1:7890"
```

## Configuration Priority

clibot uses a three-tier fallback system:

1. **Bot-level proxy** (highest priority)
2. **Global proxy** (middle priority)
3. **Environment variables** (fallback)

## Proxy Types

### HTTP/HTTPS Proxy

```yaml
proxy:
  enabled: true
  url: "http://127.0.0.1:7890"
```

### SOCKS5 Proxy

```yaml
proxy:
  enabled: true
  url: "socks5://127.0.0.1:1080"
```

## Authentication

### HTTP Proxy with Auth

```yaml
proxy:
  enabled: true
  url: "http://proxy.example.com:8080"
  username: "your_username"
  password: "your_password"
```

### SOCKS5 with Auth

```yaml
proxy:
  enabled: true
  url: "socks5://proxy.example.com:1080"
  username: "your_username"
  password: "your_password"
```

## Bot-Level Proxy Configuration

Configure different proxies for different bots:

```yaml
proxy:
  enabled: true
  url: "http://127.0.0.1:8080"

bots:
  telegram:
    enabled: true
    token: "your_token"
    proxy:
      enabled: true  # Override global proxy
      url: "socks5://127.0.0.1:1080"  # Auto-detect proxy type from URL scheme

  discord:
    enabled: true
    token: "your_token"
    # Uses global proxy (no bot-level override)
```

**Note**: Proxy type (HTTP/HTTPS/SOCKS5) is automatically detected from the URL scheme. No need to specify a `type` field.

## WebSocket Support

All proxy types support WebSocket connections:

- **HTTP Proxy**: Uses HTTP CONNECT tunneling
- **SOCKS5 Proxy**: Native WebSocket support

Platforms using WebSocket:
- Discord
- Feishu/Lark
- DingTalk

## Troubleshooting

### Proxy Connection Failed

1. Check proxy server is running
2. Verify proxy URL format
3. Test proxy manually:
   ```bash
   curl -x http://127.0.0.1:8080 https://api.telegram.org
   ```

### Authentication Failed

1. Verify username and password
2. Check proxy server logs
3. Test authentication manually:
   ```bash
   curl -x http://user:pass@127.0.0.1:8080 https://api.telegram.org
   ```

### WebSocket Connection Failed

1. Ensure proxy supports CONNECT method (for HTTP proxy)
2. Try SOCKS5 proxy for better WebSocket support
3. Check firewall rules

## Common Proxy Servers

- **Clash**: https://github.com/Dreamacro/clash
- **V2Ray**: https://www.v2fly.org/
- **Shadowsocks**: https://shadowsocks.org/

## Examples

### Clash Example

```yaml
proxy:
  enabled: true
  url: "http://127.0.0.1:7890"  # Clash default HTTP port
```

### V2Ray Example

```yaml
proxy:
  enabled: true
  url: "socks5://127.0.0.1:1080"  # V2Ray default SOCKS5 port
```

### Environment Variables Only

No config changes needed:

```bash
export HTTP_PROXY="http://127.0.0.1:7890"
export HTTPS_PROXY="http://127.0.0.1:7890"
export ALL_PROXY="socks5://127.0.0.1:1080"

clibot serve --config config.yaml
```

## Platform Support

| Platform | Config File Proxy | Environment Variable | Notes |
|----------|------------------|---------------------|-------|
| Telegram | ✅ Full Support | ✅ Supported | Both HTTP API and long polling work with config proxy |
| Discord | ⚠️ Partial | ✅ **Required** | HTTP API (sending) works with config, **WebSocket (receiving) requires env vars** |
| Feishu | ⚠️ Limited | ✅ **Recommended** | SDK doesn't support custom client, use env vars |
| DingTalk | ⚠️ Limited | ✅ **Recommended** | SDK doesn't support custom client, use env vars |

### Platform-Specific Details

#### Telegram ✅ Full Support
- **Sending messages**: Uses config file proxy ✅
- **Receiving messages**: Uses config file proxy ✅
- **Configuration**: Both operations support custom HTTP client

#### Discord ⚠️ Partial Support
- **Sending messages (HTTP API)**: Uses config file proxy ✅
- **Receiving messages (WebSocket)**: **Requires environment variables** ❌

**Why?** Discord's WebSocket connection uses `gorilla/websocket` library, which requires a concrete `*websocket.Dialer` type. This cannot be customized via configuration files.

**Recommended setup for Discord**:
```bash
# Use environment variables for WebSocket connection
export HTTP_PROXY="http://172.30.64.9:10808"
export HTTPS_PROXY="http://172.30.64.9:10808"
clibot serve --config config.yaml
```

#### Feishu & DingTalk ⚠️ Limited Support
- Their official SDKs don't support custom HTTP clients
- Config file proxy will be logged but **not actually used**
- **Must use environment variables** for reliable proxy support

**Recommended setup for Feishu/DingTalk**:
```bash
export HTTP_PROXY="http://127.0.0.1:7890"
export HTTPS_PROXY="http://127.0.0.1:7890"
clibot serve --config config.yaml
```

Or in one command:
```bash
HTTP_PROXY="http://127.0.0.1:7890" HTTPS_PROXY="http://127.0.0.1:7890" clibot serve
```
