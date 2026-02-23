CREATE TABLE booking_details (
    id               SERIAL PRIMARY KEY,
    booking_id       INT NOT NULL REFERENCES bookings(id) ON DELETE CASCADE,
    passenger_title  VARCHAR(10) NOT NULL,
    passenger_name   VARCHAR(255) NOT NULL,
    passenger_dob    DATE NOT NULL,
    passenger_type   VARCHAR(20) NOT NULL, 
    nationality      VARCHAR(50) NOT NULL,
    passport_number  VARCHAR(50),
    issuing_country  VARCHAR(50),
    valid_until      DATE,
    ticket_number    VARCHAR(50) UNIQUE NOT NULL, 
    seat_class       VARCHAR(50) NOT NULL, 
    price            NUMERIC(15,2) NOT NULL,
    created_at       TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_booking_details_booking_id ON booking_details(booking_id);