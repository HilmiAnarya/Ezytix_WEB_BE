package airport

import (
	"ezytix-be/internal/models"

	"gorm.io/gorm"
)

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

func NewAirportRepository(db *gorm.DB) AirportRepository {
	return &airportRepository{
		db: db,
	}
}

func (r *airportRepository) CreateAirport(data *models.Airport) error {
	return r.db.Create(data).Error
}

func (r *airportRepository) UpdateAirport(data *models.Airport) error {
	return r.db.Save(data).Error
}

func (r *airportRepository) DeleteAirport(id uint) error {
	return r.db.Delete(&models.Airport{}, id).Error
}

func (r *airportRepository) FindAirportByID(id uint) (*models.Airport, error) {
	var airport models.Airport
	if err := r.db.First(&airport, id).Error; err != nil {
		return nil, err
	}
	return &airport, nil
}

func (r *airportRepository) FindAirportByCode(code string) (*models.Airport, error) {
	var airport models.Airport
	if err := r.db.Where("code = ?", code).First(&airport).Error; err != nil {
		return nil, err
	}
	return &airport, nil
}

func (r *airportRepository) FindAllAirports() ([]models.Airport, error) {
	var data []models.Airport
	if err := r.db.Order("code ASC").Find(&data).Error; err != nil {
		return nil, err
	}
	return data, nil
}
