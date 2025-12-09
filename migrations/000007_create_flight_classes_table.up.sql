CREATE TABLE flight_classes (
    id                   SERIAL PRIMARY KEY,
    flight_id            INT NOT NULL REFERENCES flights(id) ON DELETE CASCADE,
    seat_class           VARCHAR(50) NOT NULL, -- Economy, Business, First Class
    price                NUMERIC(15,2) NOT NULL, -- Harga per kelas kabin
    total_seats          INT NOT NULL DEFAULT 0,
    deleted_at           TIMESTAMP NULL,
    created_at           TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMP NOT NULL DEFAULT NOW()
);

