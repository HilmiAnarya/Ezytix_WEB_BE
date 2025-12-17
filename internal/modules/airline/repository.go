package airline

import (
	"ezytix-be/internal/models"
	"gorm.io/gorm"
)

type AirlineRepository interface {
	CreateAirline(airline *models.Airline) error
	GetAllAirlines() ([]models.Airline, error)
	GetAirlineByID(id uint) (*models.Airline, error)
	UpdateAirline(airline *models.Airline) error
	DeleteAirline(id uint) error
}

type airlineRepository struct {
	db *gorm.DB
}

func NewAirlineRepository(db *gorm.DB) AirlineRepository {
	return &airlineRepository{db}
}

// CreateAirline: Menambahkan maskapai baru
func (r *airlineRepository) CreateAirline(airline *models.Airline) error {
	return r.db.Create(airline).Error
}

// GetAllAirlines: Mengambil semua data maskapai (untuk dropdown list admin)
func (r *airlineRepository) GetAllAirlines() ([]models.Airline, error) {
	var airlines []models.Airline
	// Kita urutkan berdasarkan Nama agar rapi saat ditampilkan di dropdown
	err := r.db.Order("name ASC").Find(&airlines).Error
	return airlines, err
}

// GetAirlineByID: Mencari maskapai spesifik
func (r *airlineRepository) GetAirlineByID(id uint) (*models.Airline, error) {
	var airline models.Airline
	err := r.db.First(&airline, id).Error
	if err != nil {
		return nil, err
	}
	return &airline, nil
}

// UpdateAirline: Menyimpan perubahan data maskapai
func (r *airlineRepository) UpdateAirline(airline *models.Airline) error {
	return r.db.Save(airline).Error
}

// DeleteAirline: Menghapus maskapai
func (r *airlineRepository) DeleteAirline(id uint) error {
	// Peringatan: Karena kita pakai ON DELETE CASCADE di migrasi,
	// menghapus airline akan menghapus SEMUA flight & booking terkait.
	// Admin harus diperingatkan di frontend nanti.
	return r.db.Delete(&models.Airline{}, id).Error
}