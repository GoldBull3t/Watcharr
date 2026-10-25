package watched

import (
	"testing"

	"github.com/sbondCo/Watcharr/database/entity"
	"github.com/sbondCo/Watcharr/internal/testutil"
)

func TestWatchTogether(t *testing.T) {
	db := testutil.SetupDB(t)
	svc := &Service{db: db}

	// 1. Create test users
	userA := entity.User{
		Username: "user_a",
		Password: "password",
	}
	userB := entity.User{
		Username: "user_b",
		Password: "password",
	}
	userPrivate := entity.User{
		Username: "user_private",
		Password: "password",
		UserSettings: entity.UserSettings{
			Private: boolPtr(true),
		},
	}
	if err := db.Create(&userA).Error; err != nil {
		t.Fatalf("failed to create userA: %v", err)
	}
	if err := db.Create(&userB).Error; err != nil {
		t.Fatalf("failed to create userB: %v", err)
	}
	if err := db.Create(&userPrivate).Error; err != nil {
		t.Fatalf("failed to create userPrivate: %v", err)
	}

	// 2. Create sample contents
	content1 := entity.Content{
		TmdbID: 101,
		Title:  "Dune: Part Two",
		Type:   entity.MOVIE,
	}
	content2 := entity.Content{
		TmdbID: 102,
		Title:  "Oppenheimer",
		Type:   entity.MOVIE,
	}
	content3 := entity.Content{
		TmdbID: 103,
		Title:  "Severance",
		Type:   entity.SHOW,
	}
	db.Create(&content1)
	db.Create(&content2)
	db.Create(&content3)

	c1ID := content1.ID
	c2ID := content2.ID
	c3ID := content3.ID

	// User A has Content 1 and Content 2 in PLANNED
	db.Create(&entity.Watched{
		UserID:    userA.ID,
		ContentID: &c1ID,
		Status:    entity.PLANNED,
	})
	db.Create(&entity.Watched{
		UserID:    userA.ID,
		ContentID: &c2ID,
		Status:    entity.PLANNED,
	})

	// User B has Content 1 in PLANNED (Match!), and Content 3 in FINISHED with rating 9
	db.Create(&entity.Watched{
		UserID:    userB.ID,
		ContentID: &c1ID,
		Status:    entity.PLANNED,
	})
	db.Create(&entity.Watched{
		UserID:    userB.ID,
		ContentID: &c3ID,
		Status:    entity.FINISHED,
		Rating:    9.0,
	})

	t.Run("GetWatchTogetherUsers lists candidate users", func(t *testing.T) {
		users, err := svc.GetWatchTogetherUsers(userA.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		foundB := false
		foundPrivate := false
		for _, u := range users {
			if u.ID == userB.ID {
				foundB = true
			}
			if u.ID == userPrivate.ID {
				foundPrivate = true
			}
		}
		if !foundB {
			t.Errorf("expected userB in WatchTogether users list")
		}
		if foundPrivate {
			t.Errorf("expected private user without follow to be excluded")
		}
	})

	t.Run("Cannot compare with self", func(t *testing.T) {
		_, err := svc.GetWatchTogether(userA.ID, userA.ID, "common")
		if err == nil {
			t.Errorf("expected error when comparing with self, got nil")
		}
	})

	t.Run("Private user blocked if not followed", func(t *testing.T) {
		_, err := svc.GetWatchTogether(userA.ID, userPrivate.ID, "common")
		if err == nil {
			t.Errorf("expected error for private user without follow, got nil")
		}
	})

	t.Run("Mode common returns mutual planned items", func(t *testing.T) {
		resp, err := svc.GetWatchTogether(userA.ID, userB.ID, "common")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.TotalCount != 1 {
			t.Fatalf("expected 1 common planned item, got %d", resp.TotalCount)
		}
		if resp.Items[0].Content.Title != "Dune: Part Two" {
			t.Errorf("expected Dune: Part Two, got %s", resp.Items[0].Content.Title)
		}
		if resp.Items[0].TargetUserStatus != entity.PLANNED {
			t.Errorf("expected target user status PLANNED, got %s", resp.Items[0].TargetUserStatus)
		}
	})

	t.Run("Mode unwatched returns items userB hasn't seen", func(t *testing.T) {
		resp, err := svc.GetWatchTogether(userA.ID, userB.ID, "unwatched")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// User A has Dune and Oppenheimer in PLANNED. User B hasn't seen either.
		if resp.TotalCount != 2 {
			t.Errorf("expected 2 unwatched items, got %d", resp.TotalCount)
		}
	})

	t.Run("Mode recommended returns userB top ratings unviewed by userA", func(t *testing.T) {
		resp, err := svc.GetWatchTogether(userA.ID, userB.ID, "recommended")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.TotalCount != 1 {
			t.Fatalf("expected 1 recommended item, got %d", resp.TotalCount)
		}
		if resp.Items[0].Content.Title != "Severance" {
			t.Errorf("expected Severance, got %s", resp.Items[0].Content.Title)
		}
		if resp.Items[0].TargetUserRating != 9.0 {
			t.Errorf("expected rating 9.0, got %f", resp.Items[0].TargetUserRating)
		}
	})
}

func boolPtr(b bool) *bool {
	return &b
}
