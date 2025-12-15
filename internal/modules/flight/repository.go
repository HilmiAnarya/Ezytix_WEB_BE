package flight

import (
	"ezytix-be/internal/models"
	"gorm.io/gorm"
)

type FlightRepository interface {
	CreateFlight(flight *models.Flight) error
	GetAllFlights() ([]models.Flight, error)
	GetFlightByID(id uint) (*models.Flight, error)
	UpdateFlight(flight *models.Flight) error
	DeleteFlight(id uint) error

	SearchFlights(req SearchFlightRequest) ([]models.Flight, error)
}

type flightRepository struct {
	db *gorm.DB
}

func NewFlightRepository(db *gorm.DB) FlightRepository {
	return &flightRepository{db}
}

func (r *flightRepository) CreateFlight(flight *models.Flight) error {
	return r.db.Create(flight).Error
}

func (r *flightRepository) GetAllFlights() ([]models.Flight, error) {
	var flights []models.Flight

	err := r.db.Preload("FlightLegs").Preload("FlightClasses").Find(&flights).Error
	return flights, err
}

func (r *flightRepository) GetFlightByID(id uint) (*models.Flight, error) {
	var flight models.Flight
	
	err := r.db.
		Preload("FlightLegs").
		Preload("FlightClasses").
		Preload("OriginAirport").      
		Preload("DestinationAirport"). 
		First(&flight, id).Error
	
	if err != nil {
		return nil, err
	}
	return &flight, nil
}

func (r *flightRepository) UpdateFlight(flight *models.Flight) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(flight).Omit("FlightLegs", "FlightClasses").Updates(flight).Error; err != nil {
			return err
		}

		if err := tx.Where("flight_id = ?", flight.ID).Delete(&models.FlightLeg{}).Error; err != nil {
			return err
		}

		if len(flight.FlightLegs) > 0 {
			for i := range flight.FlightLegs {
				flight.FlightLegs[i].FlightID = flight.ID
			}
			if err := tx.Create(&flight.FlightLegs).Error; err != nil {
				return err
			}
		}

		if err := tx.Where("flight_id = ?", flight.ID).Delete(&models.FlightClass{}).Error; err != nil {
			return err
		}
		if len(flight.FlightClasses) > 0 {
			for i := range flight.FlightClasses {
				flight.FlightClasses[i].FlightID = flight.ID
			}
			if err := tx.Create(&flight.FlightClasses).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *flightRepository) DeleteFlight(id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(&models.Flight{}, id).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *flightRepository) SearchFlights(req SearchFlightRequest) ([]models.Flight, error) {
	var flights []models.Flight

	query := r.db.Model(&models.Flight{}).
		Preload("FlightLegs"). 
		Joins("JOIN flight_classes ON flight_classes.flight_id = flights.id") 

	if req.OriginAirportID != 0 {
		query = query.Where("flights.origin_airport_id = ?", req.OriginAirportID)
	}
	if req.DestinationAirportID != 0 {
		query = query.Where("flights.destination_airport_id = ?", req.DestinationAirportID)
	}

	if req.DepartureDate != "" {
		query = query.Where("DATE(flights.departure_time) = ?", req.DepartureDate)
	}

	if req.SeatClass != "" {
		query = query.Where("flight_classes.seat_class = ?", req.SeatClass)
	}

	if req.PassengerCount > 0 {
		query = query.Where("flight_classes.total_seats >= ?", req.PassengerCount)
	}

	if req.SeatClass != "" {
		query = query.Preload("FlightClasses", "seat_class = ?", req.SeatClass)
	} else {
		query = query.Preload("FlightClasses")
	}

	err := query.Distinct("flights.*").Find(&flights).Error

	return flights, err
}