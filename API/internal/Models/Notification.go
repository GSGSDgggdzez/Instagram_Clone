package models

import "time"

type NotificationType string

const (
	NotifTypeLike          NotificationType = "like"
	NotifTypeComment       NotificationType = "comment"
	NotifTypeFollow        NotificationType = "follow"
	NotifTypeMention       NotificationType = "mention"
	NotifTypeStoryView     NotificationType = "story_view"
	NotifTypeStoryReaction NotificationType = "story_reaction"
	NotifTypePostReaction  NotificationType = "post_reaction"
	NotifTypePostComment   NotificationType = "post_comment"
	NotifTypePostShare     NotificationType = "post_share"
	NotifTypePostSave      NotificationType = "post_save"
	NotifTypePostTag       NotificationType = "post_tag"
	NotifTypePostLocation  NotificationType = "post_location"
	NotifTypePostHashtag   NotificationType = "post_hashtag"
)

type Notification struct {
	ID        uint             `gorm:"primaryKey;autoIncrement"`
	From      uint             `gorm:"not null"`
	To        uint             `gorm:"not null"`
	Type      NotificationType `gorm:"not null;size:20"`
	Context   string           `gorm:"type:text"`
	Read      bool             `gorm:"default:false"`
	Priority  int              `gorm:"default:0"`
	GroupID   string           `gorm:"index"`
	CreatedAt time.Time
	UpdatedAt time.Time
	UserID    uint `gorm:"not null"` // Add this
	User      User `gorm:"foreignKey:UserID"`
}
