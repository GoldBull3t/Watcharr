package user

import (
	"errors"
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/sbondCo/Watcharr/database/entity"
	"github.com/sbondCo/Watcharr/image"
	"gorm.io/gorm"
)

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{
		db,
	}
}

// Update user settings
func (s *Service) UserUpdate(userId uint, ur entity.UserSettings) (entity.UserSettings, error) {
	slog.Debug("user update request running", "user_id", userId, "ur", ur)
	toUpdate := make(map[string]interface{})

	if ur.HideSpoilers != nil {
		toUpdate["hide_spoilers"] = *ur.HideSpoilers
	}
	if ur.Private != nil {
		toUpdate["private"] = *ur.Private
	}
	if ur.PrivateThoughts != nil {
		toUpdate["private_thoughts"] = *ur.PrivateThoughts
	}
	if ur.IncludePreviouslyWatched != nil {
		toUpdate["include_previously_watched"] = *ur.IncludePreviouslyWatched
	}
	if ur.AutomateShowStatuses != nil {
		toUpdate["automate_show_statuses"] = *ur.AutomateShowStatuses
	}
	if ur.Country != nil {
		toUpdate["country"] = *ur.Country
	}
	if ur.Language != nil {
		toUpdate["language"] = *ur.Language
	}
	if ur.UiLanguage != nil {
		toUpdate["ui_language"] = *ur.UiLanguage
	}
	if ur.Theme != nil {
		toUpdate["theme"] = *ur.Theme
	}
	if ur.RatingSystem != nil {
		toUpdate["rating_system"] = *ur.RatingSystem
	}
	if ur.RatingStep != nil {
		toUpdate["rating_step"] = *ur.RatingStep
	}
	if ur.CalendarWebhookUrl != nil {
		toUpdate["calendar_webhook_url"] = *ur.CalendarWebhookUrl
	}
	if ur.CalendarWebhookEnabled != nil {
		toUpdate["calendar_webhook_enabled"] = *ur.CalendarWebhookEnabled
	}
	if ur.CalendarWebhookFilter != nil {
		toUpdate["calendar_webhook_filter"] = *ur.CalendarWebhookFilter
	}
	if ur.ActivityWebhookUrl != nil {
		toUpdate["activity_webhook_url"] = *ur.ActivityWebhookUrl
	}
	if ur.ActivityWebhookEnabled != nil {
		toUpdate["activity_webhook_enabled"] = *ur.ActivityWebhookEnabled
	}
	if ur.ActivityWebhookFilter != nil {
		toUpdate["activity_webhook_filter"] = *ur.ActivityWebhookFilter
	}

	if len(toUpdate) > 0 {
		res := s.db.Model(&entity.User{}).Where("id = ?", userId).Updates(toUpdate)
		if res.Error != nil {
			slog.Error("user update save failed", "user_id", userId, "error", res.Error)
			return entity.UserSettings{}, errors.New("failed to save user settings")
		}
		if res.RowsAffected == 0 {
			var count int64
			s.db.Model(&entity.User{}).Where("id = ?", userId).Count(&count)
			if count == 0 {
				slog.Error("user update failed: user not found", "user_id", userId)
				return entity.UserSettings{}, errors.New("user not found")
			}
		}
	}

	return s.UserGetSettings(userId)
}

func (s *Service) UserGetSettings(userId uint) (entity.UserSettings, error) {
	slog.Debug("UserGetSettings: Request running.", "user_id", userId)
	var user entity.User
	res := s.db.Where("id = ?", userId).Take(&user)
	if res.Error != nil {
		slog.Error("user get failed", "user_id", userId, "error", res.Error)
		return entity.UserSettings{}, errors.New("failed to retrieve user")
	}
	return entity.UserSettings{
		Private:                  user.Private,
		PrivateThoughts:          user.PrivateThoughts,
		HideSpoilers:             user.HideSpoilers,
		IncludePreviouslyWatched: user.IncludePreviouslyWatched,
		AutomateShowStatuses:     user.AutomateShowStatuses,
		Country:                  user.Country,
		Language:                 user.Language,
		UiLanguage:               user.UiLanguage,
		Theme:                    user.Theme,
		RatingSystem:             user.RatingSystem,
		RatingStep:               user.RatingStep,
		CalendarWebhookUrl:       user.CalendarWebhookUrl,
		CalendarWebhookEnabled:   user.CalendarWebhookEnabled,
		CalendarWebhookFilter:    user.CalendarWebhookFilter,
		ActivityWebhookUrl:       user.ActivityWebhookUrl,
		ActivityWebhookEnabled:   user.ActivityWebhookEnabled,
		ActivityWebhookFilter:    user.ActivityWebhookFilter,
	}, nil
}

func (s *Service) UserSearch(currentUsersId uint, q string) ([]entity.PublicUser, error) {
	slog.Debug("user search request running", "query", q)
	users := new([]entity.PublicUser)
	res := s.db.Where("id != ? AND private = ? AND username LIKE ?", currentUsersId, false, "%"+q+"%").Limit(50).Table("users").Find(&users)
	if res.Error != nil {
		slog.Error("user search failed", "error", res.Error)
		return []entity.PublicUser{}, errors.New("failed to find users")
	}
	return *users, nil
}

func (s *Service) GetUserInfo(currentUsersId uint) (entity.PrivateUser, error) {
	slog.Debug("user get info request running")
	user := new(entity.PrivateUser)
	res := s.db.Where("id = ?", currentUsersId).Table("users").Preload("Avatar").Take(&user)
	if res.Error != nil {
		slog.Error("user get info failed", "error", res.Error)
		return entity.PrivateUser{}, errors.New("failed to find current user")
	}
	return *user, nil
}

// For getting a public user's info, when viewing their list for example
func (s *Service) GetUserPublicInfo(userId uint, username string) (entity.PublicUser, error) {
	slog.Debug("user get info request running")
	user := new(entity.PublicUser)
	res := s.db.Where("private = ? AND id = ? AND username = ?", false, userId, username).Table("users").Preload("Avatar").Take(&user)
	if res.Error != nil {
		slog.Error("public user get info failed", "error", res.Error)
		return entity.PublicUser{}, errors.New("failed to find user")
	}
	return *user, nil
}

func (s *Service) UserUpdateBio(userId uint, newBio string) error {
	slog.Debug("userUpdateBio request running", "user_id", userId, "newBio", newBio)
	if res := s.db.Model(&entity.User{}).Where("id = ?", userId).Update("bio", newBio); res.Error != nil {
		slog.Error("userUpdateBio failed", "user_id", userId, "error", res.Error)
		return errors.New("failed to update bio")
	}
	return nil
}

func (s *Service) UploadUserAvatar(
	c *gin.Context,
	userId uint,
) (entity.Image, error) {
	file, err := c.FormFile("avatar")
	if err != nil {
		slog.Error("failed to get file", "error", err)
		return entity.Image{}, errors.New("no file found")
	}

	slog.Debug("UploadUserAvatar: An avatar is being uploaded",
		"name", file.Filename)

	f, _ := file.Open()
	defer f.Close()

	img, err := image.
		NewSaver(s.db, "up", image.ValidateOptions{}).
		DownloadAndInsert(f)
	if err != nil {
		slog.Error("UploadUserAvatar: DownloadAndInsert failed!",
			"error", err)
		return entity.Image{}, errors.New("processing image failed")
	}

	// No need to remove old image, the daily cleanup task will handle removing
	// unused ones.

	// Update users avatar to newly inserted
	res := s.db.
		Where("id = ?", userId).
		Updates(&entity.User{AvatarID: &img.ID})
	if res.Error != nil {
		slog.Error("UploadUserAvatar: Updating the users avatar in db failed!",
			"error", err)
		return entity.Image{}, errors.New("updating user failed")
	}

	return img, nil
}

// Get a User Service by Name.
// This is only useful for services where the user can only have one connected.
func (s *Service) GetUserServiceByName(
	userID uint,
	name entity.UserServiceName,
) (entity.UserServices, error) {
	userService := new(entity.UserServices)
	err := s.db.
		Model(&entity.UserServices{}).
		Where("user_id = ? AND name = ?", userID, name).
		Take(&userService).
		Error
	if err != nil {
		slog.Error("GetUserServiceByName: Query failed!", "error", err)
		return *userService, err
	}
	return *userService, nil
}
