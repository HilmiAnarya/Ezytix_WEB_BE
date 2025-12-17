package airline

import (
	"errors"
	"strings" // Penting untuk memaksa Uppercase IATA
	"ezytix-be/internal/models"
)

type AirlineService interface {
	CreateAirline(req CreateAirlineRequest) (*models.Airline, error)
	GetAllAirlines() ([]models.Airline, error)
	GetAirlineByID(id uint) (*models.Airline, error)
	UpdateAirline(id uint, req UpdateAirlineRequest) (*models.Airline, error)
	DeleteAirline(id uint) error
}

type airlineService struct {
	repo AirlineRepository
}

func NewAirlineService(repo AirlineRepository) AirlineService {
	return &airlineService{repo}
}

func (s *airlineService) CreateAirline(req CreateAirlineRequest) (*models.Airline, error) {
	// Business Logic: IATA Code wajib Uppercase (misal: "ga" -> "GA")
	iataUpper := strings.ToUpper(req.IATA)

	airline := &models.Airline{
		IATA:    iataUpper,
		Name:    req.Name,
		LogoURL: req.LogoURL,
	}

	if err := s.repo.CreateAirline(airline); err != nil {
		// Handle duplicate IATA error dari database
		if strings.Contains(err.Error(), "duplicate key") {
			return nil, errors.New("airline with this IATA code already exists")
		}
		return nil, err
	}

	return airline, nil
}

func (s *airlineService) GetAllAirlines() ([]models.Airline, error) {
	return s.repo.GetAllAirlines()
}

func (s *airlineService) GetAirlineByID(id uint) (*models.Airline, error) {
	airline, err := s.repo.GetAirlineByID(id)
	if err != nil {
		return nil, errors.New("airline not found")
	}
	return airline, nil
}

func (s *airlineService) UpdateAirline(id uint, req UpdateAirlineRequest) (*models.Airline, error) {
	// 1. Cek apakah data ada
	existingAirline, err := s.repo.GetAirlineByID(id)
	if err != nil {
		return nil, errors.New("airline not found")
	}

	// 2. Partial Update (Hanya update jika field diisi)
	if req.Name != "" {
		existingAirline.Name = req.Name
	}
	if req.IATA != "" {
		existingAirline.IATA = strings.ToUpper(req.IATA)
	}
	if req.LogoURL != "" {
		existingAirline.LogoURL = req.LogoURL
	}

	// 3. Simpan perubahan
	if err := s.repo.UpdateAirline(existingAirline); err != nil {
		return nil, err
	}

	return existingAirline, nil
}

func (s *airlineService) DeleteAirline(id uint) error {
	// Cek keberadaan data sebelum delete
	_, err := s.repo.GetAirlineByID(id)
	if err != nil {
		return errors.New("airline not found")
	}

	return s.repo.DeleteAirline(id)
}