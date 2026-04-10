package repository

import "catch/internal/domain/entity"

type ConfigRepository interface {
	Load() (*entity.AppConfig, error)
	Save(config *entity.AppConfig) error
	GetConfigPath() string
	Exists() bool
}
