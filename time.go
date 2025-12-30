package mgo

import "time"

// TimeHook interface for auto time management
type TimeHook interface {
	SetCreatedAt(time.Time)
	SetUpdatedAt(time.Time)
}

// TimeFields is a helper struct to be embedded in user models
type TimeFields struct {
	CreatedAt time.Time `bson:"created_at"`
	UpdatedAt time.Time `bson:"updated_at"`
}

func (t *TimeFields) SetCreatedAt(now time.Time) {
	if t.CreatedAt.IsZero() {
		t.CreatedAt = now
	}
}

func (t *TimeFields) SetUpdatedAt(now time.Time) {
	t.UpdatedAt = now
}

// SoftDelete is a helper struct for soft delete
type SoftDelete struct {
	DeletedAt *time.Time `bson:"deleted_at,omitempty"`
}

// DayRange returns the start and end of the day in UTC
func DayRange(t time.Time) (time.Time, time.Time) {
	y, m, d := t.UTC().Date()
	start := time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
	end := start.Add(24 * time.Hour).Add(-1 * time.Nanosecond)
	return start, end
}
