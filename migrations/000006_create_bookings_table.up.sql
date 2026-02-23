DROP TYPE IF EXISTS trip_type CASCADE;

CREATE TYPE trip_type AS ENUM ('one_way', 'round_trip');

CREATE TABLE bookings (
    id                  SERIAL PRIMARY KEY,
    order_id            VARCHAR(50) NOT NULL, 
    user_id             INT NOT NULL REFERENCES users(id),
    flight_id           INT NOT NULL REFERENCES flights(id),
    booking_code        VARCHAR(20) UNIQUE NOT NULL,
    trip_type           trip_type DEFAULT 'one_way' NOT NULL,
    total_passengers    INT NOT NULL,
    total_price         NUMERIC(15,2) NOT NULL,
    status              VARCHAR(20) NOT NULL DEFAULT 'pending',
    expired_at          TIMESTAMP,
    created_at          TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_bookings_user_id ON bookings(user_id);
CREATE INDEX idx_bookings_order_id ON bookings(order_id);
CREATE INDEX idx_bookings_booking_code ON bookings(booking_code);
CREATE INDEX idx_bookings_expired_at ON bookings(expired_at);