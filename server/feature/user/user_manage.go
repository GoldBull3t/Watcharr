package user

import (
	"errors"
	"log/slog"

	"github.com/sbondCo/Watcharr/database/entity"
	"github.com/sbondCo/Watcharr/domain"
	"gorm.io/gorm"
)

type ManageService struct {
	db *gorm.DB
}

func NewManageService(db *gorm.DB) *ManageService {
	return &ManageService{
		db,
	}
}

func (s *ManageService) GetAll() ([]domain.ManagedUser, error) {
	users := []domain.ManagedUser{}
	if res := s.db.Model(&entity.User{}).Find(&users); res.Error != nil {
		slog.Error("GetAllUsers: Failed to fetch users from database", "error", res.Error)
		return []domain.ManagedUser{}, errors.New("failed to fetch users from database")
	}
	return users, nil
}

// Update a user. For management views, for admin to update another user.
func (s *ManageService) Manage(userId uint, ur domain.UpdateUserRequest) error {
	// Error now if no userId or any UpdateUserRequest property was provided.
	if userId == 0 || (ur.Permissions == nil && ur.Type == nil) {
		slog.Error("ManageUser: invalid arguments", "user_id", userId)
		return errors.New("invalid arguments, ensure a valid userId and at least one property has been provided for updating")
	}
	toUpdate := map[string]interface{}{}
	if ur.Permissions != nil {
		if *ur.Permissions == 0 {
			// If removing all perms, set to default of 1 (PERM_NONE).
			// Will avoid confusion and possibly bugs later on, though I doubt
			// we'd ever be (directly) checking a user to ensure they have no perms.
			toUpdate["permissions"] = entity.PERM_NONE
		} else {
			toUpdate["permissions"] = *ur.Permissions
		}
	}
	if ur.Type != nil {
		t := *ur.Type
		if t == entity.WATCHARR_USER || t == entity.PROXY_USER || t == entity.OIDC_USER {
			slog.Debug("ManageUser: User type is being updated.", "new_type", t)
			toUpdate["type"] = t
		} else {
			slog.Warn("ManageUser: User type will not be updated. Only watcharr/proxy/oidc types are supported for swapping.", "tried_type", t)
		}
	}
	if res := s.db.Model(&entity.User{}).Where("id = ?", userId).Updates(toUpdate); res.Error != nil {
		slog.Error("ManageUser: failed to update user in database", "user_id", userId, "error", res.Error)
		return errors.New("failed to update user in database")
	}
	slog.Debug("ManageUser: A user has been updated", "user_id", userId)
	return nil
}

// Delete permanently removes a user and their associated library data (#473).
func (s *ManageService) Delete(adminId uint, targetUserId uint) error {
	if targetUserId == 0 {
		return errors.New("invalid user id")
	}
	if adminId == targetUserId {
		return errors.New("cannot delete your own admin account")
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		var u entity.User
		if err := tx.Unscoped().Where("id = ?", targetUserId).Take(&u).Error; err != nil {
			return errors.New("user not found")
		}

		fail := func(step string, err error) error {
			slog.Error("DeleteUser: step failed", "step", step, "user_id", targetUserId, "error", err)
			return errors.New("failed to delete user data (" + step + ")")
		}

		// Remove watched<->tag associations for this user's tags and watched items.
		if err := tx.Exec(
			"DELETE FROM watched_tags WHERE tag_id IN (SELECT id FROM tags WHERE user_id = ?) OR watched_id IN (SELECT id FROM watcheds WHERE user_id = ?)",
			targetUserId, targetUserId,
		).Error; err != nil {
			return fail("watched_tags", err)
		}

		steps := []struct {
			name  string
			model any
			where string
		}{
			{"activities", &entity.Activity{}, "user_id = ?"},
			{"watched_episodes", &entity.WatchedEpisode{}, "user_id = ?"},
			{"watched_seasons", &entity.WatchedSeason{}, "user_id = ?"},
			{"watcheds", &entity.Watched{}, "user_id = ?"},
			{"tags", &entity.Tag{}, "user_id = ?"},
			{"user_services", &entity.UserServices{}, "user_id = ?"},
			{"tokens", &entity.Token{}, "user_id = ?"},
		}
		for _, st := range steps {
			if err := tx.Unscoped().Where(st.where, targetUserId).Delete(st.model).Error; err != nil {
				return fail(st.name, err)
			}
		}
		if err := tx.Unscoped().
			Where("user_id = ? OR followed_user_id = ?", targetUserId, targetUserId).
			Delete(&entity.Follow{}).Error; err != nil {
			return fail("follows", err)
		}

		if err := tx.Unscoped().Where("id = ?", targetUserId).Delete(&entity.User{}).Error; err != nil {
			slog.Error("DeleteUser: Failed to delete user record", "user_id", targetUserId, "error", err)
			return errors.New("failed to delete user")
		}
		slog.Info("DeleteUser: Admin deleted user", "admin_id", adminId, "deleted_user_id", targetUserId, "username", u.Username)
		return nil
	})
}
