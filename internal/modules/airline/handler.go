package airline

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type AirlineHandler struct {
	service AirlineService
}

func NewAirlineHandler(service AirlineService) *AirlineHandler {
	return &AirlineHandler{service}
}

// CreateAirline: Menambah maskapai baru
func (h *AirlineHandler) CreateAirline(c *fiber.Ctx) error {
	var req CreateAirlineRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	airline, err := h.service.CreateAirline(req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "airline created successfully",
		"data":    airline,
	})
}

// GetAllAirlines: Mengambil daftar semua maskapai
func (h *AirlineHandler) GetAllAirlines(c *fiber.Ctx) error {
	airlines, err := h.service.GetAllAirlines()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to fetch airlines",
		})
	}

	return c.JSON(fiber.Map{
		"data": airlines,
	})
}

// GetAirlineByID: Detail satu maskapai
func (h *AirlineHandler) GetAirlineByID(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid airline ID",
		})
	}

	airline, err := h.service.GetAirlineByID(uint(id))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "airline not found",
		})
	}

	return c.JSON(fiber.Map{
		"data": airline,
	})
}

// UpdateAirline: Edit data maskapai
func (h *AirlineHandler) UpdateAirline(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid airline ID",
		})
	}

	var req UpdateAirlineRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	airline, err := h.service.UpdateAirline(uint(id), req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "airline updated successfully",
		"data":    airline,
	})
}

// DeleteAirline: Hapus maskapai
func (h *AirlineHandler) DeleteAirline(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid airline ID",
		})
	}

	if err := h.service.DeleteAirline(uint(id)); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "airline deleted successfully",
	})
}