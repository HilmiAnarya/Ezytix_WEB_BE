DROP TYPE IF EXISTS user_role CASCADE;

CREATE TYPE user_role AS ENUM ('customer', 'admin');

CREATE TABLE users (
    id          BIGSERIAL PRIMARY KEY,
    full_name   VARCHAR(255) NOT NULL,
    username    VARCHAR(255) NOT NULL,
    email       VARCHAR(255) UNIQUE,
    phone       VARCHAR(20) UNIQUE,
    password    VARCHAR(255) NOT NULL,
    role        user_role NOT NULL DEFAULT 'customer',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ
);