package repository

import (
	"github.com/whotterre/push_microservice/internal/models"
	"gorm.io/gorm"
)

type PushRepository interface {
	GetActiveDevicesByUserID(userID string) ([]models.UserDevice, error)
	GetDeviceByPlayerID(playerID string) (*models.UserDevice, error)
	CreateDevice(device *models.UserDevice) error
	UpdateDevice(device *models.UserDevice) error
}

type pushRepository struct {
	db *gorm.DB
}

func NewPushRepository(db *gorm.DB) PushRepository {
	return &pushRepository{
		db: db,
	}
}

// GetActiveDevicesByUserID retrieves all active devices for a given user
func (r *pushRepository) GetActiveDevicesByUserID(userID string) ([]models.UserDevice, error) {
	var devices []models.UserDevice
	err := r.db.Where("user_id = ? AND is_active = ?", userID, true).Find(&devices).Error
	return devices, err
}

// GetDeviceByPlayerID retrieves a device by its OneSignal Player ID
func (r *pushRepository) GetDeviceByPlayerID(playerID string) (*models.UserDevice, error) {
	var device models.UserDevice
	err := r.db.Where("player_id = ?", playerID).First(&device).Error
	if err != nil {
		return nil, err
	}
	return &device, nil
}

// CreateDevice creates a new device record
func (r *pushRepository) CreateDevice(device *models.UserDevice) error {
	return r.db.Create(device).Error
}

// UpdateDevice updates an existing device record
func (r *pushRepository) UpdateDevice(device *models.UserDevice) error {
	return r.db.Save(device).Error
}
