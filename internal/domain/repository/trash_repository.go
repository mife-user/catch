package repository

import "catch/internal/domain/entity"

type TrashRepository interface {
	Add(item *entity.TrashItem) error
	List() ([]*entity.TrashItem, error)
	FindByOriginalPath(originalPath string) (*entity.TrashItem, error)
	Remove(originalPath string) error
	GetExpiredItems() ([]*entity.TrashItem, error)
	CleanExpired() (int, error)
}
