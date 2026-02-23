package scheduler

import (
	"log"

	"github.com/robfig/cron/v3"
)

type ExpiredBookingProcessor interface {
	ProcessExpiredBookings() error
}

func StartCronJob(processor ExpiredBookingProcessor) {
	c := cron.New(cron.WithChain(
		cron.Recover(cron.DefaultLogger), 
	))
	_, err := c.AddFunc("@every 1m", func() {
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