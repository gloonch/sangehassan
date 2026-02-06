package domain

import "time"

type Notification struct {
	ID        int64          `json:"id"`
	UserID    string         `json:"user_id"`
	Type      string         `json:"type"`
	Payload   map[string]any `json:"payload,omitempty"`
	ReadAt    *time.Time     `json:"read_at,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
}
