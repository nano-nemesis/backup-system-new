package notifier

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Telegram struct {
	botToken string
	chatID   string
	client   *http.Client
}

func NewTelegram(botToken, chatID string) *Telegram {
	return &Telegram{
		botToken: strings.TrimSpace(botToken),
		chatID:   strings.TrimSpace(chatID),
		client:   &http.Client{Timeout: 15 * time.Second},
	}
}

func (t *Telegram) Enabled() bool {
	return t.botToken != "" && t.chatID != ""
}

func (t *Telegram) Send(message string) error {
	if !t.Enabled() {
		return nil
	}

	form := url.Values{}
	form.Set("chat_id", t.chatID)
	form.Set("text", message)

	endpoint := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", t.botToken)
	resp, err := t.client.PostForm(endpoint, form)
	if err != nil {
		return fmt.Errorf("telegram send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("telegram send failed with status %s", resp.Status)
	}

	return nil
}
