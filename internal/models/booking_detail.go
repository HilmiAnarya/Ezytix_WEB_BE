package models

type FlightClass string
const (
	Economy     FlightClass = "economy"
	Business    FlightClass = "business"
	FirstClass  FlightClass = "first_class"
)

type BookingDetail struct {
	ID         uint `json:"id" gorm:"primaryKey;autoIncrement"`
	BookingID  uint `json:"booking_id"`
	FlightLegID uint `json:"flight_leg_id"`

	PassengerName string      `json:"passenger_name"`
	SeatClass     FlightClass `json:"seat_class"`
	SeatNumber    string      `json:"seat_number"`
	PriceAtBooking float64    `json:"price_at_booking" gorm:"type:numeric(15,2)"`

	// RELATION (optional)
	FlightLeg FlightLeg `json:"flight_leg" gorm:"foreignKey:FlightLegID"`
}

func (BookingDetail) TableName() string {
	return "booking_details"
}
