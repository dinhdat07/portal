package gateway

import (
	"encoding/json"
	"log"
	"net/http"
	"portal-system/internal/service"
)

type TelegramUpdate struct {
	UpdateID int              `json:"update_id"`
	Message  *TelegramMessage `json:"message,omitempty"`
}

type TelegramMessage struct {
	MessageID int          `json:"message_id"`
	From      TelegramUser `json:"from"`
	Chat      TelegramChat `json:"chat"`
	Date      int          `json:"date"`
	Text      string       `json:"text"`
}

type TelegramUser struct {
	ID        int64  `json:"id"`
	IsBot     bool   `json:"is_bot"`
	FirstName string `json:"first_name"`
	Username  string `json:"username,omitempty"`
}

type TelegramChat struct {
	ID   int64  `json:"id"`
	Type string `json:"type"`
}

func RegisterWebhookHandlers(mux *http.ServeMux, userService service.UserService) {
	if userService != nil {
		mux.HandleFunc("/api/v1/webhooks/telegram", HandleTelegramWebhook(userService))
	}
}

func HandleTelegramWebhook(userService service.UserService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var update TelegramUpdate
		if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
			log.Printf("failed to decode telegram webhook payload: %v", err)
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		// Only process messages that have text (ignore edits, channels, etc. for now)
		if update.Message != nil && update.Message.Text != "" {
			err := userService.ProcessTelegramWebhook(r.Context(), update.Message.Chat.ID, update.Message.Text)
			if err != nil {
				log.Printf("failed to process telegram webhook: %v", err)
			}
		}

		// Always respond 200 OK to Telegram so it doesn't retry
		w.WriteHeader(http.StatusOK)
	}
}
