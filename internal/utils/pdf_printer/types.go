package pdfprinter

type InvoiceItem struct {
    Number      string
    Product     string
    Description string
    Quantity    int
    Total       string
}

type Passenger struct {
    Name string
    Type string
}

type InvoiceData struct {
    HeaderImage string 
    FooterImage string
    InvoiceNumber string
    Date          string
    CustomerName  string
    CustomerEmail string
    CustomerPhone string
    PaymentMethod string
    PaymentStatus string
    Passengers []Passenger
    Items      []InvoiceItem
    SubTotal   string
    ServiceFee string
    GrandTotal string
}

type FlightPoint struct {
    Date        string
    Time        string
	CityName	string
    CityCode    string
    AirportName string 
}

type TransitDetail struct {
    IsTransit bool
    Location  string 
    Duration  string 
}

type FlightSegment struct {
    AirlineName   string
    AirlineLogo   string
    FlightNumber  string
    FlightClass   string
    Departure     FlightPoint
    Arrival       FlightPoint
    Duration      string
    Transit       TransitDetail 
}

type TicketPassenger struct {
    Number       int
    Name         string
    TicketNumber string
}

type TicketData struct {
    HeaderImage string 
    FooterImage string 
    BookingID   string
    BookingCode string
    BookingDate string
    BookerName  string
    QRCode      string
    Segments    []FlightSegment
    Passengers  []TicketPassenger
}