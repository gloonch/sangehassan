package domain

import "time"

type ContactSubmission struct {
	ID          int64     `json:"id"`
	FullName    string    `json:"full_name"`
	Email       string    `json:"email,omitempty"`
	CountryISO  string    `json:"country_iso,omitempty"`
	CountryCode string    `json:"country_code"`
	PhoneNumber string    `json:"phone_number"`
	PhoneE164   string    `json:"phone_e164"`
	Message     string    `json:"message"`
	Source      string    `json:"source"`
	CreatedAt   time.Time `json:"created_at"`
}
