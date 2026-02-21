package admin

type AdminService interface {
	GetDashboardStats() (*DashboardStatsResponse, error)
}

type adminService struct {
	repo AdminRepository
}

func NewAdminService(repo AdminRepository) AdminService {
	return &adminService{repo}
}

func (s *adminService) GetDashboardStats() (*DashboardStatsResponse, error) {
	customers, err := s.repo.CountCustomers()
	if err != nil {
		return nil, err
	}

	bookings, err := s.repo.CountBookingsToday()
	if err != nil {
		return nil, err
	}

	revenue, err := s.repo.SumRevenueToday()
	if err != nil {
		return nil, err
	}

	return &DashboardStatsResponse{
		CustomersRegistered: customers,
		FlightsBookedToday:  bookings,
		RevenueToday:        revenue,
	}, nil
}