package models

import "gorm.io/gorm"

type Like struct {
	gorm.Model
	ID     uint `gorm:"primaryKey;autoIncrement"`
	UserID uint `gorm:"not null"`
	PostID uint `gorm:"not null"`
	User   User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	Post   Post `gorm:"foreignKey:PostID;constraint:OnDelete:CASCADE"`
}
