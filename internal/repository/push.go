package repository

import "gorm.io/gorm"

type PushRepository interface {
}

type pushRepository struct {
	db *gorm.DB
}

func NewPushRepository(db *gorm.DB) PushRepository {
	return &pushRepository{
		db: db,
	}
}
