package webhook

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"casier/schemas"

	"gorm.io/gorm"
)

type Event struct {
	Type        string `json:"type"`
	SpaceID     int64  `json:"space_id"`
	Environment string `json:"environment"`
	Key         string `json:"key"`
	Version     int    `json:"version,omitempty"`
	UserID      string `json:"user_id"`
	Timestamp   string `json:"timestamp"`
}

type Dispatcher struct {
	orm    *gorm.DB
	logger *slog.Logger
	client *http.Client
}

func NewDispatcher(orm *gorm.DB, logger *slog.Logger) *Dispatcher {
	return &Dispatcher{
		orm:    orm,
		logger: logger,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (d *Dispatcher) Fire(ctx context.Context, event Event) {
	go func() {
		var webhooks []schemas.Webhook
		err := d.orm.WithContext(ctx).
			Where("space_id = ? AND active = ?", event.SpaceID, true).
			Find(&webhooks).Error
		if err != nil {
			d.logger.Error("failed to load webhooks", slog.Any("error", err))
			return
		}

		payload, err := json.Marshal(event)
		if err != nil {
			return
		}

		for _, wh := range webhooks {
			if !matchesEvent(wh.Events, event.Type) {
				continue
			}
			d.send(wh, payload)
		}
	}()
}

func (d *Dispatcher) send(wh schemas.Webhook, payload []byte) {
	req, err := http.NewRequest(http.MethodPost, wh.URL, bytes.NewReader(payload))
	if err != nil {
		d.logger.Warn("webhook request failed", slog.Int64("webhook_id", wh.ID), slog.Any("error", err))
		return
	}
	req.Header.Set("Content-Type", "application/json")

	if wh.Secret != "" {
		mac := hmac.New(sha256.New, []byte(wh.Secret))
		mac.Write(payload)
		sig := hex.EncodeToString(mac.Sum(nil))
		req.Header.Set("X-Casier-Signature", "sha256="+sig)
	}

	resp, err := d.client.Do(req)
	if err != nil {
		d.logger.Warn("webhook delivery failed", slog.Int64("webhook_id", wh.ID), slog.Any("error", err))
		return
	}
	resp.Body.Close()
	d.logger.Info("webhook delivered", slog.Int64("webhook_id", wh.ID), slog.Int("status", resp.StatusCode))
}

func matchesEvent(events string, eventType string) bool {
	if events == "" || events == "*" {
		return true
	}
	for _, e := range strings.Split(events, ",") {
		if strings.TrimSpace(e) == eventType {
			return true
		}
	}
	return false
}
