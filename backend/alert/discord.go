package alert

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type DiscordConfig struct {
	WebhookURL string `json:"webhook_url"`
	Username   string `json:"username,omitempty"`
	Enabled    bool   `json:"enabled"`
}

type DiscordService struct {
	client *http.Client
	config DiscordConfig
}

type DiscordWebhookPayload struct {
	Content   string         `json:"content,omitempty"`
	Username  string         `json:"username,omitempty"`
	AvatarURL string         `json:"avatar_url,omitempty"`
	Embeds    []DiscordEmbed `json:"embeds,omitempty"`
}

type DiscordEmbed struct {
	Author      *DiscordEmbedAuthor `json:"author,omitempty"`
	Title       string              `json:"title,omitempty"`
	Description string              `json:"description,omitempty"`
	Timestamp   string              `json:"timestamp,omitempty"`
	Fields      []DiscordEmbedField `json:"fields,omitempty"`
	Color       int                 `json:"color,omitempty"`
}

type DiscordEmbedField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline,omitempty"`
}

type DiscordEmbedAuthor struct {
	Name    string `json:"name"`
	URL     string `json:"url,omitempty"`
	IconURL string `json:"icon_url,omitempty"`
}

const (
	ColorSuccess = 0x00ff00 // Green
	ColorInfo    = 0x0099ff // Blue
	ColorWarning = 0xffff00 // Yellow
	ColorError   = 0xff0000 // Red
)

func NewDiscordService(config DiscordConfig) *DiscordService {
	return &DiscordService{
		config: config,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (d *DiscordService) SendMessage(ctx context.Context, msg Message) error {
	if !d.config.Enabled {
		return fmt.Errorf("discord service is disabled")
	}

	embed := d.createStructuredEmbed(msg)

	payload := DiscordWebhookPayload{
		Username: d.config.Username,
		Embeds:   []DiscordEmbed{embed},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal Discord payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, d.config.WebhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create Discord request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := d.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send Discord webhook: %w", err)
	}
	defer func(resp *http.Response) {
		_ = resp.Body.Close()
	}(resp)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("discord webhook failed with status: %d", resp.StatusCode)
	}

	return nil
}

func (d *DiscordService) createStructuredEmbed(msg Message) DiscordEmbed {
	color := d.getColorFromSeverity(msg.Severity)

	embed := DiscordEmbed{
		Color:     color,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Author: &DiscordEmbedAuthor{
			Name: msg.Severity,
		},
	}

	if msg.Title != "" {
		embed.Title = msg.Title
	}

	if msg.Content != "" {
		embed.Description = fmt.Sprintf("```\n%s\n```", msg.Content)
	}

	fields := []DiscordEmbedField{
		{
			Name:   "Time",
			Value:  time.Now().Format("15:04:05 MST"),
			Inline: true,
		},
	}

	if msg.Channel != "" {
		fields = append(fields, DiscordEmbedField{
			Name:   "Channel",
			Value:  msg.Channel,
			Inline: true,
		})
	}

	embed.Fields = fields
	return embed
}

func (d *DiscordService) getColorFromSeverity(severity string) int {
	colorMap := map[string]int{
		"error":   ColorError,
		"warning": ColorWarning,
		"info":    ColorInfo,
		"success": ColorSuccess,
	}

	if color, exists := colorMap[severity]; exists {
		return color
	}
	return ColorInfo
}

func (d *DiscordService) IsEnabled() bool {
	return d.config.Enabled && d.config.WebhookURL != ""
}

func (d *DiscordService) GetServiceName() string {
	return "discord"
}

func (d *DiscordService) UpdateConfig(config DiscordConfig) {
	d.config = config
}
