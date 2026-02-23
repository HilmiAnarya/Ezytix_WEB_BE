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

func (r *airlineRepository) CreateAirline(airline *models.Airline) error {
	return r.db.Create(airline).Error
}

func (r *airlineRepository) GetAllAirlines() ([]models.Airline, error) {
	var airlines []models.Airline
	err := r.db.Order("name ASC").Find(&airlines).Error
	return airlines, err
}

func (r *airlineRepository) GetAirlineByID(id uint) (*models.Airline, error) {
	var airline models.Airline
	err := r.db.First(&airline, id).Error
	if err != nil {
		return nil, err
	}
	return &airline, nil
}

func (r *airlineRepository) UpdateAirline(airline *models.Airline) error {
	return r.db.Save(airline).Error
}

func (r *airlineRepository) DeleteAirline(id uint) error {
	return r.db.Delete(&models.Airline{}, id).Error
}