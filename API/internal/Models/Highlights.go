package models

import "gorm.io/gorm"

type Highlight struct {
	gorm.Model
	ID         uint    `gorm:"primaryKey;autoIncrement"`
	UserID     uint    `gorm:"not null"`
	Title      string  `gorm:"not null;size:255"`
	CoverImage string  `gorm:"not null"`
	Stories    []Story `gorm:"many2many:highlight_stories"`
}
