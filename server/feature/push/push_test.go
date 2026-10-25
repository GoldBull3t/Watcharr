package push_test

import (
	"testing"

	"github.com/sbondCo/Watcharr/database/entity"
	"github.com/sbondCo/Watcharr/feature/push"
	"github.com/sbondCo/Watcharr/internal/testutil"
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
		Endpoint: "https://fcm.googleapis.com/fcm/send/test-token-12345",
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

	// Test SendTest with 0 subscriptions doesn't panic
	sent, err := svc.SendTest(user.ID)
	if err != nil {
		t.Fatalf("SendTest errored unexpectedly: %v", err)
	}
	if sent != 0 {
		t.Fatalf("expected 0 sent notifications, got %d", sent)
	}
}

