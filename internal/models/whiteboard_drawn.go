package models

import (
	"gorm.io/gorm"
)

type Drawn struct {
	gorm.Model
	Drawer       uint    `json:"drawer_user_id" gorm:"column:user_id"`
	Points       *Points `json:"points" gorm:"type:jsonb"`
	WhiteboardId uint    `json:"whiteboard_id" gorm:"not null"`
}