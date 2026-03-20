package bot

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/keepmind9/clibot/internal/logger"
	"github.com/keepmind9/clibot/internal/proxy"
	"github.com/keepmind9/clibot/pkg/constants"
	"github.com/sirupsen/logrus"
)

// TelegramBot implements BotAdapter interface for Telegram using long polling
type TelegramBot struct {
	DefaultTypingIndicator
	mu             sync.RWMutex
	token          string
	bot            *tgbotapi.BotAPI
	messageHandler func(BotMessage)
	ctx            context.Context
	cancel         context.CancelFunc
	proxyMgr       proxy.Manager
}

// NewTelegramBot creates a new Telegram bot instance
func NewTelegramBot(token string) *TelegramBot {
	return &TelegramBot{
		token: token,
	}
}

// SetProxyManager sets the proxy manager for the Telegram bot
func (t *TelegramBot) SetProxyManager(mgr proxy.Manager) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.proxyMgr = mgr
}

// Start establishes long polling connection to Telegram and begins listening for messages
func (t *TelegramBot) Start(messageHandler func(BotMessage)) error {
	t.SetMessageHandler(messageHandler)
	t.ctx, t.cancel = context.WithCancel(context.Background())

	logger.WithFields(logrus.Fields{
		"token": maskSecret(t.token),
	}).Info("starting-telegram-bot-with-long-polling")

	var err error
	t.mu.Lock()
	defer t.mu.Unlock()

	// Use proxy manager if available
	if t.proxyMgr != nil {
		client, clientErr := t.proxyMgr.GetHTTPClient("telegram")
		if clientErr != nil {
			logger.WithField("error", clientErr).Error("failed-to-create-proxy-client")
			return fmt.Errorf("failed to create proxy client: %w", clientErr)
		}
		t.bot, err = tgbotapi.NewBotAPIWithClient(t.token, tgbotapi.APIEndpoint, client)
	} else {
		t.bot, err = tgbotapi.NewBotAPI(t.token)
	}

	if err != nil {
		logger.WithFields(logrus.Fields{
			"error": err,
		}).Error("failed-to-initialize-telegram-bot")
		return fmt.Errorf("failed to initialize Telegram bot: %w", err)
	}

	bot := t.bot

	logger.WithFields(logrus.Fields{
		"bot_username": bot.Self.UserName,
		"bot_id":       bot.Self.ID,
	}).Info("telegram-bot-initialized-successfully")

	// Set up long polling configuration
	u := tgbotapi.NewUpdate(0)
	u.Timeout = int(constants.DefaultPollTimeout.Seconds()) // Long poll timeout in seconds

	// Start receiving updates via long polling
	updates := bot.GetUpdatesChan(u)

	// Process updates in background
	go func() {
		for {
			select {
			case <-t.ctx.Done():
				logger.Info("telegram-long-polling-stopped")
				return
			case update, ok := <-updates:
				if !ok {
					logger.Info("telegram-updates-channel-closed")
					return
				}

				if update.Message != nil {
					t.handleMessage(update.Message)
				}
			}
		}
	}()

	logger.Info("telegram-long-polling-connection-started")
	return nil
}

// handleMessage handles incoming message events from Telegram
func (t *TelegramBot) handleMessage(message *tgbotapi.Message) {
	if message == nil {
		return
	}

	// Extract message information
	var userID, chatID, content string
	var userName, firstName, lastName string

	if message.From != nil {
		userID = fmt.Sprintf("%d", message.From.ID)
		userName = message.From.UserName
		firstName = message.From.FirstName
		lastName = message.From.LastName
	}

	if message.Chat != nil {
		chatID = fmt.Sprintf("%d", message.Chat.ID)
	}

	if message.Text != "" {
		content = message.Text
	}

	// Log parsed message data
	logger.WithFields(logrus.Fields{
		"platform":    "telegram",
		"user_id":     userID,
		"username":    userName,
		"first_name":  firstName,
		"last_name":   lastName,
		"chat_id":     chatID,
		"chat_type":   message.Chat.Type,
		"message_id":  message.MessageID,
		"content":     content,
		"content_len": len(content),
	}).Info("received-telegram-message-parsed")

	// Only process text messages
	if message.Text != "" {
		// Call the handler with BotMessage
		handler := t.GetMessageHandler()
		if handler != nil {
			handler(BotMessage{
				Platform:  "telegram",
				UserID:    userID,
				Channel:   chatID,
				Content:   content,
				Timestamp: time.Now(),
			})
		}
	}
}

// SendMessage sends a message to a Telegram chat.
// Decision logic:
// 1. If message length > 4096 bytes, send as file attachment
// 2. If message length <= 4096 bytes:
//   - Send as file if contains code blocks, tables, or Mermaid diagrams
//   - Otherwise send as plain text message
func (t *TelegramBot) SendMessage(chatID, message string) error {
	t.mu.RLock()
	bot := t.bot
	t.mu.RUnlock()

	if bot == nil {
		return fmt.Errorf("telegram bot not initialized")
	}

	if chatID == "" {
		return fmt.Errorf("chat ID is required for Telegram")
	}

	// Parse chat ID (convert string to int64)
	var chatIDInt int64
	if _, err := fmt.Sscanf(chatID, "%d", &chatIDInt); err != nil {
		return fmt.Errorf("invalid chat ID format: %w", err)
	}

	// Decide whether to send as file or text message
	if shouldSendAsFile(message) {
		return t.sendFile(chatID, chatIDInt, message)
	}

	// Send as plain text message
	return t.sendTextMessage(chatID, chatIDInt, message)
}

// sendFile sends a message as a markdown file attachment
func (t *TelegramBot) sendFile(chatID string, chatIDInt int64, message string) error {
	t.mu.RLock()
	bot := t.bot
	t.mu.RUnlock()

	// Create a temporary markdown file
	tmpDir := os.TempDir()
	timestamp := time.Now().Format("20060102-150405")
	fileName := fmt.Sprintf("clibot-response-%s.md", timestamp)
	filePath := filepath.Join(tmpDir, fileName)

	// Write message content to the markdown file
	if err := os.WriteFile(filePath, []byte(message), 0644); err != nil {
		logger.WithField("error", err).Error("failed-to-create-temp-markdown-file")
		return fmt.Errorf("failed to create temp markdown file: %w", err)
	}

	// Ensure cleanup of temporary file
	defer func() {
		if removeErr := os.Remove(filePath); removeErr != nil {
			logger.WithField("error", removeErr).Warn("failed-to-remove-temp-file")
		}
	}()

	// Open file for reading
	fileReader, err := os.Open(filePath)
	if err != nil {
		logger.WithField("error", err).Error("failed-to-open-temp-file")
		return fmt.Errorf("failed to open temp file: %w", err)
	}
	defer fileReader.Close()

	// Create file document for sending with FileReader to set custom filename
	file := tgbotapi.NewDocument(chatIDInt, tgbotapi.FileReader{
		Name:   fileName,
		Reader: fileReader,
	})
	// Set caption to provide context
	file.Caption = "📄 Response as markdown file"

	// Send the document
	_, err = bot.Send(file)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"chat_id":   chatID,
			"file_path": filePath,
			"error":     err,
		}).Error("failed-to-send-markdown-file-to-telegram")
		return fmt.Errorf("failed to send markdown file to chat %s: %w", chatID, err)
	}

	logger.WithFields(logrus.Fields{
		"chat_id":      chatID,
		"file_name":    fileName,
		"content_size": len(message),
	}).Info("markdown-file-sent-to-telegram")
	return nil
}

// sendTextMessage sends a message as a plain text message (no Markdown parsing)
func (t *TelegramBot) sendTextMessage(chatID string, chatIDInt int64, message string) error {
	t.mu.RLock()
	bot := t.bot
	t.mu.RUnlock()

	// Create text message without parse mode to avoid formatting errors
	msg := tgbotapi.NewMessage(chatIDInt, message)

	// Send the message
	_, err := bot.Send(msg)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"chat_id":      chatID,
			"content_size": len(message),
			"error":        err,
		}).Error("failed-to-send-text-message-to-telegram")
		return fmt.Errorf("failed to send text message to chat %s: %w", chatID, err)
	}

	logger.WithFields(logrus.Fields{
		"chat_id":      chatID,
		"content_size": len(message),
	}).Info("text-message-sent-to-telegram")
	return nil
}

// Telegram Bot constants
const (
	// telegramMaxMessageLength is the maximum message length for Telegram (4096 bytes)
	telegramMaxMessageLength = 4096

	// telegramParseMode is the parse mode for formatted text
	telegramParseMode = "Markdown"
)

// Regular expressions for detecting special Markdown content
var (
	// Detects code blocks: ```language ... ```
	regexCodeBlock = regexp.MustCompile("```[a-zA-Z]*\\n[\\s\\S]*?```")

	// Detects Markdown tables: | ... | ... |
	regexTable = regexp.MustCompile(`\\|[^\\n]+\\|`)

	// Detects Mermaid diagrams: ```mermaid ... ```
	regexMermaid = regexp.MustCompile("```mermaid\\n[\\s\\S]*?```")
)

// shouldSendAsFile determines if a message should be sent as a file attachment.
// Priority:
// 1. If message length > 4096 bytes, send as file (no content check needed)
// 2. If message length <= 4096 bytes, check for special content:
//   - Code blocks (```)
//   - Markdown tables (|)
//   - Mermaid diagrams (```mermaid)
//
// 3. If special content found, send as file; otherwise send as text
func shouldSendAsFile(message string) bool {
	// Priority 1: Check length first (4096 bytes limit for Telegram)
	if len(message) > telegramMaxMessageLength {
		logger.WithFields(logrus.Fields{
			"length": len(message),
			"limit":  telegramMaxMessageLength,
		}).Debug("message-exceeds-length-limit-sending-as-file")
		return true
	}

	// Priority 2: For shorter messages, check for special Markdown content
	hasCodeBlock := regexCodeBlock.MatchString(message)
	hasTable := regexTable.MatchString(message) && strings.Count(message, "|") >= 4
	hasMermaid := regexMermaid.MatchString(message)

	shouldSend := hasCodeBlock || hasTable || hasMermaid

	if shouldSend {
		logger.WithFields(logrus.Fields{
			"length":         len(message),
			"has_code_block": hasCodeBlock,
			"has_table":      hasTable,
			"has_mermaid":    hasMermaid,
		}).Debug("message-contains-special-markdown-content-sending-as-file")
	}

	return shouldSend
}

// Stop closes the Telegram long polling connection and cleans up resources
func (t *TelegramBot) Stop() error {
	if t.cancel != nil {
		t.cancel()
	}

	t.mu.Lock()
	bot := t.bot
	t.bot = nil
	t.mu.Unlock()

	if bot != nil {
		bot.StopReceivingUpdates()
		logger.Info("telegram-long-polling-stopped")
	}

	logger.Info("telegram-bot-stopped")
	return nil
}

// SetMessageHandler sets the message handler in a thread-safe manner
func (t *TelegramBot) SetMessageHandler(handler func(BotMessage)) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.messageHandler = handler
}

// GetMessageHandler gets the message handler in a thread-safe manner
func (t *TelegramBot) GetMessageHandler() func(BotMessage) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.messageHandler
}
