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

	"github.com/FacileStudio/Casier/apps/api/schemas"

	"gorm.io/gorm"
)

const (
	lookupTimeout   = 10 * time.Second
	deliveryTimeout = 10 * time.Second
)

// Event is one webhook delivery payload.
type Event struct {
	Type        string `json:"type"`
	ProjectID   int64  `json:"project_id"`
	Environment string `json:"environment"`
	Key         string `json:"key"`
	Version     int    `json:"version,omitempty"`
	UserID      string `json:"user_id"`
	Timestamp   string `json:"timestamp"`
}

// Dispatcher delivers events to the webhooks registered on a project, in the
// background, bounded by a timeout.
type Dispatcher struct {
	orm    *gorm.DB
	logger *slog.Logger
	client *http.Client
	load   func(ctx context.Context, projectID int64) ([]schemas.Webhook, error)
}

// NewDispatcher wires a webhook dispatcher onto the database.
func NewDispatcher(orm *gorm.DB, logger *slog.Logger) *Dispatcher {
	d := &Dispatcher{
		orm:    orm,
		logger: logger,
		client: &http.Client{Timeout: deliveryTimeout},
	}
	d.load = d.loadActive
	return d
}

// Fire delivers the event in the background. The caller's context is detached
// with context.WithoutCancel so its values survive for logging and tracing
// while the HTTP handler returning does not cancel the delivery.
func (d *Dispatcher) Fire(ctx context.Context, event Event) {
	go d.dispatch(context.WithoutCancel(ctx), event)
}

func (d *Dispatcher) dispatch(ctx context.Context, event Event) {
	lookupCtx, cancel := context.WithTimeout(ctx, lookupTimeout)
	defer cancel()

	webhooks, err := d.load(lookupCtx, event.ProjectID)
	if err != nil {
		d.logger.Error("failed to load webhooks",
			slog.Int64("project_id", event.ProjectID),
			slog.String("event", event.Type),
			slog.Any("error", err))
		return
	}

	payload, err := json.Marshal(event)
	if err != nil {
		d.logger.Error("failed to encode webhook payload",
			slog.Int64("project_id", event.ProjectID),
			slog.String("event", event.Type),
			slog.Any("error", err))
		return
	}

	for _, wh := range webhooks {
		if !matchesEvent(wh.Events, event.Type) {
			continue
		}
		d.send(ctx, wh, payload)
	}
}

func (d *Dispatcher) loadActive(ctx context.Context, projectID int64) ([]schemas.Webhook, error) {
	var webhooks []schemas.Webhook
	err := d.orm.WithContext(ctx).
		Where("project_id = ? AND active = ?", projectID, true).
		Find(&webhooks).Error
	return webhooks, err
}

func (d *Dispatcher) send(ctx context.Context, wh schemas.Webhook, payload []byte) {
	ctx, cancel := context.WithTimeout(ctx, deliveryTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, wh.URL, bytes.NewReader(payload))
	if err != nil {
		d.logger.Warn("webhook request failed", slog.Int64("webhook_id", wh.ID), slog.Any("error", err))
		return
	}
	req.Header.Set("Content-Type", "application/json")

	if wh.Secret != "" {
		req.Header.Set("X-Casier-Signature", Signature(wh.Secret, payload))
	}

	resp, err := d.client.Do(req)
	if err != nil {
		d.logger.Warn("webhook delivery failed", slog.Int64("webhook_id", wh.ID), slog.Any("error", err))
		return
	}
	resp.Body.Close()
	d.logger.Info("webhook delivered", slog.Int64("webhook_id", wh.ID), slog.Int("status", resp.StatusCode))
}

// Signature is the value of the X-Casier-Signature header: the hex HMAC-SHA256
// of the raw payload, prefixed with the digest name.
func Signature(secret string, payload []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
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
