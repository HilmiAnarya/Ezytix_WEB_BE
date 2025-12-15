package flight

import (
	"errors"
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

	transitCount := len(req.Legs) - 1
	if transitCount < 0 {
		transitCount = 0 
	}

	flight := &models.Flight{
		FlightCode:           req.FlightCode,
		AirlineName:          req.AirlineName,
		OriginAirportID:      req.OriginAirportID,
		DestinationAirportID: req.DestinationAirportID,
		DepartureTime:        req.DepartureTime,
		ArrivalTime:          req.ArrivalTime,
		TotalDuration:        req.TotalDuration,
		TransitCount:         transitCount, 
		TransitInfo:          req.TransitInfo,
	}

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

	var classes []models.FlightClass
	for _, classReq := range req.Classes {
		classes = append(classes, models.FlightClass{
			SeatClass:  classReq.SeatClass,
			Price:      classReq.Price,
			TotalSeats: classReq.TotalSeats,
		})
	}
	flight.FlightClasses = classes

	if err := s.repo.CreateFlight(flight); err != nil {
		return nil, err
	}
	return flight, nil
}

func (s *flightService) GetAllFlights() ([]models.Flight, error) {
	return s.repo.GetAllFlights()
}

func (s *flightService) GetFlightByID(id uint) (*models.Flight, error) {
	return s.repo.GetFlightByID(id)
}

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

	existingFlight.FlightCode = req.FlightCode
	existingFlight.AirlineName = req.AirlineName
	existingFlight.OriginAirportID = req.OriginAirportID
	existingFlight.DestinationAirportID = req.DestinationAirportID
	existingFlight.DepartureTime = req.DepartureTime
	existingFlight.ArrivalTime = req.ArrivalTime
	existingFlight.TotalDuration = req.TotalDuration
	existingFlight.TransitCount = len(req.Legs) - 1
	if existingFlight.TransitCount < 0 { existingFlight.TransitCount = 0 }
	existingFlight.TransitInfo = req.TransitInfo

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

	var newClasses []models.FlightClass
	for _, classReq := range req.Classes {
		newClasses = append(newClasses, models.FlightClass{
			SeatClass:  classReq.SeatClass,
			Price:      classReq.Price,
			TotalSeats: classReq.TotalSeats,
		})
	}
	existingFlight.FlightClasses = newClasses

	if err := s.repo.UpdateFlight(existingFlight); err != nil {
		return nil, err
	}
	return existingFlight, nil
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