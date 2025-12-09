package flight

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type FlightHandler struct {
	service FlightService
}

func NewFlightHandler(service FlightService) *FlightHandler {
	return &FlightHandler{service}
}

// ================================================
// 1. CREATE FLIGHT (POST)
// ================================================
func (h *FlightHandler) CreateFlight(c *fiber.Ctx) error {
	var req CreateFlightRequest

	// Parsing Body
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	// TODO: Di sini idealnya kita panggil 'go-playground/validator'
	// untuk mengecek tag `validate` di DTO. 
	// (Nanti kita bahas cara pasang validator global biar clean).

	flight, err := h.service.CreateFlight(req)
	if err != nil {
		// Asumsi error bisnis return 400/500
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "flight created successfully",
		"data":    flight,
	})
}

// ================================================
// 2. GET ALL FLIGHTS (GET)
// ================================================
// GET ALL FLIGHTS (Smart Endpoint: Search or List)
func (h *FlightHandler) GetAllFlights(c *fiber.Ctx) error {
	var req SearchFlightRequest
	
	// Parsing Query Param: ?origin=1&destination=2&date=2025-10-08&seat_class=economy&passengers=1
	if err := c.QueryParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid query params"})
	}

	// DETEKSI MODE: Apakah ini pencarian?
	// Syarat search: Harus ada Origin, Destination, dan Date.
	if req.OriginAirportID != 0 && req.DestinationAirportID != 0 && req.DepartureDate != "" {
		flights, err := h.service.SearchFlights(req)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(fiber.Map{"data": flights})
	}

	// MODE DEFAULT: Get All (Untuk Admin list / Debugging)
	flights, err := h.service.GetAllFlights()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to fetch flights",
		})
	}

	return c.JSON(fiber.Map{
		"data": flights,
	})
}

// ================================================
// 3. GET FLIGHT BY ID (GET /:id)
// ================================================
func (h *FlightHandler) GetFlightByID(c *fiber.Ctx) error {
	// Parsing ID dari URL param
	idStr := c.Params("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid flight ID",
		})
	}

	flight, err := h.service.GetFlightByID(uint(id))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "flight not found",
		})
	}

	return c.JSON(fiber.Map{
		"data": flight,
	})
}

// ================================================
// 4. UPDATE FLIGHT (PUT /:id)
// ================================================
func (h *FlightHandler) UpdateFlight(c *fiber.Ctx) error {
	// 1. Parsing ID
	idStr := c.Params("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid flight ID",
		})
	}

	// 2. Parsing Body (Payload Full Replacement)
	var req CreateFlightRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	// 3. Panggil Service
	flight, err := h.service.UpdateFlight(uint(id), req)
	if err != nil {
		// Kita bisa membedakan error Not Found vs Internal Server Error
		// Tapi untuk simplifikasi awal, kita return 500 atau 400
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "flight updated successfully",
		"data":    flight,
	})
}

// ================================================
// 5. DELETE FLIGHT (DELETE /:id)
// ================================================
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