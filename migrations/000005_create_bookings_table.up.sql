CREATE TABLE bookings (
    id                  SERIAL PRIMARY KEY,
    user_id             INT NOT NULL REFERENCES users(id),
    flight_id           INT NOT NULL REFERENCES flights(id), -- Parent Flight
    
    booking_code        VARCHAR(20) UNIQUE NOT NULL, -- Contoh: EZY-882X (Short & Unique)
    
    total_passengers    INT NOT NULL,
    total_price         NUMERIC(15,2) NOT NULL, -- Total bayar (Snapshot backend)
    status              VARCHAR(20) NOT NULL DEFAULT 'pending', -- pending, paid, cancelled, failed
    
    created_at          TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Index untuk mempercepat query history user
CREATE INDEX idx_bookings_user ON bookings(user_id);
-- Index untuk mempercepat lookup booking code
CREATE INDEX idx_bookings_code ON bookings(booking_code);