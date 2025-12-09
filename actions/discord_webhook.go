package actions

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
)

type DiscordWebHook struct {
	URLs []string
}

func (dwh DiscordWebHook) PostMessage(message string) ([]*http.Response, error) {
	// Discord only allows a max of 2000 characters in the content field.
	if len(message) > 2000 {
		return nil, fmt.Errorf("message of %d characters is larger than the max of 2000 allowed for Discord", len(message))
	}

	responses := make([]*http.Response, 0)
	body_struct := struct {
		Content string `json:"content"`
	}{
		Content: message,
	}
	body_bytes, err := json.Marshal(body_struct)
	if err != nil {
		return nil, err
	}
	body := string(body_bytes)
	client := http.Client{}

	for _, url := range dwh.URLs {
		response, err := client.Post(url, "Application/json", strings.NewReader(body))
		if err != nil {
			return responses, err
		}
		slog.Debug("Discord Webhook Response:", "Status Code", response.StatusCode, "Status", response.Status)
		if response.StatusCode < 200 || response.StatusCode >= 300 {
			response_content, err := io.ReadAll(response.Body)
			if err != nil {
				slog.Error("Error reading response body: ", "Error", err)
			}
			slog.Debug(fmt.Sprintf(`Response Content: "%s"`, response_content))
		}
	}
	return responses, err
}

func (dwh DiscordWebHook) Act(message string) error {
	_, err := dwh.PostMessage(message)
	return err
}
