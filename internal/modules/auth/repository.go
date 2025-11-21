package auth

import (
	"errors"
	"ezytix-be/internal/models"

	"gorm.io/gorm"
)

type AuthRepository interface {
	// For register validation
	FindByEmail(email string) (*models.User, error)
	FindByPhone(phone string) (*models.User, error)
	FindByUsername(username string) (*models.User, error)

	// For login
	FindByEmailOrPhone(identifier string) (*models.User, error)

	// For refresh token
	FindByID(id uint) (*models.User, error)

	// Create user
	CreateUser(user *models.User) error
}

type authRepository struct {
	db *gorm.DB
}

func NewAuthRepository(db *gorm.DB) AuthRepository {
	return &authRepository{db}
}


func (r *authRepository) FindByEmailOrPhone(identifier string) (*models.User, error) {
	var user models.User

	err := r.db.
		Where("email = ? OR phone = ?", identifier, identifier).
		First(&user).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("user tidak ditemukan")
	}

	return &user, err
}


func (r *authRepository) FindByID(id uint) (*models.User, error) {
	var user models.User

	err := r.db.First(&user, id).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("user tidak ditemukan")
	}

	return &user, err
}

func (r *authRepository) CreateUser(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *authRepository) FindByEmail(email string) (*models.User, error) {
	var user models.User

	err := r.db.Where("email = ?", email).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil // return nil, nil → artinya email BELUM dipakai
	}
	return &user, err
}

func (r *authRepository) FindByPhone(phone string) (*models.User, error) {
	var user models.User

	err := r.db.Where("phone = ?", phone).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &user, err
}

func (r *authRepository) FindByUsername(username string) (*models.User, error) {
	var user models.User

	err := r.db.Where("username = ?", username).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &user, err
}

