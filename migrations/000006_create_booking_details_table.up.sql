CREATE TABLE booking_details (
    id               SERIAL PRIMARY KEY,
    booking_id       INT NOT NULL REFERENCES bookings(id) ON DELETE CASCADE,

    -- [DATA PERSONAL]
    passenger_title  VARCHAR(10) NOT NULL,
    passenger_name   VARCHAR(255) NOT NULL,
    passenger_dob    DATE NOT NULL,
    
    -- [NEW] Tipe Penumpang (Snapshot dari DOB saat booking)
    -- Values: 'adult', 'child', 'infant' -> Frontend display jadi (Dewasa), (Anak)
    passenger_type   VARCHAR(20) NOT NULL, 

    nationality      VARCHAR(50) NOT NULL,
    
    -- [DATA DOKUMEN]
    passport_number  VARCHAR(50),
    issuing_country  VARCHAR(50),
    valid_until      DATE,

    -- [TIKET & HARGA]
    -- [NEW] Nomor Tiket Unik per Orang (Beda dengan Booking Code)
    ticket_number    VARCHAR(50) UNIQUE NOT NULL, 

    seat_class       VARCHAR(50) NOT NULL, 
    price            NUMERIC(15,2) NOT NULL,

    created_at       TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_booking_details_booking_id ON booking_details(booking_id);