CREATE TABLE flights (
    id                   SERIAL PRIMARY KEY,
    flight_code          VARCHAR(100) UNIQUE, 
    airline_name         VARCHAR(100),

    origin_airport_id      INT NOT NULL REFERENCES airports(id),
    destination_airport_id INT NOT NULL REFERENCES airports(id),

    departure_time       TIMESTAMP NOT NULL,
    arrival_time         TIMESTAMP NOT NULL,

    total_duration       VARCHAR(50) NOT NULL,
    transit_count        INT NOT NULL DEFAULT 0,
    transit_info         VARCHAR(255),        

    deleted_at           TIMESTAMP NULL,
    created_at           TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMP NOT NULL DEFAULT NOW()
);