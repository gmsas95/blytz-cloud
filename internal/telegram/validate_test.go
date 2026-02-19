package telegram

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateToken(t *testing.T) {
	tests := []struct {
		name        string
		token       string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "empty token",
			token:       "",
			expectError: true,
			errorMsg:    "telegram API",
		},
		{
			name:        "malformed token - no colon",
			token:       "malformedtoken",
			expectError: true,
			errorMsg:    "telegram API",
		},
		{
			name:        "malformed token - empty parts",
			token:       ":",
			expectError: true,
			errorMsg:    "telegram API",
		},
		{
			name:        "valid format but fake token",
			token:       "123456:FAKE_TOKEN_ABC123",
			expectError: true,
			errorMsg:    "telegram API",
		},
		{
			name:        "token with spaces",
			token:       "  123456:abc  ",
			expectError: true,
			errorMsg:    "telegram API",
		},
		{
			name:        "malformed token - no colon",
			token:       "malformedtoken",
			expectError: true,
			errorMsg:    "telegram API request failed",
		},
		{
			name:        "malformed token - empty parts",
			token:       ":",
			expectError: true,
			errorMsg:    "telegram API request failed",
		},
		{
			name:        "valid format but fake token",
			token:       "123456:FAKE_TOKEN_ABC123",
			expectError: true,
			errorMsg:    "telegram API returned status 401",
		},
		{
			name:        "token with spaces",
			token:       "  123456:abc  ",
			expectError: true,
			errorMsg:    "telegram API request failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			botInfo, err := ValidateToken(tt.token)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
				assert.Nil(t, botInfo)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, botInfo)
			}
		})
	}
}

func TestBotInfoStructure(t *testing.T) {
	// Test that BotInfo struct works correctly
	botInfo := BotInfo{
		OK: true,
		Result: struct {
			ID        int64  `json:"id"`
			IsBot     bool   `json:"is_bot"`
			FirstName string `json:"first_name"`
			Username  string `json:"username"`
		}{
			ID:        123456,
			IsBot:     true,
			FirstName: "TestBot",
			Username:  "test_bot",
		},
	}

	assert.True(t, botInfo.OK)
	assert.Equal(t, int64(123456), botInfo.Result.ID)
	assert.True(t, botInfo.Result.IsBot)
	assert.Equal(t, "TestBot", botInfo.Result.FirstName)
	assert.Equal(t, "test_bot", botInfo.Result.Username)
}

func TestValidateTokenTimeout(t *testing.T) {
	// Test with invalid token to trigger network/timeout behavior
	token := "123456789:INVALID_TOKEN_FOR_TIMEOUT_TEST"

	_, err := ValidateToken(token)

	// Should fail with API error, not timeout (in normal conditions)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "telegram API")
}
