package airport

type CreateAirportRequest struct {
	Code        string `json:"code" validate:"required,len=3"`        
	CityName    string `json:"city_name" validate:"required"`              
	AirportName string `json:"airport_name" validate:"required"`         
	Country     string `json:"country" validate:"required"`               
}

type UpdateAirportRequest struct {
	Code 		*string `json:"code"`
	CityName    *string `json:"city_name"`     
	AirportName *string `json:"airport_name"` 
	Country     *string `json:"country"`      
}
