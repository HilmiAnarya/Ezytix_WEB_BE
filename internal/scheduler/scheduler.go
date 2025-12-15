package scheduler

import (
	"log"

	"github.com/robfig/cron/v3"
)

type ExpiredBookingProcessor interface {
	ProcessExpiredBookings() error
}

func StartCronJob(processor ExpiredBookingProcessor) {
	c := cron.New()

	// "@every 1m" adalah syntax robfig/cron untuk interval 1 menit.
	_, err := c.AddFunc("@every 1m", func() {
		// Logika yang dijalankan tiap menit:
		if err := processor.ProcessExpiredBookings(); err != nil {
			log.Printf("[CRON ERROR] Gagal memproses booking expired: %v\n", err)
		}
	})

	if err != nil {
		log.Fatal("Gagal menginisialisasi Cron Job:", err)
	}

	c.Start()
	log.Println("✅ [SCHEDULER] Cron Job started: Checking expired bookings every 1 minute...")
}