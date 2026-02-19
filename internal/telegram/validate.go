package telegram

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type BotInfo struct {
	OK     bool `json:"ok"`
	Result struct {
		ID        int64  `json:"id"`
		IsBot     bool   `json:"is_bot"`
		FirstName string `json:"first_name"`
		Username  string `json:"username"`
	} `json:"result"`
}

func ValidateToken(token string) (*BotInfo, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/getMe", token)
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("telegram API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("telegram API returned status %d", resp.StatusCode)
	}

	var botInfo BotInfo
	if err := json.NewDecoder(resp.Body).Decode(&botInfo); err != nil {
		return nil, fmt.Errorf("decode telegram response: %w", err)
	}

	if !botInfo.OK {
		return nil, fmt.Errorf("invalid bot token")
	}

	if !botInfo.Result.IsBot {
		return nil, fmt.Errorf("token does not belong to a bot")
	}

	return &botInfo, nil
}
