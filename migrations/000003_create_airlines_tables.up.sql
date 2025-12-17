CREATE TABLE airlines (
    id SERIAL PRIMARY KEY,
    IATA VARCHAR(10) UNIQUE NOT NULL, -- IATA Code (e.g., GA, JT, QZ)
    name VARCHAR(100) NOT NULL,       -- e.g., Garuda Indonesia
    logo_url TEXT,                    -- URL Logo for Frontend
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);