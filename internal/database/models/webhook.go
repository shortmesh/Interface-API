package models

import (
	"time"

	"gorm.io/gorm"
)

type Webhook struct {
	ID               uint      `json:"id"`
	MatrixIdentityID uint      `json:"matrix_identity_id"`
	URL              string    `json:"url"`
	Active           bool      `json:"active"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

func (Webhook) TableName() string {
	return "webhooks"
}

func CreateWebhook(db *gorm.DB, matrixIdentityID uint, url string) (*Webhook, error) {
	webhook := &Webhook{
		MatrixIdentityID: matrixIdentityID,
		URL:              url,
		Active:           true,
	}
	err := db.Create(webhook).Error
	return webhook, err
}

func FindWebhookByIdentityAndURL(db *gorm.DB, matrixIdentityID uint, url string) (*Webhook, error) {
	var webhook Webhook
	err := db.Where("matrix_identity_id = ? AND url = ?", matrixIdentityID, url).First(&webhook).Error
	return &webhook, err
}

func FindWebhooksByIdentity(db *gorm.DB, matrixIdentityID uint) ([]Webhook, error) {
	var webhooks []Webhook
	err := db.Where("matrix_identity_id = ?", matrixIdentityID).Find(&webhooks).Error
	return webhooks, err
}

func FindActiveWebhooksByIdentity(db *gorm.DB, matrixIdentityID uint) ([]Webhook, error) {
	var webhooks []Webhook
	err := db.Where("matrix_identity_id = ? AND active = ?", matrixIdentityID, true).Find(&webhooks).Error
	return webhooks, err
}

func FindAllActiveWebhooks(db *gorm.DB) ([]Webhook, error) {
	var webhooks []Webhook
	err := db.Where("active = ?", true).Find(&webhooks).Error
	return webhooks, err
}

func FindAllWebhooks(db *gorm.DB) ([]Webhook, error) {
	var webhooks []Webhook
	err := db.Find(&webhooks).Error
	return webhooks, err
}

func DeleteWebhook(db *gorm.DB, matrixIdentityID uint, id uint) error {
	return db.Where("id = ? AND matrix_identity_id = ?", id, matrixIdentityID).Delete(&Webhook{}).Error
}

func UpdateWebhookStatus(db *gorm.DB, matrixIdentityID uint, id uint, active bool) error {
	return db.Model(&Webhook{}).Where("id = ? AND matrix_identity_id = ?", id, matrixIdentityID).Update("active", active).Error
}

func UpdateWebhook(db *gorm.DB, matrixIdentityID uint, id uint, url *string, active *bool) (*Webhook, error) {
	var webhook Webhook
	err := db.Where("id = ? AND matrix_identity_id = ?", id, matrixIdentityID).First(&webhook).Error
	if err != nil {
		return nil, err
	}

	updates := make(map[string]any)
	if url != nil {
		updates["url"] = *url
	}
	if active != nil {
		updates["active"] = *active
	}

	if len(updates) > 0 {
		err = db.Model(&webhook).Updates(updates).Error
		if err != nil {
			return nil, err
		}
	}

	return &webhook, nil
}
