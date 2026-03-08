package bot

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/keepmind9/clibot/internal/proxy"
)

// QQBot implements BotAdapter interface for QQ Bot Platform (QQ群机器人开放平台)
// Official API: https://bots.qq.com
type QQBot struct {
	DefaultTypingIndicator
	mu             sync.RWMutex
	appID          string
	appSecret      string
	accessToken    string
	tokenExpiresAt time.Time
	wsConn         *websocket.Conn
	gatewayURL     string
	messageHandler func(BotMessage)
	ctx            context.Context
	cancel         context.CancelFunc
	proxyMgr       proxy.Manager
	msgSeqMap      map[string]int // Track message sequences for passive reply
	lastSequence   *int
}

// QQ Bot API endpoints
const (
	QQTokenURL   = "https://bots.qq.com/app/getAppAccessToken"
	QQAPIBase    = "https://api.sgroup.qq.com"
	QQGatewayURL = QQAPIBase + "/gateway"
)

// WebSocket OP codes (https://bots.qq.com/docs/gateway/gateway-events)
const (
	OPDispatch       = 0  // Event dispatch
	OPHeartbeat      = 1  // Heartbeat
	OPIdentify       = 2  // Identify
	OPResume         = 6  // Resume
	OPReconnect      = 7  // Reconnect request
	OPInvalidSession = 9  // Invalid session
	OPHello          = 10 // Hello
	OPHeartbeatAck   = 11 // Heartbeat acknowledgement
)

// Intents for subscribing to events
const (
	IntentPublicMessages = 1 << 25 // Public message events (1 << 25)
)

// GatewayPayload represents a WebSocket gateway message
type GatewayPayload struct {
	OP int         `json:"op"`
	D  interface{} `json:"d,omitempty"`
	S  *int        `json:"s,omitempty"`
	T  string      `json:"t,omitempty"`
}

// C2CMessageData represents C2C (private chat) message data from WebSocket event
type C2CMessageData struct {
	ID        string `json:"id"`
	Timestamp string `json:"timestamp"`
	Author    struct {
		UserOpenID string `json:"user_openid"`
	} `json:"author"`
	Content string `json:"content"`
}

// HelloData contains heartbeat_interval from OP Hello
type HelloData struct {
	HeartbeatInterval int `json:"heartbeat_interval"`
}

// IdentifyData contains identify payload
type IdentifyData struct {
	Token   string `json:"token"`
	Intents int    `json:"intents"`
	Shard   []int  `json:"shard"`
}

// NewQQBot creates a new QQ bot instance
func NewQQBot(appID, appSecret string) *QQBot {
	return &QQBot{
		appID:     appID,
		appSecret: appSecret,
		msgSeqMap: make(map[string]int),
	}
}

// SetProxyManager sets the proxy manager for the QQ bot
func (q *QQBot) SetProxyManager(mgr proxy.Manager) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.proxyMgr = mgr
}

// SupportsTypingIndicator returns false (QQ doesn't support typing indicators)
func (q *QQBot) SupportsTypingIndicator() bool {
	return false
}

// AddTypingIndicator does nothing (not supported)
func (q *QQBot) AddTypingIndicator(messageID string) bool {
	return false
}

// RemoveTypingIndicator does nothing (not supported)
func (q *QQBot) RemoveTypingIndicator(messageID string) error {
	return nil
}

// nextMsgSeq returns the next message sequence number for tracking passive replies
func (q *QQBot) nextMsgSeq(inboundMsgID string) int {
	q.mu.Lock()
	defer q.mu.Unlock()

	seq := q.msgSeqMap[inboundMsgID] + 1
	q.msgSeqMap[inboundMsgID] = seq

	// Prevent unbounded growth
	if len(q.msgSeqMap) > 500 {
		for key := range q.msgSeqMap {
			delete(q.msgSeqMap, key)
			break
		}
	}

	return seq
}

// setMessageHandler sets the message handler (for internal use)
func (q *QQBot) setMessageHandler(handler func(BotMessage)) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.messageHandler = handler
}

// sendGateway sends a payload to the WebSocket gateway
func (q *QQBot) sendGateway(payload GatewayPayload) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	q.mu.RLock()
	defer q.mu.RUnlock()

	if q.wsConn == nil {
		return fmt.Errorf("websocket not connected")
	}

	return q.wsConn.WriteMessage(websocket.TextMessage, data)
}

// startHeartbeat starts the heartbeat loop
func (q *QQBot) startHeartbeat(intervalMs int) {
	ticker := time.NewTicker(time.Duration(intervalMs) * time.Millisecond)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-q.ctx.Done():
				return
			case <-ticker.C:
				heartbeat := GatewayPayload{OP: OPHeartbeat, D: q.lastSequence}
				if err := q.sendGateway(heartbeat); err != nil {
					log.Printf("[QQ bot] Heartbeat failed: %v", err)
				}
			}
		}
	}()
}

// Start establishes connection to QQ gateway and begins listening for messages
func (q *QQBot) Start(messageHandler func(BotMessage)) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.messageHandler = messageHandler
	q.ctx, q.cancel = context.WithCancel(context.Background())

	token, err := q.getAccessToken()
	if err != nil {
		return fmt.Errorf("get access token: %w", err)
	}

	gatewayURL, err := q.getGatewayURL(token)
	if err != nil {
		return fmt.Errorf("get gateway: %w", err)
	}
	q.gatewayURL = gatewayURL

	if err := q.connectGateway(token); err != nil {
		return fmt.Errorf("connect gateway: %w", err)
	}

	log.Printf("[QQ bot] Started")
	return nil
}

// connectGateway establishes WebSocket connection to QQ gateway
func (q *QQBot) connectGateway(token string) error {
	dialer := websocket.DefaultDialer
	ws, _, err := dialer.Dial(q.gatewayURL, nil)
	if err != nil {
		return fmt.Errorf("dial websocket: %w", err)
	}

	q.wsConn = ws
	go q.handleWebSocketMessages(token)
	return nil
}

// handleWebSocketMessages receives and processes WebSocket messages
func (q *QQBot) handleWebSocketMessages(token string) {
	for {
		select {
		case <-q.ctx.Done():
			return
		default:
			q.mu.RLock()
			ws := q.wsConn
			q.mu.RUnlock()

			if ws == nil {
				return
			}

			_, message, err := ws.ReadMessage()
			if err != nil {
				log.Printf("[QQ bot] WebSocket read error: %v", err)
				q.scheduleReconnect()
				return
			}

			var payload GatewayPayload
			if err := json.Unmarshal(message, &payload); err != nil {
				log.Printf("[QQ bot] Failed to unmarshal payload: %v", err)
				continue
			}

			q.handleGatewayPayload(payload, token)
		}
	}
}

// handleGatewayPayload processes gateway payloads based on OP code
func (q *QQBot) handleGatewayPayload(payload GatewayPayload, token string) {
	switch payload.OP {
	case OPHello:
		// Server sent hello with heartbeat interval
		if helloData, ok := payload.D.(map[string]interface{}); ok {
			if interval, ok := helloData["heartbeat_interval"].(float64); ok {
				q.startHeartbeat(int(interval))
			}
		}

		// Send identify
		identify := GatewayPayload{
			OP: OPIdentify,
			D: IdentifyData{
				Token:   fmt.Sprintf("QQBot %s", token),
				Intents: IntentPublicMessages,
				Shard:   []int{0, 1},
			},
		}
		if err := q.sendGateway(identify); err != nil {
			log.Printf("[QQ bot] Failed to send identify: %v", err)
		}

	case OPDispatch:
		// Update sequence number
		if payload.S != nil {
			q.lastSequence = payload.S
		}

		// Handle event types
		switch payload.T {
		case "READY":
			log.Printf("[QQ bot] Gateway READY")
		case "C2C_MESSAGE_CREATE":
			q.handleC2CMessage(payload.D)
		}

	case OPHeartbeatAck:
		// Heartbeat acknowledged, nothing to do
	case OPReconnect:
		log.Printf("[QQ bot] Server requested reconnection")
		q.scheduleReconnect()
	}
}

// handleC2CMessage processes C2C (private chat) messages
func (q *QQBot) handleC2CMessage(data interface{}) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Printf("[QQ bot] Failed to marshal C2C message: %v", err)
		return
	}

	var msg C2CMessageData
	if err := json.Unmarshal(jsonData, &msg); err != nil {
		log.Printf("[QQ bot] Failed to unmarshal C2C message: %v", err)
		return
	}

	// Create bot message and call handler
	botMsg := BotMessage{
		Platform:  "qq",
		UserID:    msg.Author.UserOpenID,
		Channel:   msg.Author.UserOpenID,
		Content:   msg.Content,
		Timestamp: time.Now(),
		MessageID: msg.ID,
	}

	if q.messageHandler != nil {
		q.messageHandler(botMsg)
	}
}

// scheduleReconnect schedules a reconnection attempt with exponential backoff
func (q *QQBot) scheduleReconnect() {
	// TODO: Implement exponential backoff reconnection
	// For now, just log and stop
	log.Printf("[QQ bot] Connection lost, manual restart required")
}

// Stop stops the QQ bot and cleans up resources
func (q *QQBot) Stop() error {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.cancel != nil {
		q.cancel()
	}

	if q.wsConn != nil {
		q.wsConn.Close()
		q.wsConn = nil
	}

	log.Printf("[QQ bot] Stopped")
	return nil
}
