package push

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"sync"

	"github.com/SherClockHolmes/webpush-go"
	"github.com/sbondCo/Watcharr/database/entity"
	"gorm.io/gorm"
)

var (
	vapidMu         sync.RWMutex
	cachedPublicKey  string
	cachedPrivateKey string
)

type NotificationPayload struct {
	Title string `json:"title"`
	Body  string `json:"body"`
	Icon  string `json:"icon,omitempty"`
	Badge string `json:"badge,omitempty"`
	URL   string `json:"url,omitempty"`
	Tag   string `json:"tag,omitempty"`
}

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

// GetVAPIDKeys returns the active VAPID public and private keys, generating a persistent pair if none are set.
func (s *Service) GetVAPIDKeys() (string, string, error) {
	vapidMu.RLock()
	if cachedPublicKey != "" && cachedPrivateKey != "" {
		pub, priv := cachedPublicKey, cachedPrivateKey
		vapidMu.RUnlock()
		return pub, priv, nil
	}
	vapidMu.RUnlock()

	vapidMu.Lock()
	defer vapidMu.Unlock()

	// Check if configured via environment variables
	envPub := os.Getenv("VAPID_PUBLIC_KEY")
	envPriv := os.Getenv("VAPID_PRIVATE_KEY")
	if envPub != "" && envPriv != "" {
		cachedPublicKey = envPub
		cachedPrivateKey = envPriv
		return cachedPublicKey, cachedPrivateKey, nil
	}

	// Double check under write lock
	if cachedPublicKey != "" && cachedPrivateKey != "" {
		return cachedPublicKey, cachedPrivateKey, nil
	}

	// Generate a new VAPID keypair
	priv, pub, err := webpush.GenerateVAPIDKeys()
	if err != nil {
		slog.Error("Push: Failed to generate VAPID keys", "error", err)
		return "", "", fmt.Errorf("failed to generate VAPID keys: %w", err)
	}

	cachedPublicKey = pub
	cachedPrivateKey = priv
	slog.Info("Push: Generated new in-memory VAPID keys")
	return cachedPublicKey, cachedPrivateKey, nil
}

// GetPublicKey returns the VAPID public key needed by frontend browsers to subscribe.
func (s *Service) GetPublicKey() (string, error) {
	pub, _, err := s.GetVAPIDKeys()
	return pub, err
}

type SubscribeRequest struct {
	Endpoint  string `json:"endpoint" binding:"required"`
	Keys      struct {
		P256dh string `json:"p256dh" binding:"required"`
		Auth   string `json:"auth" binding:"required"`
	} `json:"keys" binding:"required"`
	UserAgent string `json:"userAgent,omitempty"`
}

type UnsubscribeRequest struct {
	Endpoint string `json:"endpoint" binding:"required"`
}

// Subscribe registers or updates a Web Push subscription for a user.
func (s *Service) Subscribe(userId uint, req SubscribeRequest) (*entity.PushSubscription, error) {
	if req.Endpoint == "" || req.Keys.P256dh == "" || req.Keys.Auth == "" {
		return nil, errors.New("invalid push subscription parameters")
	}

	var sub entity.PushSubscription
	res := s.db.Where("endpoint = ?", req.Endpoint).First(&sub)

	if res.Error == nil {
		// Existing subscription found: update ownership & keys
		sub.UserID = userId
		sub.P256dh = req.Keys.P256dh
		sub.Auth = req.Keys.Auth
		sub.UserAgent = req.UserAgent
		if err := s.db.Save(&sub).Error; err != nil {
			slog.Error("Push: Failed to update existing subscription", "error", err)
			return nil, err
		}
		slog.Info("Push: Updated existing subscription", "user_id", userId, "sub_id", sub.ID)
		return &sub, nil
	}

	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		sub = entity.PushSubscription{
			UserID:    userId,
			Endpoint:  req.Endpoint,
			P256dh:    req.Keys.P256dh,
			Auth:      req.Keys.Auth,
			UserAgent: req.UserAgent,
		}
		if err := s.db.Create(&sub).Error; err != nil {
			slog.Error("Push: Failed to create new subscription", "error", err)
			return nil, err
		}
		slog.Info("Push: Registered new subscription", "user_id", userId, "sub_id", sub.ID)
		return &sub, nil
	}

	return nil, res.Error
}

// Unsubscribe removes a subscription for a user matching the endpoint.
func (s *Service) Unsubscribe(userId uint, endpoint string) error {
	res := s.db.Where("user_id = ? AND endpoint = ?", userId, endpoint).Delete(&entity.PushSubscription{})
	if res.Error != nil {
		slog.Error("Push: Failed to unsubscribe", "user_id", userId, "error", res.Error)
		return res.Error
	}
	slog.Info("Push: Unsubscribed device", "user_id", userId, "rows_affected", res.RowsAffected)
	return nil
}

// GetUserSubscriptions returns all push subscriptions registered for a user.
func (s *Service) GetUserSubscriptions(userId uint) ([]entity.PushSubscription, error) {
	var subs []entity.PushSubscription
	err := s.db.Where("user_id = ?", userId).Find(&subs).Error
	return subs, err
}

// SendNotification sends a web push notification to a specific subscription.
// If the push service reports 404 or 410 (Gone / unsubscribed), the record is automatically pruned.
func (s *Service) SendNotification(sub *entity.PushSubscription, payload NotificationPayload) error {
	pub, priv, err := s.GetVAPIDKeys()
	if err != nil {
		return err
	}

	if payload.Icon == "" {
		payload.Icon = "/logo-sqre-192.png"
	}
	if payload.Badge == "" {
		payload.Badge = "/logo-sqre-144.png"
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to encode push payload: %w", err)
	}

	subject := os.Getenv("VAPID_SUBJECT")
	if subject == "" {
		subject = "mailto:admin@watcharr.app"
	}

	wpSub := &webpush.Subscription{
		Endpoint: sub.Endpoint,
		Keys: webpush.Keys{
			P256dh: sub.P256dh,
			Auth:   sub.Auth,
		},
	}

	resp, err := webpush.SendNotification(bodyBytes, wpSub, &webpush.Options{
		Subscriber:      subject,
		VAPIDPublicKey:  pub,
		VAPIDPrivateKey: priv,
		TTL:             86400, // 24 hours
		Urgency:         webpush.UrgencyHigh,
	})
	if err != nil {
		slog.Warn("Push: SendNotification failed", "sub_id", sub.ID, "error", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusGone {
		slog.Info("Push: Subscription has expired or unsubscribed, pruning from database", "sub_id", sub.ID)
		_ = s.db.Delete(sub).Error
		return fmt.Errorf("subscription expired (status %d)", resp.StatusCode)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("push service returned status %d", resp.StatusCode)
	}

	return nil
}

// SendToUser broadcasts a notification payload to all registered subscriptions for that user.
func (s *Service) SendToUser(userId uint, payload NotificationPayload) (int, error) {
	subs, err := s.GetUserSubscriptions(userId)
	if err != nil {
		return 0, err
	}
	if len(subs) == 0 {
		return 0, nil
	}

	successCount := 0
	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, sub := range subs {
		wg.Add(1)
		go func(targetSub entity.PushSubscription) {
			defer wg.Done()
			err := s.SendNotification(&targetSub, payload)
			if err == nil {
				mu.Lock()
				successCount++
				mu.Unlock()
			}
		}(sub)
	}

	wg.Wait()
	slog.Info("Push: Broadcast to user completed", "user_id", userId, "sent", successCount, "total", len(subs))
	return successCount, nil
}

// SendTest sends a test notification to the given user.
func (s *Service) SendTest(userId uint) (int, error) {
	payload := NotificationPayload{
		Title: "Watcharr Notificações",
		Body:  "As tuas notificações Web Push estão ativas e a funcionar perfeitamente!",
		Icon:  "/logo-sqre-192.png",
		Badge: "/logo-sqre-144.png",
		URL:   "/",
		Tag:   "test-notification",
	}
	return s.SendToUser(userId, payload)
}
