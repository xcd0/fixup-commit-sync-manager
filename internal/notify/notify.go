package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"fixup-commit-sync-manager/internal/config"
)

type Notifier struct {
	config *config.NotifyConfig
}

type SlackMessage struct {
	Text        string       `json:"text"`
	Username    string       `json:"username,omitempty"`
	IconEmoji   string       `json:"icon_emoji,omitempty"`
	Attachments []Attachment `json:"attachments,omitempty"`
}

type Attachment struct {
	Color  string  `json:"color,omitempty"`
	Title  string  `json:"title,omitempty"`
	Text   string  `json:"text,omitempty"`
	Fields []Field `json:"fields,omitempty"`
}

type Field struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

func NewNotifier(config *config.NotifyConfig) *Notifier {
	return &Notifier{config: config}
}

func (n *Notifier) NotifyError(operation string, err error, details map[string]string) error {
	if n.config == nil || n.config.SlackWebhookURL == "" {
		return nil
	}

	message := n.createErrorMessage(operation, err, details)
	return n.sendSlackMessage(message)
}

func (n *Notifier) NotifySuccess(operation string, details map[string]string) error {
	if n.config == nil || n.config.SlackWebhookURL == "" {
		return nil
	}

	message := n.createSuccessMessage(operation, details)
	return n.sendSlackMessage(message)
}

func (n *Notifier) NotifyInfo(title, text string, details map[string]string) error {
	if n.config == nil || n.config.SlackWebhookURL == "" {
		return nil
	}

	message := n.createInfoMessage(title, text, details)
	return n.sendSlackMessage(message)
}

func (n *Notifier) createErrorMessage(operation string, err error, details map[string]string) SlackMessage {
	fields := []Field{
		{Title: "Operation", Value: operation, Short: true},
		{Title: "Error", Value: err.Error(), Short: false},
		{Title: "Timestamp", Value: time.Now().Format("2006-01-02 15:04:05"), Short: true},
	}

	for key, value := range details {
		fields = append(fields, Field{
			Title: key,
			Value: value,
			Short: true,
		})
	}

	return SlackMessage{
		Text:      fmt.Sprintf("❌ FixupCommitSyncManager Error: %s", operation),
		Username:  "FixupCommitSyncManager",
		IconEmoji: ":warning:",
		Attachments: []Attachment{
			{
				Color:  "danger",
				Title:  "Operation Failed",
				Text:   err.Error(),
				Fields: fields,
			},
		},
	}
}

func (n *Notifier) createSuccessMessage(operation string, details map[string]string) SlackMessage {
	fields := []Field{
		{Title: "Operation", Value: operation, Short: true},
		{Title: "Timestamp", Value: time.Now().Format("2006-01-02 15:04:05"), Short: true},
	}

	for key, value := range details {
		fields = append(fields, Field{
			Title: key,
			Value: value,
			Short: true,
		})
	}

	return SlackMessage{
		Text:      fmt.Sprintf("✅ FixupCommitSyncManager Success: %s", operation),
		Username:  "FixupCommitSyncManager",
		IconEmoji: ":white_check_mark:",
		Attachments: []Attachment{
			{
				Color:  "good",
				Title:  "Operation Completed",
				Fields: fields,
			},
		},
	}
}

func (n *Notifier) createInfoMessage(title, text string, details map[string]string) SlackMessage {
	fields := []Field{
		{Title: "Timestamp", Value: time.Now().Format("2006-01-02 15:04:05"), Short: true},
	}

	for key, value := range details {
		fields = append(fields, Field{
			Title: key,
			Value: value,
			Short: true,
		})
	}

	return SlackMessage{
		Text:      fmt.Sprintf("ℹ️ FixupCommitSyncManager: %s", title),
		Username:  "FixupCommitSyncManager",
		IconEmoji: ":information_source:",
		Attachments: []Attachment{
			{
				Color:  "#36a64f",
				Title:  title,
				Text:   text,
				Fields: fields,
			},
		},
	}
}

func (n *Notifier) sendSlackMessage(message SlackMessage) error {
	jsonData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal slack message: %w", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(n.config.SlackWebhookURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to send slack message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("slack webhook returned non-200 status: %d", resp.StatusCode)
	}

	return nil
}
