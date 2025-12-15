package airport

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type AirportHandler struct {
	service AirportService
}

func NewAirportHandler(service AirportService) *AirportHandler {
	return &AirportHandler{
		service: service,
	}
}

func (h *AirportHandler) CreateAirport(c *fiber.Ctx) error {
	var req CreateAirportRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
	}

	airport, err := h.service.CreateAirport(req)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(201).JSON(fiber.Map{
		"message": "airport created",
		"airport":    airport,
	})
}

func (h *AirportHandler) UpdateAirport(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.Atoi(idStr)

	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid airport id"})
	}

	var req UpdateAirportRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
	}

	airport, err := h.service.UpdateAirport(uint(id), req)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"message": "airport updated",
		"data":    airport,
	})
}

func (h *AirportHandler) DeleteAirport(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.Atoi(idStr)

	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid airport id"})
	}

	if err := h.service.DeleteAirport(uint(id)); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"message": "airport deleted",
	})
}

func (h *AirportHandler) GetAllAirports(c *fiber.Ctx) error {
	airports, err := h.service.GetAllAirports()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to fetch airports"})
	}

	return c.JSON(fiber.Map{
		"data": airports,
	})
}

func (h *AirportHandler) GetAirportByID(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.Atoi(idStr)

	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid airport id"})
	}

	airport, err := h.service.GetAirportByID(uint(id))
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "airport not found"})
	}

	return c.JSON(fiber.Map{
		"data": airport,
	})
}
