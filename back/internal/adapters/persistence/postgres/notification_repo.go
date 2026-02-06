package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"sangehassan/back/internal/domain"
)

type NotificationRepository struct {
	db *sql.DB
}

func NewNotificationRepository(db *sql.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

func (r *NotificationRepository) Create(ctx context.Context, notification domain.Notification) (domain.Notification, error) {
	if notification.Payload == nil {
		notification.Payload = map[string]any{}
	}

	payloadJSON, err := json.Marshal(notification.Payload)
	if err != nil {
		return domain.Notification{}, err
	}

	var readAt sql.NullTime
	if notification.ReadAt != nil {
		readAt = sql.NullTime{Time: *notification.ReadAt, Valid: true}
	}

	row := r.db.QueryRowContext(ctx, `
		INSERT INTO notifications (user_id, type, payload, read_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`, notification.UserID, notification.Type, payloadJSON, readAt)

	if err := row.Scan(&notification.ID, &notification.CreatedAt); err != nil {
		return domain.Notification{}, err
	}

	return notification, nil
}

func (r *NotificationRepository) ListByUser(ctx context.Context, userID string, includeRead bool) ([]domain.Notification, error) {
	query := `
		SELECT id, user_id, type, payload, read_at, created_at
		FROM notifications
		WHERE user_id = $1
	`
	if !includeRead {
		query += " AND read_at IS NULL"
	}
	query += " ORDER BY created_at DESC LIMIT 100"

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications []domain.Notification
	for rows.Next() {
		var (
			n       domain.Notification
			payload []byte
			readAt  sql.NullTime
		)
		if err := rows.Scan(&n.ID, &n.UserID, &n.Type, &payload, &readAt, &n.CreatedAt); err != nil {
			return nil, err
		}
		if len(payload) > 0 {
			if err := json.Unmarshal(payload, &n.Payload); err != nil {
				return nil, err
			}
		}
		if readAt.Valid {
			t := readAt.Time
			n.ReadAt = &t
		}
		if n.Payload == nil {
			n.Payload = map[string]any{}
		}
		notifications = append(notifications, n)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return notifications, nil
}

func (r *NotificationRepository) MarkRead(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE notifications
		SET read_at = $1
		WHERE id = $2
	`, time.Now().UTC(), id)
	return err
}
