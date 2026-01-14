package scheduler

import (
	"log"

	"github.com/robfig/cron/v3"
)

// Interface ini harus di-implement oleh BookingService
type ExpiredBookingProcessor interface {
	ProcessExpiredBookings() error
}

func StartCronJob(processor ExpiredBookingProcessor) {
	// Menggunakan Standard Chain agar log cron lebih rapi (Recover panic + Logger)
	c := cron.New(cron.WithChain(
		cron.Recover(cron.DefaultLogger), 
	))

	// Jalankan setiap 1 menit
	// Syntax: @every 1m
	_, err := c.AddFunc("@every 1m", func() {
		// Log start (opsional, bisa dimatikan kalau berisik)
		// log.Println("[SCHEDULER] Checking expired bookings...")

		if err := processor.ProcessExpiredBookings(); err != nil {
			log.Printf("❌ [SCHEDULER ERROR] Failed to process expired bookings: %v\n", err)
		}
	})

	if err != nil {
		log.Fatal("❌ [SCHEDULER] Failed to initialize Cron Job:", err)
	}

	c.Start()
	log.Println("✅ [SCHEDULER] Cron Job started: Strict Expiry Check active (Every 1 min)")
}