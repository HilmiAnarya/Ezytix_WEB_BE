package pdfprinter

// Invoice

type InvoiceItem struct {
    Number      string // 1, 2, atau kosong
    Product     string // "Tiket Pesawat"
    Description string // "Lion Air CGK - DPS..."
    Quantity    int
    Total       string // "3.000.000" (Sudah diformat string biar aman)
}

type Passenger struct {
    Name string
    Type string // "(DEWASA)" atau "(ANAK)"
}

type InvoiceData struct {
    // Aset Gambar (Base64 String)
    HeaderImage string 
    FooterImage string 

    // Metadata Invoice
    InvoiceNumber string
    Date          string // "30 Oktober 2025, 17:59 Kamis"

    // Data Pemesan
    CustomerName  string
    CustomerEmail string
    CustomerPhone string

    // Data Pembayaran
    PaymentMethod string
    PaymentStatus string

    // Data Penumpang
    Passengers []Passenger

    // Data Belanja
    Items      []InvoiceItem
    SubTotal   string
    ServiceFee string
    GrandTotal string
}

// E-Ticket

// --- E-TICKET DATA STRUCTURES ---

// 1. Data Titik Penerbangan (Keberangkatan/Kedatangan)
type FlightPoint struct {
    Date        string // "5 November 2025"
    Time        string // "09:00"
	CityName	string
    CityCode    string // "DPS"
    AirportName string // "I Gusti Ngurah Rai..."
}

// 2. Data Transit (Opsional)
type TransitDetail struct {
    IsTransit bool   // true jika ada transit setelah segment ini
    Location  string // "Jakarta"
    Duration  string // "2 Jam"
}

// 3. Data Segmen Penerbangan (Satu kali terbang)
type FlightSegment struct {
    AirlineName   string // "Lion Air"
    AirlineLogo   string // URL atau Base64
    FlightNumber  string // "JT-763"
    FlightClass   string // "Economy (W)"
    
    Departure     FlightPoint
    Arrival       FlightPoint
    Duration      string // "2 J 30 M"
    
    // Info Transit SETELAH penerbangan ini (jika ada)
    Transit       TransitDetail 
}

// 4. Data Penumpang Tiket
type TicketPassenger struct {
    Number       int    // 1, 2, 3
    Name         string // "MR. Hilmi Anarya (Dewasa)"
    TicketNumber string // "0000000001"
}

// 5. ROOT DATA: E-Ticket
type TicketData struct {
    // Aset Gambar (Base64)
    HeaderImage string 
    FooterImage string 
    
    // Booking Info
    BookingID   string // "00000001"
    BookingCode string // "SVBNMS" (PNR)
    BookingDate string // "30 Oktober 2025, 17:59"
    BookerName  string // "Hilmi Anarya"
    QRCode      string // Base64 Image String dari PNR
    
    // Flight Info (Slice karena bisa Direct atau Multi-Transit)
    Segments    []FlightSegment
    
    // Passenger Info
    Passengers  []TicketPassenger
}