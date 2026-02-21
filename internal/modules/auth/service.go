package auth

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"regexp"
	"time"

	"ezytix-be/internal/models"
	"ezytix-be/pkg/hash"
	"ezytix-be/pkg/jwt"
	"ezytix-be/pkg/mail" // [BARU] Import mail service
)

type AuthService interface {
	Register(req RegisterRequest) (*models.User, error)
	Login(req LoginRequest) (*LoginResponse, string, string, error)
	Refresh(refreshToken string) (*LoginResponse, string, string, error)
	GetUserByID(id uint) (*models.User, error) // NEW
    ChangePassword(userID uint, req ChangePasswordRequest) error
    UpdateProfile(userID uint, req UpdateProfileRequest) (*models.User, error)

	// [BARU] Interface OTP
	VerifyOTP(req VerifyOTPRequest) (*LoginResponse, string, string, error)
	ResendOTP(req ResendOTPRequest) error
}

type authService struct {
	repo AuthRepository
	mail mail.MailService // [BARU] Tambahkan properti ini
}

func NewAuthService(repo AuthRepository, mailService mail.MailService) AuthService {
	return &authService{
		repo: repo,
		mail: mailService,
	}
}

// Helper Generator OTP 6 Digit
func generateOTPCode() string {
	const charset = "0123456789"
	b := make([]byte, 6)
	for i := range b {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		b[i] = charset[n.Int64()]
	}
	return string(b)
}

func (s *authService) Register(req RegisterRequest) (*models.User, error) {
    // Validasi dasar (Tetap sama seperti aslimu)
	if req.FullName == "" || req.Username == "" || req.Email == "" || req.Phone == "" || req.Password == "" {
		return nil, errors.New("semua field harus diisi")
	}

	usernameRegex := regexp.MustCompile(`^[A-Za-z0-9]{4,16}$`)
	if !usernameRegex.MatchString(req.Username) {
		return nil, errors.New("username hanya boleh huruf dan angka (4–16 karakter)")
	}

	existingUsername, _ := s.repo.FindByUsername(req.Username)
	if existingUsername != nil {
		return nil, errors.New("username sudah digunakan")
	}

	existingEmail, _ := s.repo.FindByEmail(req.Email)
	if existingEmail != nil {
		return nil, errors.New("email sudah digunakan")
	}

	existingPhone, _ := s.repo.FindByPhone(req.Phone)
	if existingPhone != nil {
		return nil, errors.New("phone sudah digunakan")
	}

	hashed, err := hash.HashPassword(req.Password)
	if err != nil {
		return nil, errors.New("gagal menghash password")
	}

	user := &models.User{
		FullName:   req.FullName,
		Username:   req.Username,
		Email:      req.Email,
		Phone:      req.Phone,
		Password:   hashed,
		Role:       models.RoleCustomer,
		IsVerified: false, // Default belum verifikasi
	}

	if err := s.repo.CreateUser(user); err != nil {
		return nil, err
	}

	// ==========================================
	// [BARU] LOGIC CREATE & SEND OTP SETELAH REGISTER
	// ==========================================
	otpCode := generateOTPCode()
	otpData := &models.UserOTP{
		UserID:    user.ID,
		OTPCode:   otpCode,
		ExpiredAt: time.Now().Add(5 * time.Minute),
	}

	if err := s.repo.CreateOrUpdateOTP(otpData); err != nil {
		return nil, fmt.Errorf("berhasil register tapi gagal membuat OTP: %v", err)
	}

	// Kirim via Email (Berjalan asinkron agar response tidak lambat)
	go s.mail.SendOTPEmail(user.Email, user.FullName, otpCode)

	return user, nil
}


func (s *authService) Login(req LoginRequest) (*LoginResponse, string, string, error) {
    var identifier string

    if req.Email != "" {
        identifier = req.Email
    } else if req.Phone != "" {
        identifier = req.Phone
    } else {
        return nil, "", "", errors.New("email atau phone harus diisi")
    }

    user, err := s.repo.FindByEmailOrPhone(identifier)
    if err != nil {
        return nil, "", "", err
    }

    if !hash.CheckPassword(req.Password, user.Password) {
        return nil, "", "", errors.New("password salah")
    }

	if !user.IsVerified {
		return nil, "", "", errors.New("akun belum diverifikasi. silakan cek email Anda untuk memasukkan kode OTP")
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

// ==========================================
// [BARU] FUNGSI VERIFY OTP
// ==========================================
func (s *authService) VerifyOTP(req VerifyOTPRequest) (*LoginResponse, string, string, error) {
	user, err := s.repo.FindByEmail(req.Email)
	if err != nil || user == nil {
		return nil, "", "", errors.New("user tidak ditemukan")
	}

	if user.IsVerified {
		return nil, "", "", errors.New("akun ini sudah terverifikasi")
	}

	otp, err := s.repo.FindOTPByUserID(user.ID)
	if err != nil {
		return nil, "", "", errors.New("kode OTP tidak ditemukan atau sudah dihapus")
	}

	if otp.OTPCode != req.OTPCode {
		return nil, "", "", errors.New("kode OTP salah")
	}

	if time.Now().After(otp.ExpiredAt) {
		return nil, "", "", errors.New("kode OTP sudah kadaluarsa. silakan minta kirim ulang")
	}

	// Validasi Sukses: Ubah is_verified jadi true & hapus OTP
	user.IsVerified = true
	if err := s.repo.UpdateUser(user); err != nil {
		return nil, "", "", errors.New("gagal memverifikasi akun")
	}
	s.repo.DeleteOTP(user.ID) // Hapus OTP bekas

	// Langsung Buatkan Token Login
	access, _ := jwt.CreateAccessToken(user.ID, string(user.Role), user.Email, user.Phone)
	refresh, _ := jwt.CreateRefreshToken(user.ID)

	return &LoginResponse{User: user}, access, refresh, nil
}


// ==========================================
// [BARU] FUNGSI RESEND OTP
// ==========================================
func (s *authService) ResendOTP(req ResendOTPRequest) error {
	user, err := s.repo.FindByEmail(req.Email)
	if err != nil || user == nil {
		return errors.New("user tidak ditemukan")
	}

	if user.IsVerified {
		return errors.New("akun ini sudah terverifikasi")
	}

	// Generate OTP Baru & Geser expired 5 menit lagi
	newOTP := generateOTPCode()
	otpData := &models.UserOTP{
		UserID:    user.ID,
		OTPCode:   newOTP,
		ExpiredAt: time.Now().Add(5 * time.Minute),
	}

	if err := s.repo.CreateOrUpdateOTP(otpData); err != nil {
		return errors.New("gagal membuat OTP baru")
	}

	// Kirim Email
	go s.mail.SendOTPEmail(user.Email, user.FullName, newOTP)

	return nil
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
    user, err := s.repo.FindByID(userID)
    if err != nil {
        return errors.New("user tidak ditemukan")
    }

    if !hash.CheckPassword(req.OldPassword, user.Password) {
        return errors.New("password lama salah")
    }

    if len(req.NewPassword) < 8 {
        return errors.New("password baru minimal 8 karakter")
    }

    hasLetter := regexp.MustCompile(`[A-Za-z]`).MatchString(req.NewPassword)
    hasDigit := regexp.MustCompile(`\d`).MatchString(req.NewPassword)

    if !hasLetter || !hasDigit {
        return errors.New("password baru harus mengandung huruf dan angka")
    }

    hashed, err := hash.HashPassword(req.NewPassword)
    if err != nil {
        return errors.New("gagal menghash password baru")
    }

    if err := s.repo.UpdatePassword(user.ID, hashed); err != nil {
        return errors.New("gagal update password")
    }

    return nil
}

func (s *authService) UpdateProfile(userID uint, req UpdateProfileRequest) (*models.User, error) {
	// 1. Ambil user saat ini
	user, err := s.repo.FindByID(userID)
	if err != nil {
		return nil, errors.New("user tidak ditemukan")
	}

	// 2. Validasi Username jika berubah
	if req.Username != user.Username {
		existingUser, _ := s.repo.FindByUsername(req.Username)
		if existingUser != nil {
			return nil, errors.New("username sudah digunakan oleh orang lain")
		}
	}

	// 3. Validasi Email jika berubah
	if req.Email != user.Email {
		existingEmail, _ := s.repo.FindByEmail(req.Email)
		if existingEmail != nil {
			return nil, errors.New("email sudah digunakan oleh orang lain")
		}
	}

	// 4. Validasi Phone jika berubah
	if req.Phone != user.Phone {
		existingPhone, _ := s.repo.FindByPhone(req.Phone)
		if existingPhone != nil {
			return nil, errors.New("nomor telepon sudah digunakan oleh orang lain")
		}
	}

	// 5. Terapkan perubahan
	user.FullName = req.FullName
	user.Username = req.Username
	user.Email = req.Email
	user.Phone = req.Phone

	// 6. Simpan ke database
	if err := s.repo.UpdateUser(user); err != nil {
		return nil, errors.New("gagal memperbarui profil")
	}

	return user, nil
}