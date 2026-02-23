package auth

import (
	"errors"
	"ezytix-be/internal/models"

	"gorm.io/gorm"
)

type AuthRepository interface {
	FindByEmail(email string) (*models.User, error)
	FindByPhone(phone string) (*models.User, error)
	FindByUsername(username string) (*models.User, error)
	FindByEmailOrPhone(identifier string) (*models.User, error)
	FindByID(id uint) (*models.User, error)
	CreateUser(user *models.User) error
	UpdatePassword(userID uint, hashedPassword string) error
	UpdateUser(user *models.User) error
	CreateOrUpdateOTP(otp *models.UserOTP) error
	FindOTPByUserID(userID uint) (*models.UserOTP, error)
	DeleteOTP(userID uint) error
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
		return nil, nil
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

func (r *authRepository) UpdatePassword(userID uint, hashedPassword string) error {
    return r.db.Model(&models.User{}).
        Where("id = ?", userID).
        Update("password", hashedPassword).Error
}

func (r *authRepository) UpdateUser(user *models.User) error {
	return r.db.Save(user).Error
}

func (r *authRepository) CreateOrUpdateOTP(otp *models.UserOTP) error {
	var existing models.UserOTP
	err := r.db.Where("user_id = ?", otp.UserID).First(&existing).Error
	if err == nil {
		existing.OTPCode = otp.OTPCode
		existing.ExpiredAt = otp.ExpiredAt
		return r.db.Save(&existing).Error
	}
	return r.db.Create(otp).Error
}

func (r *authRepository) FindOTPByUserID(userID uint) (*models.UserOTP, error) {
	var otp models.UserOTP
	err := r.db.Where("user_id = ?", userID).First(&otp).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("kode OTP tidak ditemukan")
	}
	return &otp, err
}

func (r *authRepository) DeleteOTP(userID uint) error {
	return r.db.Where("user_id = ?", userID).Delete(&models.UserOTP{}).Error
}