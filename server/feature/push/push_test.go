package push_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sbondCo/Watcharr/config"
	"github.com/sbondCo/Watcharr/database/entity"
	"github.com/sbondCo/Watcharr/feature/auth"
	"github.com/sbondCo/Watcharr/feature/push"
	"github.com/sbondCo/Watcharr/internal/testutil"
	"github.com/sbondCo/Watcharr/router"
)

func TestPushSubscriptionLifecycle(t *testing.T) {
	db := testutil.SetupDB(t)
	svc := push.NewService(db)

	// Test VAPID Key Generation
	pubKey, err := svc.GetPublicKey()
	if err != nil {
		t.Fatalf("failed to get VAPID public key: %v", err)
	}
	if len(pubKey) < 20 {
		t.Fatalf("unexpected VAPID public key length: %s", pubKey)
	}

	// Create test user
	user := entity.User{
		Username: "pushuser",
		Type:     entity.WATCHARR_USER,
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	// Test Subscribe
	subReq := push.SubscribeRequest{
		Endpoint:  "https://fcm.googleapis.com/fcm/send/test-token-12345",
		UserAgent: "Mozilla/5.0 Chrome/130.0",
	}
	subReq.Keys.P256dh = "BCVxsGwycehlpFvT"
	subReq.Keys.Auth = "8eDyScxgAAZZ"

	sub, err := svc.Subscribe(user.ID, subReq)
	if err != nil {
		t.Fatalf("failed to subscribe: %v", err)
	}
	if sub.ID == 0 || sub.UserID != user.ID {
		t.Fatalf("unexpected subscription entity: %+v", sub)
	}

	// Test GetUserSubscriptions
	subs, err := svc.GetUserSubscriptions(user.ID)
	if err != nil {
		t.Fatalf("failed to get user subscriptions: %v", err)
	}
	if len(subs) != 1 || subs[0].Endpoint != subReq.Endpoint {
		t.Fatalf("unexpected subs list: %+v", subs)
	}

	// Test Subscribe Update (same endpoint, updated key)
	subReq.Keys.P256dh = "updated-key-token"
	subUpdated, err := svc.Subscribe(user.ID, subReq)
	if err != nil {
		t.Fatalf("failed to update subscription: %v", err)
	}
	if subUpdated.ID != sub.ID || subUpdated.P256dh != "updated-key-token" {
		t.Fatalf("expected updated subscription with same ID, got: %+v", subUpdated)
	}

	// Test Unsubscribe
	err = svc.Unsubscribe(user.ID, subReq.Endpoint)
	if err != nil {
		t.Fatalf("failed to unsubscribe: %v", err)
	}

	subsAfter, err := svc.GetUserSubscriptions(user.ID)
	if err != nil {
		t.Fatalf("failed to get user subscriptions after unsubscribe: %v", err)
	}
	if len(subsAfter) != 0 {
		t.Fatalf("expected 0 subscriptions after unsubscribe, got %d", len(subsAfter))
	}

	// Regression test DB-15: Re-subscribing with the same endpoint after Unsubscribe
	subReq.Keys.P256dh = "resubscribed-key-token"
	subResub, err := svc.Subscribe(user.ID, subReq)
	if err != nil {
		t.Fatalf("failed to re-subscribe with same endpoint after unsubscribe: %v", err)
	}
	if subResub.ID == 0 || subResub.P256dh != "resubscribed-key-token" {
		t.Fatalf("unexpected re-subscribed entity: %+v", subResub)
	}
	if err := svc.Unsubscribe(user.ID, subReq.Endpoint); err != nil {
		t.Fatalf("failed second unsubscribe: %v", err)
	}

	// Test SSRF / invalid endpoint rejection (SEC-13)
	badReq := subReq
	badReq.Endpoint = "http://127.0.0.1:8080/internal"
	if _, err := svc.Subscribe(user.ID, badReq); err == nil {
		t.Fatalf("expected loopback endpoint to be rejected")
	}

	// Test SendTest with 0 subscriptions doesn't panic
	sent, err := svc.SendTest(user.ID)
	if err != nil {
		t.Fatalf("SendTest errored unexpectedly: %v", err)
	}
	if sent != 0 {
		t.Fatalf("expected 0 sent notifications, got %d", sent)
	}
}

func TestPushRouterHandlers_NoPanic(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := testutil.SetupDB(t)
	cfg := &config.ServerConfig{
		JWT_SECRET:     "test-jwt-secret-key-for-push-router",
		SIGNUP_ENABLED: true,
	}

	authSvc := auth.NewService(db, cfg, nil)
	regResp, err := authSvc.Register(&auth.UserRegisterRequest{
		Username: "routerpushuser",
		Password: "password123",
	}, entity.PERM_NONE)
	if err != nil {
		t.Fatalf("failed to register test user: %v", err)
	}

	engine := gin.New()
	api := engine.Group("/api")
	br := router.NewBaseRouter(db, api, cfg)
	svc := push.NewService(db)
	push.NewRouter(br, svc).AddRoutes()

	// 1. POST /api/push/subscribe
	subBody, _ := json.Marshal(map[string]any{
		"endpoint": "https://fcm.googleapis.com/fcm/send/router-token-999",
		"keys": map[string]string{
			"p256dh": "router-p256dh-key",
			"auth":   "router-auth-key",
		},
		"userAgent": "TestAgent/1.0",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/push/subscribe", bytes.NewReader(subBody))
	req.Header.Set("Authorization", regResp.Token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("POST /api/push/subscribe expected 200, got %d: %s", w.Code, w.Body.String())
	}

	// 2. DELETE /api/push/subscribe
	unsubBody, _ := json.Marshal(map[string]any{
		"endpoint": "https://fcm.googleapis.com/fcm/send/router-token-999",
	})
	req = httptest.NewRequest(http.MethodDelete, "/api/push/subscribe", bytes.NewReader(unsubBody))
	req.Header.Set("Authorization", regResp.Token)
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	engine.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("DELETE /api/push/subscribe expected 200, got %d: %s", w.Code, w.Body.String())
	}

	// 3. POST /api/push/subscribe again with same endpoint (DB-15 + BUG-01 via router)
	req = httptest.NewRequest(http.MethodPost, "/api/push/subscribe", bytes.NewReader(subBody))
	req.Header.Set("Authorization", regResp.Token)
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	engine.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("re-POST /api/push/subscribe expected 200, got %d: %s", w.Code, w.Body.String())
	}

	// 4. POST /api/push/test (unsubscribe first so no external network call is needed)
	req = httptest.NewRequest(http.MethodDelete, "/api/push/subscribe", bytes.NewReader(unsubBody))
	req.Header.Set("Authorization", regResp.Token)
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	req = httptest.NewRequest(http.MethodPost, "/api/push/test", nil)
	req.Header.Set("Authorization", regResp.Token)
	w = httptest.NewRecorder()
	engine.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("POST /api/push/test expected 200, got %d: %s", w.Code, w.Body.String())
	}
}
