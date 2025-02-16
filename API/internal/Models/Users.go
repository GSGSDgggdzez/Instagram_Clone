package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	ID             uint        `gorm:"primaryKey;autoIncrement"`
	Username       string      `gorm:"unique;not null;size:30"` // Instagram-style username
	Name           string      `gorm:"not null;size:255"`
	Avatar         string      `gorm:"not null;size:255"`
	Bio            string      `gorm:"size:150"` // Instagram bio limit
	Website        string      `gorm:"size:255"`
	Email          string      `gorm:"unique;not null"`
	Phone          string      `gorm:"unique;not null"`
	FollowerCount  int         `gorm:"default:0"`
	FollowingCount int         `gorm:"default:0"`
	PostCount      int         `gorm:"default:0"`
	Privacy        bool        `gorm:"default:false"`
	IsVerified     bool        `gorm:"default:false"` // Blue check mark
	EmailVerified  bool        `gorm:"default:false"`
	Password       string      `gorm:"not null" json:"-"`
	Token          string      `gorm:"not null;size:255" json:"-"`
	Language       string      `gorm:"not null;size:20"`
	Posts          []Post      `gorm:"foreignKey:UserID"`
	Likes          []Like      `gorm:"foreignKey:UserID"`
	Comments       []Comment   `gorm:"foreignKey:UserID"`
	Stories        []Story     `gorm:"foreignKey:UserID"`
	SavedPosts     []Post      `gorm:"many2many:user_saved_posts"`
	Highlights     []Highlight `gorm:"foreignKey:UserID"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
