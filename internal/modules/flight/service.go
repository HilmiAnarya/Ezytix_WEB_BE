package flight

import (
	"errors"
	"ezytix-be/internal/models"
)

// Definisikan Kontrak Service (Interface)
type FlightService interface {
	CreateFlight(req CreateFlightRequest) (*models.Flight, error)
	GetAllFlights() ([]models.Flight, error)
	GetFlightByID(id uint) (*models.Flight, error)
	UpdateFlight(id uint, req CreateFlightRequest) (*models.Flight, error)
	DeleteFlight(id uint) error
}

type flightService struct {
	repo FlightRepository
}

// Constructor
func NewFlightService(repo FlightRepository) FlightService {
	return &flightService{repo}
}

// ==========================================
// 1. CREATE FLIGHT (The "Assembler")
// ==========================================
func (s *flightService) CreateFlight(req CreateFlightRequest) (*models.Flight, error) {
	// 1. Business Logic Validation
	if req.OriginAirportID == req.DestinationAirportID {
		return nil, errors.New("origin and destination airport cannot be the same")
	}
	if req.ArrivalTime.Before(req.DepartureTime) {
		return nil, errors.New("arrival time must be after departure time")
	}

	// 2. AUTO-CALCULATE TRANSIT COUNT
	// Logic: Jika legs = 1, transit = 0. Jika legs = 3, transit = 2.
	transitCount := len(req.Legs) - 1
	if transitCount < 0 {
		transitCount = 0 // Jaga-jaga walau sudah divalidasi min=1
	}

	// 3. Mapping Header Flight
	flight := &models.Flight{
		FlightCode:           req.FlightCode,
		AirlineName:          req.AirlineName,
		OriginAirportID:      req.OriginAirportID,
		DestinationAirportID: req.DestinationAirportID,
		DepartureTime:        req.DepartureTime,
		ArrivalTime:          req.ArrivalTime,
		TotalDuration:        req.TotalDuration,
		
		// INI DIA HASILNYA:
		TransitCount:         transitCount, 
		
		TransitInfo:          req.TransitInfo,
	}

	// 4. Mapping Legs
	var legs []models.FlightLeg
	for _, legReq := range req.Legs {
		legs = append(legs, models.FlightLeg{
			LegOrder:             legReq.LegOrder,
			OriginAirportID:      legReq.OriginAirportID,
			DestinationAirportID: legReq.DestinationAirportID,
			DepartureTime:        legReq.DepartureTime,
			ArrivalTime:          legReq.ArrivalTime,
			FlightNumber:         legReq.FlightNumber,
			AirlineName:          legReq.AirlineName,
			AirlineLogo:          legReq.AirlineLogo,
			DepartureTerminal:    legReq.DepartureTerminal,
			ArrivalTerminal:      legReq.ArrivalTerminal,
			Duration:             legReq.Duration,
			TransitNotes:         legReq.TransitNotes,
		})
	}
	flight.FlightLegs = legs

	// 5. Mapping Classes
	var classes []models.FlightClass
	for _, classReq := range req.Classes {
		classes = append(classes, models.FlightClass{
			SeatClass: classReq.SeatClass,
			Price:     classReq.Price,
		})
	}
	flight.FlightClasses = classes

	// 6. Save
	if err := s.repo.CreateFlight(flight); err != nil {
		return nil, err
	}

	return flight, nil
}

// ==========================================
// 2. GET ALL
// ==========================================
func (s *flightService) GetAllFlights() ([]models.Flight, error) {
	return s.repo.GetAllFlights()
}

// ==========================================
// 3. GET BY ID
// ==========================================
func (s *flightService) GetFlightByID(id uint) (*models.Flight, error) {
	return s.repo.GetFlightByID(id)
}

// ==========================================
// 4. UPDATE FLIGHT
// ==========================================
func (s *flightService) UpdateFlight(id uint, req CreateFlightRequest) (*models.Flight, error) {
	existingFlight, err := s.repo.GetFlightByID(id)
	if err != nil {
		return nil, errors.New("flight not found")
	}

	if req.OriginAirportID == req.DestinationAirportID {
		return nil, errors.New("origin and destination airport cannot be the same")
	}
	if req.ArrivalTime.Before(req.DepartureTime) {
		return nil, errors.New("arrival time must be after departure time")
	}

	// AUTO-CALCULATE (UPDATE)
	// Kita hitung ulang berdasarkan legs baru yang dikirim Admin
	transitCount := len(req.Legs) - 1
	if transitCount < 0 {
		transitCount = 0
	}

	existingFlight.FlightCode = req.FlightCode
	existingFlight.AirlineName = req.AirlineName
	existingFlight.OriginAirportID = req.OriginAirportID
	existingFlight.DestinationAirportID = req.DestinationAirportID
	existingFlight.DepartureTime = req.DepartureTime
	existingFlight.ArrivalTime = req.ArrivalTime
	existingFlight.TotalDuration = req.TotalDuration
	
	// UPDATE FIELD TRANSIT COUNT
	existingFlight.TransitCount = transitCount
	
	existingFlight.TransitInfo = req.TransitInfo

	// Mapping Legs (Full Replace logic)
	var newLegs []models.FlightLeg
	for _, legReq := range req.Legs {
		newLegs = append(newLegs, models.FlightLeg{
			LegOrder:             legReq.LegOrder,
			OriginAirportID:      legReq.OriginAirportID,
			DestinationAirportID: legReq.DestinationAirportID,
			DepartureTime:        legReq.DepartureTime,
			ArrivalTime:          legReq.ArrivalTime,
			FlightNumber:         legReq.FlightNumber,
			AirlineName:          legReq.AirlineName,
			AirlineLogo:          legReq.AirlineLogo,
			DepartureTerminal:    legReq.DepartureTerminal,
			ArrivalTerminal:      legReq.ArrivalTerminal,
			Duration:             legReq.Duration,
			TransitNotes:         legReq.TransitNotes,
		})
	}
	existingFlight.FlightLegs = newLegs

	// Mapping Classes
	var newClasses []models.FlightClass
	for _, classReq := range req.Classes {
		newClasses = append(newClasses, models.FlightClass{
			SeatClass: classReq.SeatClass,
			Price:     classReq.Price,
		})
	}
	existingFlight.FlightClasses = newClasses

	if err := s.repo.UpdateFlight(existingFlight); err != nil {
		return nil, err
	}

	return existingFlight, nil
}

// ==========================================
// 5. DELETE FLIGHT
// ==========================================
func (s *flightService) DeleteFlight(id uint) error {
	// Cek apakah flight ada sebelum dihapus
	_, err := s.repo.GetFlightByID(id)
	if err != nil {
		return errors.New("flight not found")
	}

	return s.repo.DeleteFlight(id)
}