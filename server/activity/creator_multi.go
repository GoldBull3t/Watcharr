package activity

import (
	"log/slog"

	"github.com/sbondCo/Watcharr/database/entity"
	"gorm.io/gorm"
)

// The way to add multiple activities easily.
// Build all activities, then finalize and save them by looping through each.
type MultiCreator struct {
	creators []*Creator
}

func NewMultiCreator() *MultiCreator {
	return &MultiCreator{}
}

// Add a creator.
func (mc *MultiCreator) AddCreator(c *Creator) {
	mc.creators = append(mc.creators, c)
}

// Create all added Creators one by one.
// Errors from individual Creator.Create() calls are ignored, but still logged.
func (mc *MultiCreator) CreateAll() []entity.Activity {
	// No error return type, so not adding a check for if len(mc.creators) <= 0.
	// It will return below anyways after doing nothing so no harm + this is an
	// internal func, we can trust ourselves RIGHT??? yas

	newActivities := []entity.Activity{}

	for _, v := range mc.creators {
		a, err := v.Create()
		if err != nil {
			slog.Error("CreateAll: Failed creating an activity.", "error", err)
			continue
		}
		newActivities = append(newActivities, a)
	}

	return newActivities
}

// CreateAllTx creates all added Creators within the specified transaction.
func (mc *MultiCreator) CreateAllTx(tx *gorm.DB) ([]entity.Activity, error) {
	newActivities := []entity.Activity{}
	for _, v := range mc.creators {
		if tx != nil {
			v.db = tx
		}
		a, err := v.Create()
		if err != nil {
			slog.Error("CreateAllTx: Failed creating an activity.", "error", err)
			return newActivities, err
		}
		newActivities = append(newActivities, a)
	}
	return newActivities, nil
}
