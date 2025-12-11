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

	// [NEW] Search dengan filter lengkap
	SearchFlights(req SearchFlightRequest) ([]models.Flight, error)
}

type flightRepository struct {
	db *gorm.DB
}

func NewFlightRepository(db *gorm.DB) FlightRepository {
	return &flightRepository{db}
}

// CREATE (Transactional)
func (r *flightRepository) CreateFlight(flight *models.Flight) error {
	return r.db.Create(flight).Error
}

// READ ALL (With Relations)
func (r *flightRepository) GetAllFlights() ([]models.Flight, error) {
	var flights []models.Flight
	// Preload wajib dipakai agar data Legs dan Classes ikut terbawa
	err := r.db.Preload("FlightLegs").Preload("FlightClasses").Find(&flights).Error
	return flights, err
}

// READ ONE (With Relations)
func (r *flightRepository) GetFlightByID(id uint) (*models.Flight, error) {
	var flight models.Flight
	
	// Tambahkan Preload OriginAirport dan DestinationAirport
	err := r.db.
		Preload("FlightLegs").
		Preload("FlightClasses").
		Preload("OriginAirport").      // <--- NEW: Agar CityName bisa diambil
		Preload("DestinationAirport"). // <--- NEW
		First(&flight, id).Error
	
	if err != nil {
		return nil, err
	}
	return &flight, nil
}

// UPDATE (Full Replacement Strategy)
func (r *flightRepository) UpdateFlight(flight *models.Flight) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 1. Update Header Flight (Flight Code, Time, dll)
		// Kita pakai Omit("FlightLegs", "FlightClasses") agar GORM tidak bingung update relasi secara default
		if err := tx.Model(flight).Omit("FlightLegs", "FlightClasses").Updates(flight).Error; err != nil {
			return err
		}

		// 2. REPLACEMENT FLIGHT LEGS
		// a. Hapus legs lama
		if err := tx.Where("flight_id = ?", flight.ID).Delete(&models.FlightLeg{}).Error; err != nil {
			return err
		}
		// b. Insert legs baru (jika ada)
		if len(flight.FlightLegs) > 0 {
			// Pastikan ID diset agar masuk ke parent yang benar
			for i := range flight.FlightLegs {
				flight.FlightLegs[i].FlightID = flight.ID
			}
			if err := tx.Create(&flight.FlightLegs).Error; err != nil {
				return err
			}
		}

		// 3. REPLACEMENT FLIGHT CLASSES
		// a. Hapus classes lama
		if err := tx.Where("flight_id = ?", flight.ID).Delete(&models.FlightClass{}).Error; err != nil {
			return err
		}
		// b. Insert classes baru (jika ada)
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

// DELETE (Transactional)
func (r *flightRepository) DeleteFlight(id uint) error {
	// Kita gunakan Transaction untuk memastikan konsistensi
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 1. Hapus Flight (Karena ada Constraint ON DELETE CASCADE di database,
		// legs dan classes harusnya otomatis terhapus di DB level.
		// Tapi untuk GORM Soft Delete, kita hapus parent-nya saja cukup).
		if err := tx.Delete(&models.Flight{}, id).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *flightRepository) SearchFlights(req SearchFlightRequest) ([]models.Flight, error) {
	var flights []models.Flight

	// Start Query dari Model Flight
	query := r.db.Model(&models.Flight{}).
		Preload("FlightLegs"). // Selalu load rute transit
		Joins("JOIN flight_classes ON flight_classes.flight_id = flights.id") // Join wajib untuk filter kelas

	// 1. Filter Rute (Origin & Destination)
	if req.OriginAirportID != 0 {
		query = query.Where("flights.origin_airport_id = ?", req.OriginAirportID)
	}
	if req.DestinationAirportID != 0 {
		query = query.Where("flights.destination_airport_id = ?", req.DestinationAirportID)
	}

	// 2. Filter Tanggal (Departure Date)
	// Kita casting timestamp ke date agar jam diabaikan (User cuma pilih tgl 8 Okt)
	if req.DepartureDate != "" {
		query = query.Where("DATE(flights.departure_time) = ?", req.DepartureDate)
	}

	// 3. Filter Kelas Kabin (Seat Class)
	if req.SeatClass != "" {
		query = query.Where("flight_classes.seat_class = ?", req.SeatClass)
	}

	// 4. Filter Ketersediaan Kursi (Passenger Count)
	if req.PassengerCount > 0 {
		query = query.Where("flight_classes.total_seats >= ?", req.PassengerCount)
	}

	// 5. Preload Classes yang relevan saja
	// Trik UX: Jika user cari "Business", jangan load harga "Economy" biar response bersih.
	// Jika user tidak filter kelas, tampilkan semua.
	if req.SeatClass != "" {
		query = query.Preload("FlightClasses", "seat_class = ?", req.SeatClass)
	} else {
		query = query.Preload("FlightClasses")
	}

	// 6. Eksekusi (Pakai Distinct karena Join bisa bikin duplikat row flight)
	err := query.Distinct("flights.*").Find(&flights).Error

	return flights, err
}