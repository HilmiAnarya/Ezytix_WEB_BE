package flight

import (
	"ezytix-be/internal/models"
	"ezytix-be/internal/modules/airline"
	"ezytix-be/internal/utils"
	"time"

	"github.com/shopspring/decimal" // [NEW] Import Library
)

// ==========================================
// 1. REQUEST DTO (Input dari User/Admin)
// ==========================================

type CreateFlightClassRequest struct {
	SeatClass  string          `json:"seat_class" validate:"required,oneof=economy business first_class"`
	Price      decimal.Decimal `json:"price" validate:"required"`
	TotalSeats int             `json:"total_seats" validate:"required,min=1"`
}

type CreateFlightLegRequest struct {
	LegOrder             int       `json:"leg_order" validate:"required,min=1"`
	AirlineID            uint      `json:"airline_id" validate:"required"`
	OriginAirportID      uint      `json:"origin_airport_id" validate:"required"`
	DestinationAirportID uint      `json:"destination_airport_id" validate:"required"`

	DepartureTime        time.Time `json:"departure_time" validate:"required"`
	ArrivalTime          time.Time `json:"arrival_time" validate:"required"`
	FlightNumber         string    `json:"flight_number" validate:"required"`
	
	TransitNotes         string    `json:"transit_notes"`
}

type CreateFlightRequest struct {
	FlightCode           string    `json:"flight_code" validate:"required"`
	AirlineID            uint   `json:"airline_id" validate:"required"`
	OriginAirportID      uint      `json:"origin_airport_id" validate:"required"`
	DestinationAirportID uint      `json:"destination_airport_id" validate:"required"`

	DepartureTime        time.Time `json:"departure_time" validate:"required"`
	ArrivalTime          time.Time `json:"arrival_time" validate:"required"`

	FlightLegs    []CreateFlightLegRequest   `json:"flight_legs" validate:"required,dive"`
	FlightClasses []CreateFlightClassRequest `json:"flight_classes" validate:"required,dive"`
}

type SearchFlightRequest struct {
	OriginAirportID      uint   `query:"origin"`
	DestinationAirportID uint   `query:"destination"`
	DepartureDate        string `query:"date"`
	SeatClass            string `query:"seat_class"`
	PassengerCount       int    `query:"passengers"`
}

// ==========================================
// 2. RESPONSE DTO (Output ke Frontend)
// ==========================================
// Kita WAJIB tambahkan ini agar Service Refactor nanti tidak error.

type FlightClassResponse struct {
	SeatClass  string          `json:"seat_class"`
	Price      decimal.Decimal `json:"price"`
	TotalSeats int             `json:"total_seats"`
}

type FlightLegResponse struct {
	ID            		 uint      `json:"id"`
	LegOrder             int       `json:"leg_order"`

	Airline       		 airline.AirlineSimpleResponse `json:"airline"`

	Origin        		 models.Airport `json:"origin"`
	Destination   		 models.Airport `json:"destination"`

	DepartureTime        time.Time `json:"departure_time"`
	ArrivalTime          time.Time `json:"arrival_time"`

	DurationMinutes   int    `json:"duration_minutes"`
	DurationFormatted string `json:"duration_formatted"` // "45m"

	// --- NEW FIELD: Info Transit setelah leg ini ---
	LayoverDurationMinutes   int    `json:"layover_duration_minutes,omitempty"`
	LayoverDurationFormatted string `json:"layover_duration_formatted,omitempty"`

	FlightNumber         string    `json:"flight_number"`
	TransitNotes  		string `json:"transit_notes"`
}

type FlightResponse struct {
	ID                   uint                  			`json:"id"`
	FlightCode           string                			`json:"flight_code"`
	Airline       		 airline.AirlineSimpleResponse  `json:"airline"`

	Origin        		 models.Airport `json:"origin"`
	Destination   		 models.Airport `json:"destination"`

	DepartureTime        time.Time             			`json:"departure_time"`
	ArrivalTime          time.Time             			`json:"arrival_time"`

	// Smart Backend: Kirim Int dan String
	TotalDuration    	 int    						`json:"total_duration_minutes"` 
	DurationFormatted 	 string 						`json:"duration_formatted"` // "1j 30m"

	TransitCount  		 int    						`json:"transit_count"`
	TransitInfo   		 string 						`json:"transit_info"`

	FlightLegs    []FlightLegResponse    `json:"flight_legs"`
	FlightClasses []models.FlightClass   `json:"flight_classes"`
}

// ToFlightResponse maps model to DTO with SAFETY CHECKS & LAYOVER LOGIC
func ToFlightResponse(f models.Flight) FlightResponse {
	// 1. Map Legs with Safety Checks & Layover Calculation
	var legResponses []FlightLegResponse
	
	// Gunakan index 'i' untuk mengintip leg berikutnya
	for i, leg := range f.FlightLegs {
		// Safety Check for Origin
		var origin models.Airport
		if leg.OriginAirport != nil {
			origin = *leg.OriginAirport
		}

		// Safety Check for Destination
		var destination models.Airport
		if leg.DestinationAirport != nil {
			destination = *leg.DestinationAirport
		}

		legResp := FlightLegResponse{
			ID:                leg.ID,
			LegOrder:          leg.LegOrder,
			Origin:            origin,
			Destination:       destination,
			DepartureTime:     leg.DepartureTime,
			ArrivalTime:       leg.ArrivalTime,
			FlightNumber:      leg.FlightNumber,
			TransitNotes:      leg.TransitNotes,
			DurationMinutes:   leg.Duration,
			DurationFormatted: utils.FormatDuration(leg.Duration),
		}

		// Safety Check for Airline
		if leg.Airline != nil {
			legResp.Airline = airline.AirlineSimpleResponse{
				ID:      leg.Airline.ID,
				IATA:    leg.Airline.IATA,
				Name:    leg.Airline.Name,
				LogoURL: leg.Airline.LogoURL,
			}
		}

		// --- LOGIC LAYOVER / TRANSIT TIME ---
		// Cek apakah ini bukan leg terakhir?
		if i < len(f.FlightLegs)-1 {
			nextLeg := f.FlightLegs[i+1]
			
			// Hitung selisih: Departure Leg Berikutnya - Arrival Leg Ini
			layoverMinutes := int(nextLeg.DepartureTime.Sub(leg.ArrivalTime).Minutes())
			
			if layoverMinutes > 0 {
				legResp.LayoverDurationMinutes = layoverMinutes
				legResp.LayoverDurationFormatted = utils.FormatDuration(layoverMinutes)
			}
		}

		legResponses = append(legResponses, legResp)
	}

	// 2. Map Main Flight
	var origin models.Airport
	if f.OriginAirport != nil {
		origin = *f.OriginAirport
	}

	var destination models.Airport
	if f.DestinationAirport != nil {
		destination = *f.DestinationAirport
	}

	res := FlightResponse{
		ID:                f.ID,
		FlightCode:        f.FlightCode,
		Origin:            origin,
		Destination:       destination,
		DepartureTime:     f.DepartureTime,
		ArrivalTime:       f.ArrivalTime,
		TransitCount:      f.TransitCount,
		TransitInfo:       f.TransitInfo,
		FlightLegs:        legResponses,
		FlightClasses:     f.FlightClasses,
		TotalDuration:     f.TotalDuration,
		DurationFormatted: utils.FormatDuration(f.TotalDuration),
	}

	// Safety check Airline Master
	if f.Airline != nil {
		res.Airline = airline.AirlineSimpleResponse{
			ID:      f.Airline.ID,
			IATA:    f.Airline.IATA,
			Name:    f.Airline.Name,
			LogoURL: f.Airline.LogoURL,
		}
	}

	return res
}