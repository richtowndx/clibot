package bot

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestTelegramBot_SetMessageHandler tests the SetMessageHandler method
func TestTelegramBot_SetMessageHandler(t *testing.T) {
	bot := &TelegramBot{}

	// Verify initial handler is nil
	assert.Nil(t, bot.GetMessageHandler())

	// Test setting message handler
	called := false
	handler := func(msg BotMessage) {
		called = true
	}
	bot.SetMessageHandler(handler)

	// Verify handler was set
	retrievedHandler := bot.GetMessageHandler()
	assert.NotNil(t, retrievedHandler)

	// Test that the handler works
	retrievedHandler(BotMessage{})
	assert.True(t, called, "handler should be called")

	// Test updating handler
	newCalled := false
	newHandler := func(msg BotMessage) {
		newCalled = true
	}
	bot.SetMessageHandler(newHandler)
	bot.GetMessageHandler()(BotMessage{})
	assert.True(t, newCalled, "new handler should be called")
}

// TestTelegramBot_GetMessageHandler tests the GetMessageHandler method
func TestTelegramBot_GetMessageHandler(t *testing.T) {
	bot := &TelegramBot{}

	// Test getting handler when none is set
	assert.Nil(t, bot.GetMessageHandler())

	// Test getting handler after setting one
	handler := func(msg BotMessage) {
		// Test handler
	}
	bot.SetMessageHandler(handler)
	assert.NotNil(t, bot.GetMessageHandler())
}

// TestTelegramBot_Stop tests the Stop method
func TestTelegramBot_Stop(t *testing.T) {
	t.Run("stop with nil cancel", func(t *testing.T) {
		bot := &TelegramBot{}
		err := bot.Stop()
		assert.NoError(t, err)
	})

	t.Run("stop with nil bot", func(t *testing.T) {
		bot := &TelegramBot{bot: nil}
		err := bot.Stop()
		assert.NoError(t, err)
	})
}

// TestTelegramBot_NewTelegramBot tests the NewTelegramBot constructor
func TestTelegramBot_NewTelegramBot(t *testing.T) {
	t.Run("creates bot with token", func(t *testing.T) {
		token := "test-token-123"
		bot := NewTelegramBot(token)

		assert.NotNil(t, bot)
		assert.Equal(t, token, bot.token)
	})

	t.Run("creates bot with empty token", func(t *testing.T) {
		bot := NewTelegramBot("")
		assert.NotNil(t, bot)
		assert.Equal(t, "", bot.token)
	})
}

// TestShouldSendAsFile tests the shouldSendAsFile function with various scenarios
func TestShouldSendAsFile(t *testing.T) {
	tests := []struct {
		name     string
		message  string
		expected bool
		reason   string
	}{
		// Priority 1: Length checks
		{
			name:     "message exactly at limit (4096 bytes)",
			message:  string(make([]byte, 4096)),
			expected: false,
			reason:   "exactly at limit, no special content",
		},
		{
			name:     "message exceeds 4096 bytes",
			message:  string(make([]byte, 4097)),
			expected: true,
			reason:   "exceeds length limit, send as file regardless of content",
		},
		{
			name:     "very long message (10000 bytes)",
			message:  string(make([]byte, 10000)),
			expected: true,
			reason:   "exceeds length limit, send as file",
		},

		// Priority 2: Special content detection (within length limit)
		{
			name:     "plain text message",
			message:  "Hello, this is a simple message without any special formatting.",
			expected: false,
			reason:   "no special content, send as text",
		},
		{
			name:     "message with code block",
			message:  "Here is some code:\n```go\nfunc main() {\n\tfmt.Println(\"Hello\")\n}\n```",
			expected: true,
			reason:   "contains code block, send as file",
		},
		{
			name:     "message with multiple code blocks",
			message:  "```js\nconsole.log('hello');\n```\n\n```python\nprint('world')\n```",
			expected: true,
			reason:   "contains code blocks, send as file",
		},
		{
			name:     "message with inline code (not block)",
			message:  "Use `print()` function to output text.",
			expected: false,
			reason:   "inline code is not a code block, send as text",
		},
		{
			name:     "message with markdown table",
			message:  "| Header 1 | Header 2 |\n|----------|----------|\n| Cell 1   | Cell 2   |",
			expected: true,
			reason:   "contains table, send as file",
		},
		{
			name:     "message with complex table",
			message:  "| Name | Age | City |\n|------|-----|------|\n| Alice | 30 | NYC |\n| Bob | 25 | LA |",
			expected: true,
			reason:   "contains table, send as file",
		},
		{
			name:     "message with single pipe (not table)",
			message:  "Use the | operator for bitwise OR in many programming languages.",
			expected: false,
			reason:   "single pipe is not a table, send as text",
		},
		{
			name:     "message with mermaid diagram",
			message:  "```mermaid\ngraph TD\n    A[Start] --> B[End]\n```",
			expected: true,
			reason:   "contains mermaid diagram, send as file",
		},
		{
			name:     "message with complex mermaid flowchart",
			message:  "```mermaid\nflowchart TD\n    A[Start] --> B{Decision}\n    B -->|Yes| C[Action 1]\n    B -->|No| D[Action 2]\n```",
			expected: true,
			reason:   "contains mermaid diagram, send as file",
		},
		{
			name:     "message with both code and table",
			message:  "Code:\n```\nconsole.log('test');\n```\n\nTable:\n| A | B |\n|---|---|\n| 1 | 2 |",
			expected: true,
			reason:   "contains both code and table, send as file",
		},
		{
			name:     "message with mermaid and code",
			message:  "Diagram:\n```mermaid\ngraph LR\nA-->B\n```\n\nCode:\n```js\nx = 1;\n```",
			expected: true,
			reason:   "contains mermaid and code, send as file",
		},
		{
			name:     "message with bold/italic markdown (not special)",
			message:  "This is **bold** and this is *italic* text.",
			expected: false,
			reason:   "only basic markdown, send as text",
		},
		{
			name:     "empty message",
			message:  "",
			expected: false,
			reason:   "empty message, send as text (will be filtered elsewhere)",
		},
		{
			name:     "short message with headers only",
			message:  "# Header 1\n## Header 2\n### Header 3",
			expected: false,
			reason:   "only headers, no special content, send as text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := shouldSendAsFile(tt.message)
			assert.Equal(t, tt.expected, result, tt.reason)
		})
	}
}
