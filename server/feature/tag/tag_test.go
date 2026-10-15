package tag_test

import (
	"testing"

	"github.com/sbondCo/Watcharr/database/entity"
	"github.com/sbondCo/Watcharr/domain"
	"github.com/sbondCo/Watcharr/feature/tag"
	"github.com/sbondCo/Watcharr/internal/testutil"
)

func TestTagCRUD_WithDescription(t *testing.T) {
	db := testutil.SetupDB(t)
	svc := tag.NewService(db, nil)

	// Create user
	user := entity.User{
		Username: "tagtester",
		Type:     entity.WATCHARR_USER,
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	// 1. Add tag with description
	created, err := svc.AddTag(user.ID, domain.TagAddRequest{
		Name:        "Favorites",
		Color:       "#ffffff",
		BgColor:     "#ff0055",
		Description: "All-time favorite cinematic masterworks",
	})
	if err != nil {
		t.Fatalf("failed to add tag: %v", err)
	}
	if created.Name != "Favorites" {
		t.Fatalf("expected name 'Favorites', got '%s'", created.Name)
	}
	if created.Description != "All-time favorite cinematic masterworks" {
		t.Fatalf("expected description 'All-time favorite cinematic masterworks', got '%s'", created.Description)
	}

	// 2. Get tag
	fetched, err := svc.GetTag(user.ID, created.ID)
	if err != nil {
		t.Fatalf("failed to get tag: %v", err)
	}
	if fetched.Description != "All-time favorite cinematic masterworks" {
		t.Fatalf("expected fetched description to match, got '%s'", fetched.Description)
	}

	// 3. Update tag (update name and clear/change description)
	err = svc.UpdateTag(user.ID, created.ID, domain.TagAddRequest{
		Name:        "Ultimate Favorites",
		Color:       "#000000",
		BgColor:     "#36BA98",
		Description: "Updated curated list description",
	})
	if err != nil {
		t.Fatalf("failed to update tag: %v", err)
	}

	fetchedUpdated, err := svc.GetTag(user.ID, created.ID)
	if err != nil {
		t.Fatalf("failed to get updated tag: %v", err)
	}
	if fetchedUpdated.Name != "Ultimate Favorites" {
		t.Fatalf("expected name 'Ultimate Favorites', got '%s'", fetchedUpdated.Name)
	}
	if fetchedUpdated.Description != "Updated curated list description" {
		t.Fatalf("expected updated description, got '%s'", fetchedUpdated.Description)
	}

	// 4. Delete tag
	err = svc.DeleteTag(user.ID, created.ID)
	if err != nil {
		t.Fatalf("failed to delete tag: %v", err)
	}

	_, err = svc.GetTag(user.ID, created.ID)
	if err == nil {
		t.Fatalf("expected error getting deleted tag, got nil")
	}
}

