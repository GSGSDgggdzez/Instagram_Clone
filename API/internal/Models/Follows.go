package models

import "gorm.io/gorm"

type Follow struct {
	gorm.Model
	ID         uint `gorm:"primaryKey;autoIncrement"`
	FollowerID uint `gorm:"not null"` // Who is following
	FollowedID uint `gorm:"not null"` // Who is being followed
	IsAccepted bool `gorm:"default:false"`
	Follower   User `gorm:"foreignKey:FollowerID"`
	Followed   User `gorm:"foreignKey:FollowedID"`
}
