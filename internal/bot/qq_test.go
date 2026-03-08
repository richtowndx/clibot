package bot

import (
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewQQBot(t *testing.T) {
	bot := NewQQBot("test_app_id", "test_secret")
	assert.NotNil(t, bot)
	assert.Equal(t, "test_app_id", bot.appID)
	assert.Equal(t, "test_secret", bot.appSecret)
	assert.NotNil(t, bot.msgSeqMap)
}

func TestQQBot_SetProxyManager(t *testing.T) {
	qqBot := NewQQBot("app", "secret")
	mockMgr := &mockProxyManager{}

	qqBot.SetProxyManager(mockMgr)

	qqBot.mu.RLock()
	defer qqBot.mu.RUnlock()
	assert.Equal(t, mockMgr, qqBot.proxyMgr)
}

func TestQQBot_nextMsgSeq(t *testing.T) {
	qqBot := NewQQBot("app", "secret")

	// Test sequence increment
	seq1 := qqBot.nextMsgSeq("msg1")
	seq2 := qqBot.nextMsgSeq("msg1")
	seq3 := qqBot.nextMsgSeq("msg2")

	assert.Equal(t, 1, seq1)
	assert.Equal(t, 2, seq2)
	assert.Equal(t, 1, seq3)

	// Test map growth prevention
	qqBot.msgSeqMap = make(map[string]int)
	for i := 0; i < 600; i++ {
		qqBot.nextMsgSeq(string(rune(i)))
	}
	// Map should be pruned to prevent unbounded growth
	assert.LessOrEqual(t, len(qqBot.msgSeqMap), 500)
}

func TestQQBot_SupportsTypingIndicator(t *testing.T) {
	qqBot := NewQQBot("app", "secret")
	assert.False(t, qqBot.SupportsTypingIndicator())
}

func TestQQBot_AddTypingIndicator(t *testing.T) {
	qqBot := NewQQBot("app", "secret")
	assert.False(t, qqBot.AddTypingIndicator("test_msg_id"))
}

func TestQQBot_RemoveTypingIndicator(t *testing.T) {
	qqBot := NewQQBot("app", "secret")
	assert.NoError(t, qqBot.RemoveTypingIndicator("test_msg_id"))
}

func TestSplitMessage(t *testing.T) {
	tests := []struct {
		name              string
		message           string
		maxLen            int
		expectedPartCount int
	}{
		{
			name:              "short message",
			message:           "hello",
			maxLen:            2000,
			expectedPartCount: 1,
		},
		{
			name:              "empty string",
			message:           "",
			maxLen:            2000,
			expectedPartCount: 1,
		},
		{
			name:              "long message without newlines",
			message:           string(make([]byte, 3000)),
			maxLen:            2000,
			expectedPartCount: 2,
		},
		{
			name:              "message with newline at split boundary",
			message:           "line1\n" + strings.Repeat("a", 1995) + "\nline2\n" + strings.Repeat("b", 1000),
			maxLen:            2000,
			expectedPartCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := splitMessage(tt.message, tt.maxLen)
			assert.Equal(t, tt.expectedPartCount, len(result), "should split into correct number of parts")

			// Verify each part is within max length
			for i, part := range result {
				assert.LessOrEqual(t, len(part), tt.maxLen, "part %d should be within max length", i)
			}

			// Verify reconstruction (except for empty string case)
			if tt.message != "" {
				reconstructed := strings.Join(result, "")
				assert.Equal(t, tt.message, reconstructed, "reconstructed message should match original")
			}
		})
	}
}

// mockProxyManager is a mock implementation of proxy.Manager for testing
type mockProxyManager struct{}

func (m *mockProxyManager) GetHTTPClient(platform string) (*http.Client, error) {
	return &http.Client{}, nil
}

func (m *mockProxyManager) GetProxyURL(platform string) string {
	return ""
}
