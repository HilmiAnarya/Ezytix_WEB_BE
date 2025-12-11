package booking

import (
	"time"
)

// ==========================================
// 1. INPUT (REQUEST)
// ==========================================

// Data Per Penumpang (Sesuai Screenshot UI: Paspor, Title, DOB)
type PassengerRequest struct {
	// Identitas Utama
	Title       string `json:"title" validate:"required,oneof=tuan nyonya nona mr ms mrs"`
	FullName    string `json:"full_name" validate:"required,min=2"`
	
	// Tanggal Lahir (Format YYYY-MM-DD)
	// Penting: Backend akan hitung umur dari sini untuk tentukan tipe (Adult/Child/Infant)
	DOB         string `json:"dob" validate:"required,datetime=2006-01-02"` 
	
	Nationality string `json:"nationality" validate:"required"`

	// Data Dokumen (Opsional di DTO, Wajib/Tidaknya divalidasi Service berdasarkan Rute)
	PassportNumber string `json:"passport_number"`
	IssuingCountry string `json:"issuing_country"`
	ValidUntil     string `json:"valid_until" validate:"omitempty,datetime=2006-01-02"`
}

// Item Booking (Mewakili Satu Penerbangan)
type BookingItemRequest struct {
	FlightID   uint               `json:"flight_id" validate:"required"`
	SeatClass  string             `json:"seat_class" validate:"required,oneof=economy business first_class"`
	
	// List Penumpang di penerbangan ini
	Passengers []PassengerRequest `json:"passengers" validate:"required,min=1,dive"` 
}

// Request Utama (Create Order)
// Bisa menampung Round Trip (2 Items) atau One Way (1 Item)
type CreateOrderRequest struct {
	Items []BookingItemRequest `json:"items" validate:"required,min=1,dive"`
}

// ==========================================
// 2. OUTPUT (RESPONSE)
// ==========================================

// Detail per Flight yang berhasil dibooking
type BookingDetailResponse struct {
	BookingCode     string    `json:"booking_code"` // PNR (EZY-XXX)
	FlightCode      string    `json:"flight_code"`  // No Pesawat (GA-123)
	Origin          string    `json:"origin"`
	Destination     string    `json:"destination"`
	DepartureTime   time.Time `json:"departure_time"`
	TotalPassengers int       `json:"total_passengers"`
	TotalPrice      float64   `json:"total_price"`
}

// Response Utama (Dikirim ke Frontend setelah Checkout)
type BookingResponse struct {
	OrderID         string    `json:"order_id"`
	TotalAmount     float64   `json:"total_amount"`
	Status          string    `json:"status"` // PENDING
	TransactionTime time.Time `json:"transaction_time"`
	
	// Informasi Pembayaran (Dari Xendit)
	PaymentURL      string    `json:"payment_url"` // Redirect user ke sini
	ExpiryDate      *time.Time `json:"expiry_date"`
	
	// Detail Penerbangan
	Bookings        []BookingDetailResponse `json:"bookings"`
}