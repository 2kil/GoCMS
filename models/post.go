package models

import "time"

type Post struct {
	ID            uint       `gorm:"primaryKey" json:"id"`
	Title         string     `gorm:"size:256;not null" json:"title"`
	Slug          string     `gorm:"size:256;uniqueIndex;not null" json:"slug"`
	Summary       string     `gorm:"size:512" json:"summary"`
	Content       string     `gorm:"type:text" json:"content"`
	ContentFormat string     `gorm:"size:20;default:markdown" json:"content_format"`
	CoverImage    string     `gorm:"size:512" json:"cover_image"`
	Published     bool       `gorm:"default:false" json:"published"`
	ScheduledAt   *time.Time `json:"scheduled_at"`
	ColumnID      *uint      `json:"column_id"`
	Column        *Column    `json:"column,omitempty"`
	UserID        uint       `json:"user_id"`
	User          *User      `json:"user,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}
