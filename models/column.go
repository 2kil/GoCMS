package models

import "time"

type Column struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Name         string    `gorm:"size:128;not null" json:"name"`
	Slug         string    `gorm:"size:128;uniqueIndex;not null" json:"slug"`
	IsPage       bool      `gorm:"default:false" json:"is_page"`
	PageTemplate string    `gorm:"size:255" json:"page_template"`
	Content      string    `gorm:"type:text" json:"content"`
	SortOrder    int       `gorm:"default:0" json:"sort_order"`
	ParentID     *uint     `json:"parent_id"`
	Parent       *Column   `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
	Children     []Column  `gorm:"foreignKey:ParentID" json:"children,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Posts        []Post    `gorm:"foreignKey:ColumnID" json:"posts,omitempty"`
}
