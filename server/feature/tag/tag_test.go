package tag_test

import (
	"testing"

	"github.com/sbondCo/Watcharr/database/entity"
	"github.com/sbondCo/Watcharr/domain"
	"github.com/sbondCo/Watcharr/feature/tag"
	"github.com/sbondCo/Watcharr/internal/testutil"
)

func boolPtr(b bool) *bool {
	return &b
}

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
		IsPublic:    boolPtr(false),
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
	if created.IsPublic != false {
		t.Fatalf("expected IsPublic to be false, got true")
	}

	// 2. Get tag
	fetched, err := svc.GetTag(user.ID, created.ID)
	if err != nil {
		t.Fatalf("failed to get tag: %v", err)
	}
	if fetched.Description != "All-time favorite cinematic masterworks" {
		t.Fatalf("expected fetched description to match, got '%s'", fetched.Description)
	}

	// 3. Update tag (update name and clear/change description and make public)
	err = svc.UpdateTag(user.ID, created.ID, domain.TagAddRequest{
		Name:        "Ultimate Favorites",
		Color:       "#000000",
		BgColor:     "#36BA98",
		Description: "Updated curated list description",
		IsPublic:    boolPtr(true),
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
	if !fetchedUpdated.IsPublic {
		t.Fatalf("expected IsPublic to be true after update, got false")
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

func TestThematicList_PublicSharing(t *testing.T) {
	db := testutil.SetupDB(t)
	svc := tag.NewService(db, nil)

	user := entity.User{
		Username: "cinemaphile",
		Type:     entity.WATCHARR_USER,
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	// Create private list
	privateList, err := svc.AddTag(user.ID, domain.TagAddRequest{
		Name:        "Secret Sci-Fi",
		Color:       "#ffffff",
		BgColor:     "#121212",
		Description: "Personal watch queue",
		IsPublic:    boolPtr(false),
	})
	if err != nil {
		t.Fatalf("failed to create private list: %v", err)
	}

	// Try getting private list through public endpoint -> must fail!
	_, err = svc.GetPublicTag(privateList.ID)
	if err == nil {
		t.Fatalf("expected error accessing private list publicly, got nil")
	}
	if err.Error() != "this list is private" {
		t.Fatalf("expected 'this list is private' error, got: %v", err)
	}

	// Toggle list to public
	err = svc.ToggleVisibility(user.ID, privateList.ID, true)
	if err != nil {
		t.Fatalf("failed to toggle visibility: %v", err)
	}

	// Try getting public list -> must succeed!
	pubResp, err := svc.GetPublicTag(privateList.ID)
	if err != nil {
		t.Fatalf("expected success getting public list, got: %v", err)
	}
	if pubResp.Name != "Secret Sci-Fi" || pubResp.AuthorName != "cinemaphile" {
		t.Fatalf("unexpected public tag response: %+v", pubResp)
	}
	if !pubResp.IsPublic {
		t.Fatalf("expected pubResp.IsPublic to be true")
	}
}
