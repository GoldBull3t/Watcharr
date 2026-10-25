package watched

import (
	"errors"
	"log/slog"

	"github.com/sbondCo/Watcharr/database/entity"
)

type WatchTogetherUser struct {
	ID         uint          `json:"id"`
	Username   string        `json:"username"`
	Bio        string        `json:"bio,omitempty"`
	Avatar     *entity.Image `json:"avatar,omitempty"`
	IsFollowed bool          `json:"isFollowed"`
}

type WatchTogetherItem struct {
	ID               uint                 `json:"id"`
	ContentID        int                  `json:"contentId"`
	Status           entity.WatchedStatus `json:"status"`
	Rating           float64              `json:"rating"`
	Content          *entity.Content      `json:"content,omitempty"`
	TargetUserStatus entity.WatchedStatus `json:"targetUserStatus,omitempty"`
	TargetUserRating float64              `json:"targetUserRating,omitempty"`
}

type WatchTogetherResponse struct {
	TargetUser entity.PublicUser   `json:"targetUser"`
	Mode       string              `json:"mode"`
	TotalCount int                 `json:"totalCount"`
	Items      []WatchTogetherItem `json:"items"`
}

// GetWatchTogetherUsers returns other users on the instance that the current user can compare watchlists with.
// Prioritizes followed users, followed by other non-private public users.
func (s *Service) GetWatchTogetherUsers(currentUserId uint) ([]WatchTogetherUser, error) {
	// 1. Get IDs of users currentUserId follows
	var follows []entity.Follow
	if err := s.db.Where("user_id = ?", currentUserId).Find(&follows).Error; err != nil {
		slog.Error("GetWatchTogetherUsers: Failed to get follows", "error", err)
		return nil, err
	}

	followedMap := make(map[uint]bool)
	var followedIDs []uint
	for _, f := range follows {
		followedMap[f.FollowedUserID] = true
		followedIDs = append(followedIDs, f.FollowedUserID)
	}

	// 2. Fetch candidates: users that are either followed OR have public profile, excluding self
	var users []entity.User
	query := s.db.Model(&entity.User{}).
		Preload("Avatar").
		Where("id != ?", currentUserId)

	if len(followedIDs) > 0 {
		query = query.Where("private = ? OR id IN (?)", false, followedIDs)
	} else {
		query = query.Where("private = ?", false)
	}

	if err := query.Order("username ASC").Find(&users).Error; err != nil {
		slog.Error("GetWatchTogetherUsers: Failed to get users", "error", err)
		return nil, err
	}

	var results []WatchTogetherUser
	var followedUsers []WatchTogetherUser
	var otherUsers []WatchTogetherUser

	for _, u := range users {
		item := WatchTogetherUser{
			ID:         u.ID,
			Username:   u.Username,
			Bio:        u.Bio,
			IsFollowed: followedMap[u.ID],
		}
		if u.AvatarID != nil && u.Avatar.ID != 0 {
			item.Avatar = &u.Avatar
		}

		if item.IsFollowed {
			followedUsers = append(followedUsers, item)
		} else {
			otherUsers = append(otherUsers, item)
		}
	}

	// Followed users first, then remaining public users
	results = append(results, followedUsers...)
	results = append(results, otherUsers...)

	return results, nil
}

// GetWatchTogether compares the current user's library with targetUserId according to mode:
// - "common": Items that BOTH users have on their watchlist (Status == PLANNED)
// - "unwatched": Items on current user's watchlist (PLANNED) that targetUser has NOT watched (not FINISHED, not WATCHING)
// - "recommended": Items targetUser rated highly (>= 8) that current user has NOT watched yet
func (s *Service) GetWatchTogether(currentUserId, targetUserId uint, mode string) (WatchTogetherResponse, error) {
	if currentUserId == targetUserId {
		return WatchTogetherResponse{}, errors.New("cannot compare watchlist with yourself")
	}

	// 1. Verify target user
	var targetUser entity.User
	if err := s.db.Model(&entity.User{}).Preload("Avatar").Where("id = ?", targetUserId).First(&targetUser).Error; err != nil {
		slog.Error("GetWatchTogether: Target user not found", "targetUserId", targetUserId, "error", err)
		return WatchTogetherResponse{}, errors.New("target user not found")
	}

	// 2. Privacy check: if target user is private, verify current user follows them
	if targetUser.Private != nil && *targetUser.Private {
		var follow entity.Follow
		if err := s.db.Where("user_id = ? AND followed_user_id = ?", currentUserId, targetUserId).First(&follow).Error; err != nil {
			return WatchTogetherResponse{}, errors.New("this user profile is private")
		}
	}

	if mode == "" {
		mode = "common"
	}

	var items []WatchTogetherItem

	switch mode {
	case "common":
		// Both users have the item in PLANNED status
		var currentWatched []entity.Watched
		err := s.db.Model(&entity.Watched{}).
			Preload("Content").
			Where("user_id = ? AND status = ? AND content_id IN (SELECT content_id FROM watcheds WHERE user_id = ? AND status = ? AND deleted_at IS NULL)",
				currentUserId, entity.PLANNED, targetUserId, entity.PLANNED).
			Order("updated_at DESC").
			Find(&currentWatched).Error

		if err != nil {
			slog.Error("GetWatchTogether common query error", "error", err)
			return WatchTogetherResponse{}, err
		}

		// Fetch target user's entries for ratings/status
		var targetWatched []entity.Watched
		s.db.Model(&entity.Watched{}).
			Where("user_id = ? AND status = ?", targetUserId, entity.PLANNED).
			Find(&targetWatched)

		targetMap := make(map[int]entity.Watched)
		for _, tw := range targetWatched {
			if tw.ContentID != nil {
				targetMap[*tw.ContentID] = tw
			}
		}

		for _, cw := range currentWatched {
			if cw.Content == nil || cw.ContentID == nil {
				continue
			}
			tw := targetMap[*cw.ContentID]
			items = append(items, WatchTogetherItem{
				ID:               cw.ID,
				ContentID:        *cw.ContentID,
				Status:           cw.Status,
				Rating:           cw.Rating,
				Content:          cw.Content,
				TargetUserStatus: tw.Status,
				TargetUserRating: tw.Rating,
			})
		}

	case "unwatched":
		// On current user's watchlist, but target user has NOT watched or started it
		var currentWatched []entity.Watched
		err := s.db.Model(&entity.Watched{}).
			Preload("Content").
			Where("user_id = ? AND status = ? AND content_id NOT IN (SELECT content_id FROM watcheds WHERE user_id = ? AND status IN (?, ?) AND deleted_at IS NULL)",
				currentUserId, entity.PLANNED, targetUserId, entity.FINISHED, entity.WATCHING).
			Order("updated_at DESC").
			Find(&currentWatched).Error

		if err != nil {
			slog.Error("GetWatchTogether unwatched query error", "error", err)
			return WatchTogetherResponse{}, err
		}

		// Fetch any target user watched entry for metadata
		var targetWatched []entity.Watched
		s.db.Model(&entity.Watched{}).
			Where("user_id = ?", targetUserId).
			Find(&targetWatched)

		targetMap := make(map[int]entity.Watched)
		for _, tw := range targetWatched {
			if tw.ContentID != nil {
				targetMap[*tw.ContentID] = tw
			}
		}

		for _, cw := range currentWatched {
			if cw.Content == nil || cw.ContentID == nil {
				continue
			}
			var twStatus entity.WatchedStatus
			var twRating float64
			if tw, ok := targetMap[*cw.ContentID]; ok {
				twStatus = tw.Status
				twRating = tw.Rating
			}
			items = append(items, WatchTogetherItem{
				ID:               cw.ID,
				ContentID:        *cw.ContentID,
				Status:           cw.Status,
				Rating:           cw.Rating,
				Content:          cw.Content,
				TargetUserStatus: twStatus,
				TargetUserRating: twRating,
			})
		}

	case "recommended":
		// Target user loved it (FINISHED with rating >= 8), but current user hasn't watched it
		var targetWatched []entity.Watched
		err := s.db.Model(&entity.Watched{}).
			Preload("Content").
			Where("user_id = ? AND status = ? AND rating >= ? AND content_id NOT IN (SELECT content_id FROM watcheds WHERE user_id = ? AND status IN (?, ?) AND deleted_at IS NULL)",
				targetUserId, entity.FINISHED, 8.0, currentUserId, entity.FINISHED, entity.WATCHING).
			Order("rating DESC, updated_at DESC").
			Find(&targetWatched).Error

		if err != nil {
			slog.Error("GetWatchTogether recommended query error", "error", err)
			return WatchTogetherResponse{}, err
		}

		// Fetch any current user entry
		var currentWatched []entity.Watched
		s.db.Model(&entity.Watched{}).
			Where("user_id = ?", currentUserId).
			Find(&currentWatched)

		currentMap := make(map[int]entity.Watched)
		for _, cw := range currentWatched {
			if cw.ContentID != nil {
				currentMap[*cw.ContentID] = cw
			}
		}

		for _, tw := range targetWatched {
			if tw.Content == nil || tw.ContentID == nil {
				continue
			}
			var cwID uint
			var cwStatus entity.WatchedStatus
			var cwRating float64
			if cw, ok := currentMap[*tw.ContentID]; ok {
				cwID = cw.ID
				cwStatus = cw.Status
				cwRating = cw.Rating
			}
			items = append(items, WatchTogetherItem{
				ID:               cwID,
				ContentID:        *tw.ContentID,
				Status:           cwStatus,
				Rating:           cwRating,
				Content:          tw.Content,
				TargetUserStatus: tw.Status,
				TargetUserRating: tw.Rating,
			})
		}

	default:
		return WatchTogetherResponse{}, errors.New("invalid mode: must be common, unwatched, or recommended")
	}

	return WatchTogetherResponse{
		TargetUser: targetUser.GetSafe(),
		Mode:       mode,
		TotalCount: len(items),
		Items:      items,
	}, nil
}
