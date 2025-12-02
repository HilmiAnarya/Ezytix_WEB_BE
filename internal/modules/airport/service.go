package airport

import (
	"errors"
	"strings"

	"ezytix-be/internal/models"
)

type AirportService interface {
	CreateAirport(req CreateAirportRequest) (*models.Airport, error)
	UpdateAirport(id uint, req UpdateAirportRequest) (*models.Airport, error)
	DeleteAirport(id uint) error
	GetAirportByID(id uint) (*models.Airport, error)
	GetAllAirports() ([]models.Airport, error)
}

type airportService struct {
	repo AirportRepository
}

func NewAirportService(repo AirportRepository) AirportService {
	return &airportService{
		repo: repo,
	}
}

//////////////////////////////////////////////////
// CREATE
//////////////////////////////////////////////////

func (s *airportService) CreateAirport(req CreateAirportRequest) (*models.Airport, error) {
	// Normalisasi & validasi code
	code := strings.ToUpper(strings.TrimSpace(req.Code))
	if len(code) != 3 {
		return nil, errors.New("airport code harus 3 huruf (IATA)")
	}

	// Cek duplikasi code
	existing, _ := s.repo.FindAirportByCode(code)
	if existing != nil {
		return nil, errors.New("airport code sudah digunakan")
	}

	airport := &models.Airport{
		Code:        code,
		CityName:    strings.TrimSpace(req.CityName),
		AirportName: strings.TrimSpace(req.AirportName),
		Country:     strings.TrimSpace(req.Country),
	}

	if err := s.repo.CreateAirport(airport); err != nil {
		return nil, err
	}

	return airport, nil
}

//////////////////////////////////////////////////
// UPDATE
//////////////////////////////////////////////////

func (s *airportService) UpdateAirport(id uint, req UpdateAirportRequest) (*models.Airport, error) {
	airport, err := s.repo.FindAirportByID(id)
	if err != nil {
		return nil, errors.New("airport tidak ditemukan")
	}

	// Update code kalau dikirim
	if req.Code != nil {
		code := strings.ToUpper(strings.TrimSpace(*req.Code))
		if len(code) != 3 {
			return nil, errors.New("airport code harus 3 huruf (IATA)")
		}

		// Cek duplikasi code (pastikan bukan dirinya sendiri)
		existing, _ := s.repo.FindAirportByCode(code)
		if existing != nil && existing.ID != id {
			return nil, errors.New("airport code sudah dipakai airport lain")
		}

		airport.Code = code
	}

	// Update city_name kalau dikirim
	if req.CityName != nil {
		city := strings.TrimSpace(*req.CityName)
		if city != "" {
			airport.CityName = city
		}
	}

	// Update airport_name kalau dikirim
	if req.AirportName != nil {
		name := strings.TrimSpace(*req.AirportName)
		if name != "" {
			airport.AirportName = name
		}
	}

	// Update country kalau dikirim
	if req.Country != nil {
		country := strings.TrimSpace(*req.Country)
		if country != "" {
			airport.Country = country
		}
	}

	if err := s.repo.UpdateAirport(airport); err != nil {
		return nil, err
	}

	return airport, nil
}

//////////////////////////////////////////////////
// DELETE
//////////////////////////////////////////////////

func (s *airportService) DeleteAirport(id uint) error {
	_, err := s.repo.FindAirportByID(id)
	if err != nil {
		return errors.New("airport tidak ditemukan")
	}

	return s.repo.DeleteAirport(id)
}

//////////////////////////////////////////////////
// GET BY ID
//////////////////////////////////////////////////

func (s *airportService) GetAirportByID(id uint) (*models.Airport, error) {
	airport, err := s.repo.FindAirportByID(id)
	if err != nil {
		return nil, errors.New("airport tidak ditemukan")
	}
	return airport, nil
}

//////////////////////////////////////////////////
// GET ALL
//////////////////////////////////////////////////

func (s *airportService) GetAllAirports() ([]models.Airport, error) {
	return s.repo.FindAllAirports()
}
