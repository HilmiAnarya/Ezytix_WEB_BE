package flight

import (
	"ezytix-be/internal/models"
	"time"

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

	// Kita perlu preload Airline (Master) untuk ditampilkan di list admin
	err := r.db.
		Preload("Airline").                 // Ambil Validating Carrier
		Preload("OriginAirport").
		Preload("DestinationAirport").
		Preload("FlightClasses").
		Order("created_at DESC").           // Best practice: urutkan dari terbaru
		Find(&flights).Error

	return flights, err
}

func (r *flightRepository) GetFlightByID(id uint) (*models.Flight, error) {
	var flight models.Flight

	err := r.db.
		// 1. Info Utama
		Preload("Airline").
		Preload("OriginAirport").
		Preload("DestinationAirport").
		
		// 2. Info Detail Legs (Nested Preload)
		Preload("FlightLegs").
		Preload("FlightLegs.Airline").          // PENTING: Logo maskapai per leg
		Preload("FlightLegs.OriginAirport").    // PENTING: Nama bandara per leg
		Preload("FlightLegs.DestinationAirport"). 
		
		// 3. Info Harga
		Preload("FlightClasses").
		
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

	// Start Query
	query := r.db.Model(&models.Flight{}).
		Preload("Airline").                 // Logo Header
		Preload("OriginAirport").
		Preload("DestinationAirport").
		Preload("FlightLegs").
		Preload("FlightLegs.Airline").      // Logo Detail Transit
		Preload("FlightLegs.OriginAirport").
		Preload("FlightLegs.DestinationAirport").
		Joins("JOIN flight_classes ON flight_classes.flight_id = flights.id")

	// --- Filtering ---

	if req.OriginAirportID != 0 {
		query = query.Where("flights.origin_airport_id = ?", req.OriginAirportID)
	}
	if req.DestinationAirportID != 0 {
		query = query.Where("flights.destination_airport_id = ?", req.DestinationAirportID)
	}

	// 🔥 FIX UTAMA DI SINI (DATE RANGE FILTER) 🔥
	if req.DepartureDate != "" {
		// 1. Parsing string tanggal (Format YYYY-MM-DD)
		// Kita asumsikan input "2024-05-20"
		dateLayout := "2006-01-02" 
		parsedDate, err := time.Parse(dateLayout, req.DepartureDate)
		
		if err == nil {
			// 2. Buat Rentang Waktu (Start of Day - End of Day)
			// PENTING: Jika DB menyimpan UTC, pastikan range ini mencakup "overlap" 
			// atau kita cari "sepanjang hari itu" secara absolut.
			
			// Start: 2024-05-20 00:00:00
			startOfDay := parsedDate 
			
			// End:   2024-05-20 23:59:59 (atau < 2024-05-21 00:00:00)
			endOfDay := startOfDay.Add(24 * time.Hour)

			// Query: "departure_time >= Start AND departure_time < End"
			// Ini menangani TIMESTAMP dengan sangat cepat (memakai Index)
			query = query.Where("flights.departure_time >= ? AND flights.departure_time < ?", startOfDay, endOfDay)
		}
	}

	// Filter Seat Class (Economy, Business)
	// Note: Pastikan kolom di DB namanya 'name' atau 'seat_class' sesuai model FlightClass
	if req.SeatClass != "" {
		query = query.Where("flight_classes.seat_class = ?", req.SeatClass) 
	}

	// Filter Jumlah Penumpang (Pastikan ketersediaan kursi cukup)
	if req.PassengerCount > 0 {
		query = query.Where("flight_classes.total_seats >= ?", req.PassengerCount)
	}

	// Preload spesifik kelas yang dicari agar tidak semua kelas termuat (Optional Optimization)
	if req.SeatClass != "" {
		query = query.Preload("FlightClasses", "seat_class = ?", req.SeatClass)
	} else {
		query = query.Preload("FlightClasses")
	}

	// Eksekusi (Distinct agar tidak duplikat jika join match multiple classes)
	err := query.Distinct("flights.*").Find(&flights).Error

	return flights, err
}