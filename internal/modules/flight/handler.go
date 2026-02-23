package flight

import (
	"ezytix-be/internal/models"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type FlightHandler struct {
	service FlightService
}

func NewFlightHandler(service FlightService) *FlightHandler {
	return &FlightHandler{service}
}

func (h *FlightHandler) CreateFlight(c *fiber.Ctx) error {
	var req CreateFlightRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	flightModel, err := h.service.CreateFlight(req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	flightResponse := ToFlightResponse(*flightModel)

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "flight created successfully",
		"data":    flightResponse,
	})
}

func (h *FlightHandler) GetAllFlights(c *fiber.Ctx) error {
	var req SearchFlightRequest
	
	if err := c.QueryParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid query params"})
	}

	var flights []models.Flight
	var err error

	if req.OriginAirportID != 0 && req.DestinationAirportID != 0 && req.DepartureDate != "" {
		flights, err = h.service.SearchFlights(req)
	} else {
		flights, err = h.service.GetAllFlights()
	}

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	var flightResponses []FlightResponse
	for _, f := range flights {
		flightResponses = append(flightResponses, ToFlightResponse(f))
	}

	if flightResponses == nil {
		flightResponses = []FlightResponse{}
	}

	return c.JSON(fiber.Map{
		"data": flightResponses,
	})
}

func (h *FlightHandler) GetFlightByID(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid flight ID",
		})
	}

	flightModel, err := h.service.GetFlightByID(uint(id))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "flight not found",
		})
	}

	flightResponse := ToFlightResponse(*flightModel)

	return c.JSON(fiber.Map{
		"data": flightResponse,
	})
}

func (h *FlightHandler) UpdateFlight(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid flight ID",
		})
	}

	var req CreateFlightRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	flightModel, err := h.service.UpdateFlight(uint(id), req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	flightResponse := ToFlightResponse(*flightModel)

	return c.JSON(fiber.Map{
		"message": "flight updated successfully",
		"data":    flightResponse,
	})
}

func (h *FlightHandler) DeleteFlight(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid flight ID",
		})
	}

	if err := h.service.DeleteFlight(uint(id)); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "flight deleted successfully",
	})
}