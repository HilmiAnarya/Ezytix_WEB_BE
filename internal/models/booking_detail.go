package models

import "time"

const (
	PassengerTypeAdult  = "adult"
	PassengerTypeChild  = "child"
	PassengerTypeInfant = "infant"
)

type BookingDetail struct {
	ID uint `json:"id" gorm:"primaryKey;autoIncrement"`

	// Link ke Parent
	BookingID uint `json:"booking_id" gorm:"not null"`

	// ==============================
	// 1. IDENTITAS PERSONAL
	// ==============================
	PassengerTitle string    `json:"passenger_title" gorm:"size:10;not null"` // tuan, nyonya, nona
	PassengerName  string    `json:"passenger_name" gorm:"size:255;not null"`
	PassengerDOB   time.Time `json:"passenger_dob" gorm:"type:date;not null"`
	
	// [NEW] Kategori Penumpang (Snapshot dari DOB saat booking)
	PassengerType  string    `json:"passenger_type" gorm:"size:20;not null"` // adult, child, infant

	Nationality    string    `json:"nationality" gorm:"size:50;not null"`

	// ==============================
	// 2. DOKUMEN (PASPOR / KTP)
	// ==============================
	// Gunakan Pointer (*) agar bisa menyimpan NULL di database (Opsional)
	PassportNumber *string    `json:"passport_number" gorm:"size:50"`
	IssuingCountry *string    `json:"issuing_country" gorm:"size:50"`
	ValidUntil     *time.Time `json:"valid_until" gorm:"type:date"`

	// ==============================
	// 3. TIKET & HARGA
	// ==============================
	// [NEW] Nomor Tiket Unik per Orang
	TicketNumber string `json:"ticket_number" gorm:"size:50;uniqueIndex;not null"`

	// Snapshot Transaksi
	SeatClass string  `json:"seat_class" gorm:"size:50;not null"` // economy, business
	Price     float64 `json:"price" gorm:"type:numeric(15,2);not null"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (BookingDetail) TableName() string {
	return "booking_details"
}