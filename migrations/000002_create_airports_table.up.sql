CREATE TABLE airports (
    id          SERIAL PRIMARY KEY,
    code        VARCHAR(3) UNIQUE NOT NULL,
    city_name   VARCHAR(100) NOT NULL,
    airport_name VARCHAR(150) NOT NULL,
    country     VARCHAR(100) NOT NULL,
    created_at  TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMP NOT NULL DEFAULT NOW()
);
