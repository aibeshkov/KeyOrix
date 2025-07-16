package repository

import (
	"github.com/secretlyhq/secretly/internal/storage/models"
	"gorm.io/gorm"
)

type UserRepository interface {
	Create(user *models.User) error
	FindByUsername(username string) (*models.User, error)
	FindByID(id uint) (*models.User, error)
	List() ([]models.User, error)
	Delete(id uint) error
}

type userRepo struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepo{db}
}

// Create добавляет нового пользователя в базу
func (r *userRepo) Create(user *models.User) error {
	return r.db.Create(user).Error
}

// FindByUsername ищет пользователя по имени пользователя
func (r *userRepo) FindByUsername(username string) (*models.User, error) {
	var user models.User
	err := r.db.Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByID ищет пользователя по ID
func (r *userRepo) FindByID(id uint) (*models.User, error) {
	var user models.User
	err := r.db.First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// List возвращает всех пользователей
func (r *userRepo) List() ([]models.User, error) {
	var users []models.User
	err := r.db.Find(&users).Error
	return users, err
}

// Delete удаляет пользователя по ID
func (r *userRepo) Delete(id uint) error {
	return r.db.Delete(&models.User{}, id).Error
}
