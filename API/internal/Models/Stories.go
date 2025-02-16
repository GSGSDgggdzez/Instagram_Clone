package models

import "gorm.io/gorm"

type Story struct {
	gorm.Model
	ID        uint   `gorm:"primaryKey;autoIncrement"`
	UserID    uint   `gorm:"not null"`
	User      User   `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	MediaURL  string `gorm:"not null"`
	StoryType string `gorm:"not null"`   // photo or video
	Duration  int    `gorm:"default:24"` // hours
	ViewCount int    `gorm:"default:0"`
	IsExpired bool   `gorm:"default:false"`
	ViewedBy  []User `gorm:"many2many:story_views"`
}
