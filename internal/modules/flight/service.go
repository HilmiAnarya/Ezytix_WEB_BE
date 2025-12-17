package flight

import (
	"errors"
	"fmt"
	"ezytix-be/internal/models"
)

type FlightService interface {
	CreateFlight(req CreateFlightRequest) (*models.Flight, error)
	GetAllFlights() ([]models.Flight, error)
	GetFlightByID(id uint) (*models.Flight, error)
	UpdateFlight(id uint, req CreateFlightRequest) (*models.Flight, error)
	DeleteFlight(id uint) error

	SearchFlights(req SearchFlightRequest) ([]models.Flight, error)
}

type flightService struct {
	repo FlightRepository
}

func NewFlightService(repo FlightRepository) FlightService {
	return &flightService{repo}
}

func (s *flightService) CreateFlight(req CreateFlightRequest) (*models.Flight, error) {
	if req.OriginAirportID == req.DestinationAirportID {
		return nil, errors.New("origin and destination airport cannot be the same")
	}
	if req.ArrivalTime.Before(req.DepartureTime) {
		return nil, errors.New("arrival time must be after departure time")
	}

	// 2. Auto-Calculate Total Duration (Menit)
	// Kita hitung selisih waktu dalam menit, lalu convert ke int
	totalDurationMinutes := int(req.ArrivalTime.Sub(req.DepartureTime).Minutes())

	// 3. Auto-Calculate Transit Info
	transitCount := len(req.FlightLegs) - 1
	if transitCount < 0 {
		transitCount = 0
	}

	transitInfo := "Direct"
	if transitCount > 0 {
		transitInfo = fmt.Sprintf("%d Transit", transitCount)
	}

	flight := &models.Flight{
		FlightCode:           req.FlightCode,
		AirlineID:            req.AirlineID,
		OriginAirportID:      req.OriginAirportID,
		DestinationAirportID: req.DestinationAirportID,
		DepartureTime:        req.DepartureTime,
		ArrivalTime:          req.ArrivalTime,
		// Computed Values
		TotalDuration:        totalDurationMinutes,
		TransitCount:         transitCount,
		TransitInfo:          transitInfo,
	}

	var legs []models.FlightLeg
	for _, legReq := range req.FlightLegs {
		// Validasi per leg
		if legReq.ArrivalTime.Before(legReq.DepartureTime) {
			return nil, errors.New("leg arrival time must be after departure time")
		}

		// Calculate Leg Duration
		legDuration := int(legReq.ArrivalTime.Sub(legReq.DepartureTime).Minutes())

		legs = append(legs, models.FlightLeg{
			LegOrder:             legReq.LegOrder,
			OriginAirportID:      legReq.OriginAirportID,
			DestinationAirportID: legReq.DestinationAirportID,
			DepartureTime:        legReq.DepartureTime,
			ArrivalTime:          legReq.ArrivalTime,
			FlightNumber:         legReq.FlightNumber,

			// Normalized & Computed
			AirlineID:            legReq.AirlineID, // Operating Carrier

			Duration:             legDuration,
			TransitNotes:         legReq.TransitNotes,
		})
	}
	flight.FlightLegs = legs

	// 6. Mapping Classes
	var classes []models.FlightClass
	for _, classReq := range req.FlightClasses {
		classes = append(classes, models.FlightClass{
			SeatClass:       classReq.SeatClass, // Sesuaikan dengan field di Model (Name/SeatClass)
			Price:      classReq.Price,
			TotalSeats: classReq.TotalSeats,
		})
	}
	flight.FlightClasses = classes

	if err := s.repo.CreateFlight(flight); err != nil {
		return nil, err
	}
	return s.repo.GetFlightByID(flight.ID)
}

func (s *flightService) GetAllFlights() ([]models.Flight, error) {
	return s.repo.GetAllFlights()
}

func (s *flightService) GetFlightByID(id uint) (*models.Flight, error) {
	return s.repo.GetFlightByID(id)
}

func (s *flightService) UpdateFlight(id uint, req CreateFlightRequest) (*models.Flight, error) {
	// Ambil data lama
	existingFlight, err := s.repo.GetFlightByID(id)
	if err != nil {
		return nil, errors.New("flight not found")
	}

	// Validasi dasar
	if req.OriginAirportID == req.DestinationAirportID {
		return nil, errors.New("origin and destination airport cannot be the same")
	}
	if req.ArrivalTime.Before(req.DepartureTime) {
		return nil, errors.New("arrival time must be after departure time")
	}

	// Recalculate Logic
	totalDurationMinutes := int(req.ArrivalTime.Sub(req.DepartureTime).Minutes())
	transitCount := len(req.FlightLegs) - 1
	if transitCount < 0 { transitCount = 0 }
	
	transitInfo := "Direct"
	if transitCount > 0 {
		transitInfo = fmt.Sprintf("%d Transit", transitCount)
	}

	// Update Fields
	existingFlight.FlightCode = req.FlightCode
	existingFlight.AirlineID = req.AirlineID
	existingFlight.OriginAirportID = req.OriginAirportID
	existingFlight.DestinationAirportID = req.DestinationAirportID
	existingFlight.DepartureTime = req.DepartureTime
	existingFlight.ArrivalTime = req.ArrivalTime
	
	// Update Computed Values
	existingFlight.TotalDuration = totalDurationMinutes
	existingFlight.TransitCount = transitCount
	existingFlight.TransitInfo = transitInfo

	// Update Legs
	var newLegs []models.FlightLeg
	for _, legReq := range req.FlightLegs {
		if legReq.ArrivalTime.Before(legReq.DepartureTime) {
			return nil, errors.New("leg arrival time must be after departure time")
		}
		legDuration := int(legReq.ArrivalTime.Sub(legReq.DepartureTime).Minutes())

		newLegs = append(newLegs, models.FlightLeg{
			LegOrder:             legReq.LegOrder,
			OriginAirportID:      legReq.OriginAirportID,
			DestinationAirportID: legReq.DestinationAirportID,
			DepartureTime:        legReq.DepartureTime,
			ArrivalTime:          legReq.ArrivalTime,
			FlightNumber:         legReq.FlightNumber,
			AirlineID:            legReq.AirlineID,
			Duration:             legDuration,
			TransitNotes:         legReq.TransitNotes,
		})
	}
	existingFlight.FlightLegs = newLegs

	// Update Classes
	var newClasses []models.FlightClass
	for _, classReq := range req.FlightClasses {
		newClasses = append(newClasses, models.FlightClass{
			SeatClass:       classReq.SeatClass,
			Price:      classReq.Price,
			TotalSeats: classReq.TotalSeats,
		})
	}
	existingFlight.FlightClasses = newClasses

	if err := s.repo.UpdateFlight(existingFlight); err != nil {
		return nil, err
	}
	return s.repo.GetFlightByID(existingFlight.ID)
}

func (s *flightService) DeleteFlight(id uint) error {
	_, err := s.repo.GetFlightByID(id)
	if err != nil {
		return errors.New("flight not found")
	}

	return s.repo.DeleteFlight(id)
}

func (s *flightService) SearchFlights(req SearchFlightRequest) ([]models.Flight, error) {
	if req.OriginAirportID == 0 || req.DestinationAirportID == 0 || req.DepartureDate == "" {
		return nil, errors.New("origin, destination, and date are required")
	}

	if req.PassengerCount <= 0 {
		req.PassengerCount = 1 
	}

	return s.repo.SearchFlights(req)
}