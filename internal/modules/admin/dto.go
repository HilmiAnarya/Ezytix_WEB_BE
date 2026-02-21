package admin

type DashboardStatsResponse struct {
	CustomersRegistered int64   `json:"customers_registered"`
	FlightsBookedToday  int64   `json:"flights_booked_today"`
	RevenueToday        float64 `json:"revenue_today"`
}