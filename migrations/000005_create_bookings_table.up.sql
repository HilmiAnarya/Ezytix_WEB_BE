DROP TYPE IF EXISTS trip_type CASCADE;

CREATE TYPE trip_type AS ENUM ('one_way', 'round_trip');

CREATE TABLE bookings (
    id                  SERIAL PRIMARY KEY,

    user_id             INT NOT NULL REFERENCES users(id),
    flight_id           INT NOT NULL REFERENCES flights(id),

    booking_code        VARCHAR(50) UNIQUE NOT NULL,
    booking_date        TIMESTAMP NOT NULL DEFAULT NOW(),

    trip_type           trip_type DEFAULT 'one_way' NOT NULL,

    total_passengers    INT NOT NULL,

    base_price_snapshot NUMERIC(15,2) NOT NULL,
    total_price         NUMERIC(15,2) NOT NULL,

    status              VARCHAR(30) NOT NULL, -- pending, paid, cancelled

    
    created_at          TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Optional index untuk pencarian riwayat user
CREATE INDEX idx_bookings_user_id ON bookings(user_id);