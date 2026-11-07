package image

import (
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/sbondCo/Watcharr/config"
	"github.com/sbondCo/Watcharr/database/entity"
	"github.com/sbondCo/Watcharr/internal/testutil"
)

func TestDownloadAndInsertFromUrl(t *testing.T) {
	testutil.SetupLogging()
	db := testutil.SetupDB(t)

	i, err := NewSaver(db, "test", ValidateOptions{}).
		DownloadAndInsertFromUrl(
			"https://github.com/sbondCo/Watcharr/raw/dev/screenshot/homepage.png")
	if err != nil {
		t.Fatalf("DownloadAndInsert call failed: %v", err)
	}

	if i.ID == 0 || i.Path == "" || i.BlurHash == "" {
		t.Fatal("returned entity.Image doesn't have certain fields!",
			"id", i.ID, "path", i.Path, "blurhash", i.BlurHash)
	}

	fullImgDataPath := path.Join(config.DataPath, i.Path)

	// Verify file exists and looks right.
	if fi, err := os.Stat(fullImgDataPath); err != nil {
		t.Fatalf("os.stat failed %v", err)
	} else if fi.Size() <= 1 {
		t.Fatalf("image file size doesn't seem right: %v", fi.Size())
	} else if filepath.Ext(fi.Name()) != ".jpg" {
		t.Fatalf("image file name doesn't have .jpg ext: %s", fi.Name())
	}

	if err := os.Remove(fullImgDataPath); err != nil {
		// Not a fatal error because this is extra logic only for testing,
		// but failing test because something in the real logic might possibly
		// have something to do with it failing and we should probably know.
		t.Errorf("removing the image file errored: %v", err)
	}

	// Verify image is in db
	var c int64
	if res := db.
		Model(&entity.Image{}).
		Where(&entity.Image{ID: i.ID}).
		Count(&c); res.Error != nil {
		t.Fatalf("verification query failed: %v", res.Error)
	}
	if c != 1 {
		t.Fatalf("count of images in db doesn't look right: %v", c)
	}
}

func TestCleanupImagesPreservesGamePosterAndUserAvatar(t *testing.T) {
	testutil.SetupLogging()
	db := testutil.SetupDB(t)

	avatarImg := entity.Image{Hash: "avatar_hash", Path: "img/test/a/avatar_hash.jpg", BlurHash: "bh1"}
	gameImg := entity.Image{Hash: "game_poster_hash", Path: "img/test/g/game_poster_hash.jpg", BlurHash: "bh2"}
	orphanImg := entity.Image{Hash: "orphan_hash", Path: "img/test/o/orphan_hash.jpg", BlurHash: "bh3"}

	if err := db.Create(&avatarImg).Error; err != nil {
		t.Fatalf("failed to create avatar image: %v", err)
	}
	if err := db.Create(&gameImg).Error; err != nil {
		t.Fatalf("failed to create game image: %v", err)
	}
	if err := db.Create(&orphanImg).Error; err != nil {
		t.Fatalf("failed to create orphan image: %v", err)
	}

	user := entity.User{
		Username: "img_user",
		Password: "pw",
		AvatarID: &avatarImg.ID,
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	game := entity.Game{
		IgdbID:   999123,
		Name:     "Test Game",
		PosterID: &gameImg.ID,
	}
	if err := db.Create(&game).Error; err != nil {
		t.Fatalf("failed to create game: %v", err)
	}

	CleanupImages(db)

	var remaining []entity.Image
	if err := db.Find(&remaining).Error; err != nil {
		t.Fatalf("failed to query images after cleanup: %v", err)
	}
	if len(remaining) != 2 {
		t.Fatalf("expected 2 images remaining (avatar + game poster), got %d", len(remaining))
	}
	for _, img := range remaining {
		if img.ID == orphanImg.ID {
			t.Errorf("orphan image %d should have been cleaned up", orphanImg.ID)
		}
	}
}
