CREATE TABLE flight_legs (
    id                    SERIAL PRIMARY KEY,
    flight_id             INT NOT NULL REFERENCES flights(id) ON DELETE CASCADE,

    leg_order             INT NOT NULL,
    airline_id             INT REFERENCES airlines(id),

    departure_time        TIMESTAMP NOT NULL,
    arrival_time          TIMESTAMP NOT NULL,

    origin_airport_id     INT NOT NULL REFERENCES airports(id),
    destination_airport_id INT NOT NULL REFERENCES airports(id),

    flight_number         VARCHAR(50),

    duration              INT,

    transit_notes         VARCHAR(255), 

    deleted_at            TIMESTAMP NULL,
    created_at            TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at            TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Index untuk performa pencarian legs per flight
CREATE INDEX idx_flight_legs_flight_id ON flight_legs(flight_id);