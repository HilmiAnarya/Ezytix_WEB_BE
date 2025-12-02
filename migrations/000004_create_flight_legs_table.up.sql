CREATE TABLE flight_legs (
    id                    SERIAL PRIMARY KEY,
    flight_id             INT NOT NULL REFERENCES flights(id) ON DELETE CASCADE,

    leg_order             INT NOT NULL, -- 1,2,3,...

    departure_time        TIMESTAMP NOT NULL,
    arrival_time          TIMESTAMP NOT NULL,

    origin_airport_id     INT NOT NULL REFERENCES airports(id),
    destination_airport_id INT NOT NULL REFERENCES airports(id),

    flight_number         VARCHAR(50),    -- JT-763, GA-412, dll.
    airline_name          VARCHAR(100),
    airline_logo          VARCHAR(255),

    departure_terminal    VARCHAR(50),
    arrival_terminal      VARCHAR(50),

    duration              VARCHAR(50),    -- "1j 10m"

    transit_notes         VARCHAR(255),   -- "Transit di SUB 45m"

    deleted_at            TIMESTAMP NULL,
    created_at            TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at            TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Index untuk performa pencarian legs per flight
CREATE INDEX idx_flight_legs_flight_id ON flight_legs(flight_id);