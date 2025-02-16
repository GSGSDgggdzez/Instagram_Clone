package models

import (
	"gorm.io/gorm"
)

type Post struct {
	gorm.Model
	ID            uint      `gorm:"primaryKey;autoIncrement"`
	UserID        uint      `gorm:"not null"`
	User          User      `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	Caption       string    `gorm:"type:text;size:2200"` // Instagram caption limit
	MediaURLs     []string  `gorm:"type:text[]"`         // Multiple media support
	Location      string    `gorm:"size:255"`
	PostType      string    `gorm:"not null;size:20"` // photo, video, carousel
	Filter        string    `gorm:"size:50"`
	AspectRatio   float64   `gorm:"default:1.0"`
	IsArchived    bool      `gorm:"default:false"`
	IsPinned      bool      `gorm:"default:false"`
	LikesCount    int       `gorm:"default:0"`
	CommentsCount int       `gorm:"default:0"`
	Likes         []Like    `gorm:"foreignKey:PostID"`
	Comments      []Comment `gorm:"foreignKey:PostID"`
	Hashtags      []Hashtag `gorm:"many2many:post_hashtags"`
	TaggedUsers   []User    `gorm:"many2many:post_tagged_users"`
}
