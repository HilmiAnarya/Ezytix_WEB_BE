package auth

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	jwt "ezytix-be/pkg/jwt"
)

type AuthHandler struct {
	service AuthService
}

func NewAuthHandler(db *gorm.DB) *AuthHandler {
	return &AuthHandler{
		service: NewAuthService(NewAuthRepository(db)),
	}
}

func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req RegisterRequest

	// Parse JSON request
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	// Register logic
	user, err := h.service.Register(req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Response
	return c.JSON(fiber.Map{
		"message": "register success",
		"user":    user,
	})
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req LoginRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}

	resp, access, refresh, err := h.service.Login(req)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": err.Error()})
	}

	// SET ACCESS TOKEN (HttpOnly)
	c.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    access,
		HTTPOnly: true,
		SameSite: "Strict",
		Path:     "/",
		MaxAge:   60 * 15,
	})

	// SET REFRESH TOKEN (HttpOnly)
	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    refresh,
		HTTPOnly: true,
		SameSite: "Strict",
		Path:     "/",
		MaxAge:   60 * 60 * 24 * 7, // 7 days
	})

	return c.JSON(fiber.Map{
		"message": "login success",
		"user":    resp.User,
	})
}


func (h *AuthHandler) Refresh(c *fiber.Ctx) error {
	refreshToken := c.Cookies("refresh_token")
	if refreshToken == "" {
		return c.Status(400).JSON(fiber.Map{"error": "missing refresh token"})
	}

	resp, access, refresh, err := h.service.Refresh(refreshToken)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": err.Error()})
	}

	// SET NEW TOKENS
	c.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    access,
		HTTPOnly: true,
		SameSite: "Strict",
		Path:     "/",
		MaxAge:   60 * 15,
	})

	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    refresh,
		HTTPOnly: true,
		SameSite: "Strict",
		Path:     "/",
		MaxAge:   60 * 60 * 24 * 7,
	})

	return c.JSON(fiber.Map{
		"message": "refresh success",
		"user":    resp.User,
	})
}


func (h *AuthHandler) Me(c *fiber.Ctx) error {
    claims := c.Locals("user").(*jwt.JWTClaims)

    user, err := h.service.GetUserByID(uint(claims.UserID))
    if err != nil {
        return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
            "error": "user tidak ditemukan",
        })
    }

    return c.JSON(user)
}

func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	c.Cookie(&fiber.Cookie{Name: "access_token", Value: "", MaxAge: -1, Path: "/"})
	c.Cookie(&fiber.Cookie{Name: "refresh_token", Value: "", MaxAge: -1, Path: "/"})

	return c.JSON(fiber.Map{"message": "logged out"})
}

func (h *AuthHandler) ChangePassword(c *fiber.Ctx) error {
    var req ChangePasswordRequest

    if err := c.BodyParser(&req); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
    }

    // Ambil user ID dari JWT (middleware sudah set)
    claims := c.Locals("user").(*jwt.JWTClaims)
    userID := claims.UserID

    if err := h.service.ChangePassword(userID, req); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": err.Error()})
    }

    // Auto logout (frontend redirect)
    c.Cookie(&fiber.Cookie{Name: "access_token", Value: "", MaxAge: -1, Path: "/"})
    c.Cookie(&fiber.Cookie{Name: "refresh_token", Value: "", MaxAge: -1, Path: "/"})

    return c.JSON(fiber.Map{
        "message": "password changed successfully. please login again",
    })
}
