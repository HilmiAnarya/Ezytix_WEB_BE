package airport

// ==============================
// REQUEST DTO
// ==============================

// CreateAirportRequest adalah struktur data untuk membuat airport baru
type CreateAirportRequest struct {
	Code        string `json:"code" validate:"required,len=3"`            // CGK
	CityName    string `json:"city_name" validate:"required"`              // Jakarta
	AirportName string `json:"airport_name" validate:"required"`          // Soekarno-Hatta International Airport
	Country     string `json:"country" validate:"required"`               // Indonesia
}

// UpdateAirportRequest adalah struktur data untuk update airport
type UpdateAirportRequest struct {
	Code 		*string `json:"code"`
	CityName    *string `json:"city_name"`     // optional
	AirportName *string `json:"airport_name"`  // optional
	Country     *string `json:"country"`       // optional
}
