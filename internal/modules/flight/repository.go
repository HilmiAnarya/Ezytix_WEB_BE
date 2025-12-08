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
	err := r.db.Preload("FlightLegs").Preload("FlightClasses").First(&flight, id).Error
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