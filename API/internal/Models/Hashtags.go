package models

import "gorm.io/gorm"

type Hashtag struct {
	gorm.Model
	ID        uint   `gorm:"primaryKey;autoIncrement"`
	Name      string `gorm:"unique;not null"`
	PostCount int    `gorm:"default:0"`
	Posts     []Post `gorm:"many2many:post_hashtags"`
}
