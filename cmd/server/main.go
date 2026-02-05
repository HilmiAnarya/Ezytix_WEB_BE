package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"ezytix-be/internal/config"
	"ezytix-be/internal/server"
	
	// Pastikan alias ini benar jika nama package di dalam foldernya berbeda
	pdfprinter "ezytix-be/internal/utils/pdf_printer"

	"github.com/joho/godotenv"
)

func main() {
	log.Println("🖨️  Mulai Test Printer Engine...")

	// --- 1. SETUP PATH ASSETS (PERBAIKAN UTAMA DISINI) ---
	// Kita cari tahu dulu kita sedang jalan di folder mana
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("❌ Gagal mendapatkan working directory: %v", err)
	}

	// Tentukan lokasi folder assets secara absolut
	// Asumsi: folder 'assets' ada di root project (sejajar dengan main.go/go.mod)
	assetsPath := filepath.Join(cwd, "internal", "assets", "images")
	
	fmt.Printf("📂 Mencari gambar di folder: %s\n", assetsPath)

	// Helper function biar kodingan rapi & error tertangkap
	loadImage := func(filename string) string {
		fullPath := filepath.Join(assetsPath, filename)
		base64Str, err := pdfprinter.ImageToBase64(fullPath)
		if err != nil {
			// LANGSUNG STOP PROGRAM JIKA GAMBAR TIDAK KETEMU
			// Supaya kamu sadar kalau path-nya salah
			log.Fatalf("❌ FATAL: Gagal load gambar '%s'. \n   Cek apakah file ada di: %s\n   Error: %v", filename, fullPath, err)
		}
		fmt.Printf("✅ Berhasil load: %s\n", filename)
		return base64Str
	}

	// --- 2. LOAD GAMBAR (Dengan Error Checking) ---
	// Perhatikan nama file harus SAMA PERSIS (termasuk huruf besar/kecil)
	invHeader := loadImage("invoice_header.png")
	invFooter := loadImage("invoice_footer.png")
	
	// Cek nama file tiket kamu, sesuaikan dengan yang ada di folder assets
	// Di chat sebelumnya kamu pakai "ticket_header.png", di main.go kamu pakai "eticket_header.png"
	// Saya pakai "ticket_header.png" sesuai yang kamu upload sebelumnya. 
	// GANTI JIKA NAMA FILE DI KOMPUTERMU BEDA.
	ticketHeader := loadImage("eticket_header.png") 
	ticketFooter := loadImage("eticket_footer.png")

	lion := loadImage("Lion.png")
	
	dummyQR := ticketHeader 

	// --- 3. TEST INVOICE ---
	invoiceData := pdfprinter.InvoiceData{
		HeaderImage:   invHeader,
		FooterImage:   invFooter,
		InvoiceNumber: "INV/2025/11/001",
		Date:          "05 November 2025",
		CustomerName:  "Hilmi Anarya",
		CustomerEmail: "hilmi@example.com",
		CustomerPhone: "08123456789",
		PaymentMethod: "BCA Virtual Account",
		PaymentStatus: "LUNAS",
		Passengers: []pdfprinter.Passenger{
			{Name: "Hilmi Anarya", Type: "(DEWASA)"},
			{Name: "Rizky Oryza", Type: "(ANAK)"},
		},
		Items: []pdfprinter.InvoiceItem{
			{Number: "1", Product: "Tiket Pesawat", Description: "Lion Air CGK-DPS (Dewasa)", Quantity: 1, Total: "1.500.000"},
			{Number: "", Product: "", Description: "Lion Air CGK-DPS (Anak)", Quantity: 1, Total: "1.000.000"},
		},
		SubTotal:   "2.500.000",
		ServiceFee: "20.000",
		GrandTotal: "2.520.000",
	}

	// Gunakan path absolut juga untuk template
	templatePathInvoice := "invoice.html" // Pastikan logic di GeneratePDF membacanya dari folder templates
	err = pdfprinter.GeneratePDF(templatePathInvoice, invoiceData, "test_invoice_final.pdf")
	if err != nil {
		log.Fatalf("❌ Error Generate Invoice: %v", err)
	}
	log.Println("✅ Invoice PDF berhasil dibuat!")

	var manyPassengers []pdfprinter.TicketPassenger
    for i := 1; i <= 40; i++ {
        manyPassengers = append(manyPassengers, pdfprinter.TicketPassenger{
            Number:       i,
            Name:         fmt.Sprintf("Penumpang Ke-%d Yang Sangat Berharga", i),
            TicketNumber: fmt.Sprintf("TIX-0000%d", i),
        })
    }

	// --- 4. TEST E-TICKET ---
	ticketData := pdfprinter.TicketData{
		HeaderImage: ticketHeader,
		FooterImage: ticketFooter,
		BookingID:   "ORD-9991",
		BookingCode: "SVBNMS",
		BookingDate: "30 Oktober 2025, 17:59",
		BookerName:  "Hilmi Anarya",
		QRCode:      dummyQR,
		
		Segments: []pdfprinter.FlightSegment{
			// SEGMENT 1: CGK -> DPS
			{
				AirlineName:  "Lion Air",
				AirlineLogo:  lion,
				FlightNumber: "JT-763",
				FlightClass:  "Economy (W)",
				Departure: pdfprinter.FlightPoint{
					Date: "5 Nov 2025", Time: "09:00", CityName: "Jakarta", CityCode: "CGK", AirportName: "Soekarno Hatta Intl",
				},
				Arrival: pdfprinter.FlightPoint{
					Date: "5 Nov 2025", Time: "11:30", CityName: "Bali", CityCode: "DPS", AirportName: "Ngurah Rai Intl",
				},
				Duration: "2 J 30 M",
				Transit: pdfprinter.TransitDetail{IsTransit: true, Location: "Denpasar", Duration: "2 Jam"},
			},
			// SEGMENT 2: DPS -> LOP
			{
				AirlineName:  "Wings Air",
				AirlineLogo:  lion,
				FlightNumber: "IW-1850",
				FlightClass:  "Economy (Y)",
				Departure: pdfprinter.FlightPoint{
					Date: "5 Nov 2025", Time: "13:30", CityName: "Bali", CityCode: "DPS", AirportName: "Ngurah Rai Intl",
				},
				Arrival: pdfprinter.FlightPoint{
					Date: "5 Nov 2025", Time: "14:15", CityName: "Lombok", CityCode: "LOP", AirportName: "Zainuddin Abdul Madjid",
				},
				Duration: "45 M",
			},
		},

		Passengers: manyPassengers,
	}

	err = pdfprinter.GeneratePDF("ticket.html", ticketData, "test_ticket_final.pdf")
	if err != nil {
		log.Fatalf("❌ Error Generate Ticket: %v", err)
	}
	log.Println("✅ Ticket PDF berhasil dibuat!")

	// --- 5. START SERVER (KODE ASLI KAMU) ---
	godotenv.Load()
	config.LoadConfig()

	srv := server.New()
	srv.RegisterRoutes()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("Server running on port %s", config.AppConfig.Port)
		if err := srv.Listen(":" + config.AppConfig.Port); err != nil {
			log.Printf("Server error: %v", err)
		}
	}()

	<-signalChan
	log.Println("Received shutdown signal...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.ShutdownWithContext(ctx); err != nil {
		log.Println("Forced shutdown:", err)
	}

	if err := srv.DB.Close(); err != nil {
		log.Println("Error closing DB:", err)
	}

	log.Println("Server exited gracefully.")
}