package auth

import (
	"errors"
	"regexp"

	"ezytix-be/internal/models"
	"ezytix-be/pkg/hash"
	"ezytix-be/pkg/jwt"
)

type AuthService interface {
	Register(req RegisterRequest) (*models.User, error)
	Login(req LoginRequest) (*LoginResponse, string, string, error)
	Refresh(refreshToken string) (*LoginResponse, string, string, error)
	GetUserByID(id uint) (*models.User, error) // NEW
    ChangePassword(userID uint, req ChangePasswordRequest) error

}

type authService struct {
	repo AuthRepository
}

func NewAuthService(repo AuthRepository) AuthService {
	return &authService{repo}
}

func (s *authService) Register(req RegisterRequest) (*models.User, error) {
	// 1. Validation: all fields required
	if req.FullName == "" || req.Username == "" || req.Email == "" || req.Phone == "" || req.Password == "" {
		return nil, errors.New("semua field harus diisi")
	}

	// 2. Username validation: regex + length
	usernameRegex := regexp.MustCompile(`^[A-Za-z0-9]{4,16}$`)
	if !usernameRegex.MatchString(req.Username) {
		return nil, errors.New("username hanya boleh huruf dan angka (4–16 karakter)")
	}

	// 3. Check unique username
	existingUsername, _ := s.repo.FindByUsername(req.Username)
	if existingUsername != nil {
		return nil, errors.New("username sudah digunakan")
	}

	// 4. Check unique email
	existingEmail, _ := s.repo.FindByEmail(req.Email)
	if existingEmail != nil {
		return nil, errors.New("email sudah digunakan")
	}

	// 5. Check unique phone
	existingPhone, _ := s.repo.FindByPhone(req.Phone)
	if existingPhone != nil {
		return nil, errors.New("phone sudah digunakan")
	}

	// 6. Hash password
	hashed, err := hash.HashPassword(req.Password)
	if err != nil {
		return nil, errors.New("gagal menghash password")
	}

	// 7. Create user model
	user := &models.User{
		FullName: req.FullName,
		Username: req.Username,
		Email:    req.Email,
		Phone:    req.Phone,
		Password: hashed,
		Role:     models.RoleCustomer,
	}

	// 8. Store to DB
	if err := s.repo.CreateUser(user); err != nil {
		return nil, err
	}

	return user, nil
}


func (s *authService) Login(req LoginRequest) (*LoginResponse, string, string, error) {
    // 1. Wajib isi email ATAU phone
    var identifier string

    if req.Email != "" {
        identifier = req.Email
    } else if req.Phone != "" {
        identifier = req.Phone
    } else {
        return nil, "", "", errors.New("email atau phone harus diisi")
    }

    // 2. Cari user berdasarkan email ATAU phone
    user, err := s.repo.FindByEmailOrPhone(identifier)
    if err != nil {
        return nil, "", "", err
    }

    // 3. Cek password
    if !hash.CheckPassword(req.Password, user.Password) {
        return nil, "", "", errors.New("password salah")
    }

    // 4. Buat access token
    access, err := jwt.CreateAccessToken(user.ID, string(user.Role), user.Email, user.Phone)
    if err != nil {
        return nil, "", "", errors.New("gagal membuat access token")
    }

    // 5. Buat refresh token
    refresh, err := jwt.CreateRefreshToken(user.ID)
    if err != nil {
        return nil, "", "", errors.New("gagal membuat refresh token")
    }

    // 6. Return response
    return &LoginResponse{
        User: user,
    }, access, refresh, nil
}



func (s *authService) Refresh(refreshToken string) (*LoginResponse, string, string, error) {
    userID, err := jwt.ValidateRefreshToken(refreshToken)
    if err != nil {
        return nil, "", "", errors.New("invalid refresh token")
    }

    user, err := s.repo.FindByID(userID)
    if err != nil {
        return nil, "", "", errors.New("user tidak ditemukan")
    }

    access, err := jwt.CreateAccessToken(user.ID, string(user.Role), user.Email, user.Phone)
    if err != nil {
        return nil, "", "", errors.New("gagal membuat access token")
    }

    refresh, err := jwt.CreateRefreshToken(user.ID)
    if err != nil {
        return nil, "", "", errors.New("gagal membuat refresh token")
    }

    return &LoginResponse{
        User: user,
    }, access, refresh, nil
}

func (s *authService) GetUserByID(id uint) (*models.User, error) {
    user, err := s.repo.FindByID(id)
    if err != nil {
        return nil, errors.New("user tidak ditemukan")
    }
    return user, nil
}

func (s *authService) ChangePassword(userID uint, req ChangePasswordRequest) error {

    // 1. Ambil user
    user, err := s.repo.FindByID(userID)
    if err != nil {
        return errors.New("user tidak ditemukan")
    }

    // 2. Validate old password
    if !hash.CheckPassword(req.OldPassword, user.Password) {
        return errors.New("password lama salah")
    }

    // 3. Validasi password baru
    //    — minimal 8 karakter
    //    — harus ada huruf (A–Z / a–z)
    //    — harus ada angka (0–9)
    if len(req.NewPassword) < 8 {
        return errors.New("password baru minimal 8 karakter")
    }

    hasLetter := regexp.MustCompile(`[A-Za-z]`).MatchString(req.NewPassword)
    hasDigit := regexp.MustCompile(`\d`).MatchString(req.NewPassword)

    if !hasLetter || !hasDigit {
        return errors.New("password baru harus mengandung huruf dan angka")
    }

    // 4. Hash new password
    hashed, err := hash.HashPassword(req.NewPassword)
    if err != nil {
        return errors.New("gagal menghash password baru")
    }

    // 5. Update password
    if err := s.repo.UpdatePassword(user.ID, hashed); err != nil {
        return errors.New("gagal update password")
    }

    return nil
}

