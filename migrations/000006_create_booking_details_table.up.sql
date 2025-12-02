DROP TYPE IF EXISTS flight_class CASCADE;

CREATE TYPE flight_class AS ENUM ('economy', 'business', 'first_class');

CREATE TABLE booking_details (
    id               SERIAL PRIMARY KEY,
    booking_id       INT NOT NULL REFERENCES bookings(id) ON DELETE CASCADE,

    passenger_name   VARCHAR(150) NOT NULL,
    seat_class       flight_class DEFAULT 'economy' NOT NULL,
    seat_number      VARCHAR(20),

    price_at_booking NUMERIC(15,2) NOT NULL,


    created_at       TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_booking_details_booking_id ON booking_details(booking_id);