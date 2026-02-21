-- Buat tabel user_otps
CREATE TABLE user_otps (
    id          BIGSERIAL PRIMARY KEY,
    user_id     BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    otp_code    VARCHAR(6) NOT NULL,
    expired_at  TIMESTAMPTZ NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Bikin Index biar pencarian kode OTP berdasarkan user_id berjalan super cepat
CREATE INDEX idx_user_otps_user_id ON user_otps(user_id);