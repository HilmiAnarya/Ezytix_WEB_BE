package airline

import "time"

// Request untuk Create/Update
type CreateAirlineRequest struct {
	IATA    string `json:"iata" validate:"required,len=2,uppercase"` // Validasi kode IATA 2 huruf
	Name    string `json:"name" validate:"required,min=3"`
	LogoURL string `json:"logo_url" validate:"omitempty,url"`
}

type UpdateAirlineRequest struct {
	Name    string `json:"name,omitempty"`
	LogoURL string `json:"logo_url,omitempty" validate:"omitempty,url"`
	IATA    string `json:"iata" validate:"required,len=2,uppercase"`
}

// Response (Single Object)
type AirlineResponse struct {
	ID        uint      `json:"id"`
	IATA      string    `json:"iata"`
	Name      string    `json:"name"`
	LogoURL   string    `json:"logo_url"`
	CreatedAt time.Time `json:"created_at"`
}

// Response (Simple/Minimalist untuk embedding di Flight)
type AirlineSimpleResponse struct {
	ID      uint   `json:"id"`
	IATA    string `json:"iata"`
	Name    string `json:"name"`
	LogoURL string `json:"logo_url"`
}