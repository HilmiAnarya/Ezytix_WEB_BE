-- Pastikan tipe enum ada (jika belum ada di migrasi sebelumnya)
-- DROP TYPE IF EXISTS flight_class; 
-- CREATE TYPE flight_class AS ENUM ('economy', 'business', 'first_class');
-- (Note: Enum flight_class biasanya sudah ada di migrasi flight_classes, jadi kita pakai saja)

CREATE TABLE booking_details (
    id               SERIAL PRIMARY KEY,
    booking_id       INT NOT NULL REFERENCES bookings(id) ON DELETE CASCADE,

    passenger_name   VARCHAR(100) NOT NULL,
    
    -- Kita simpan kelas kursi di sini juga untuk snapshot tiket
    -- Gunakan VARCHAR biasa agar fleksibel atau ENUM flight_class jika ingin strict
    seat_class       VARCHAR(50) NOT NULL, 
    
    -- Harga SATUAN saat tiket dibeli (Snapshot Anti-Manipulasi)
    price            NUMERIC(15,2) NOT NULL,

    created_at       TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_booking_details_booking ON booking_details(booking_id);