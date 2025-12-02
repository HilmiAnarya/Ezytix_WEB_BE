package airport

import (
	"ezytix-be/internal/models"

	"gorm.io/gorm"
)

// ==============================================
// INTERFACE
// ==============================================

type AirportRepository interface {
	CreateAirport(data *models.Airport) error
	UpdateAirport(data *models.Airport) error
	DeleteAirport(id uint) error
	FindAirportByID(id uint) (*models.Airport, error)
	FindAirportByCode(code string) (*models.Airport, error)
	FindAllAirports() ([]models.Airport, error)
}

type airportRepository struct {
	db *gorm.DB
}

// Constructor
func NewAirportRepository(db *gorm.DB) AirportRepository {
	return &airportRepository{
		db: db,
	}
}

// ==============================================
// IMPLEMENTATION
// ==============================================

// Create Airport
func (r *airportRepository) CreateAirport(data *models.Airport) error {
	return r.db.Create(data).Error
}

// Update Airport
func (r *airportRepository) UpdateAirport(data *models.Airport) error {
	return r.db.Save(data).Error
}

// Soft Delete Airport
func (r *airportRepository) DeleteAirport(id uint) error {
	return r.db.Delete(&models.Airport{}, id).Error
}

// Find Airport by ID
func (r *airportRepository) FindAirportByID(id uint) (*models.Airport, error) {
	var airport models.Airport
	if err := r.db.First(&airport, id).Error; err != nil {
		return nil, err
	}
	return &airport, nil
}

// Find Airport by IATA Code
func (r *airportRepository) FindAirportByCode(code string) (*models.Airport, error) {
	var airport models.Airport
	if err := r.db.Where("code = ?", code).First(&airport).Error; err != nil {
		return nil, err
	}
	return &airport, nil
}

// Get all airports
func (r *airportRepository) FindAllAirports() ([]models.Airport, error) {
	var data []models.Airport
	if err := r.db.Order("code ASC").Find(&data).Error; err != nil {
		return nil, err
	}
	return data, nil
}
