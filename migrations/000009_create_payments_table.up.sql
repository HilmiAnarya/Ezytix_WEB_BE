CREATE TABLE payments (
    id                  SERIAL PRIMARY KEY,
    
    -- [THE GLUE] Terhubung ke Booking via Order ID
    order_id            VARCHAR(50) NOT NULL, 
    
    -- [XENDIT INFO]
    xendit_id           VARCHAR(100), -- ID dari response Xendit (Payment Request ID / VA ID)
    payment_method      VARCHAR(50),  -- Contoh: 'BCA', 'MANDIRI', 'OVO', 'QRIS'
    payment_channel     VARCHAR(50),  -- [NEW] Contoh: 'VIRTUAL_ACCOUNT', 'QR_CODE', 'E_WALLET'
    payment_status      VARCHAR(20) NOT NULL DEFAULT 'PENDING', -- PENDING, PAID, EXPIRED, FAILED
    
    -- [TRANSACTION INFO]
    amount              NUMERIC(15,2) NOT NULL,
    currency            VARCHAR(3) DEFAULT 'IDR',
    
    -- [FRONTEND DISPLAY INFO]
    payment_code        VARCHAR(50),  -- [NEW] Menyimpan Nomor VA (Contoh: 880812345678)
    qr_string           TEXT,         -- [NEW] Menyimpan String QR Raw (untuk generate QR Code)
    payment_url         TEXT,         -- Menyimpan Redirect URL (untuk E-Wallet / Invoice)
    expiry_time         TIMESTAMP,    -- [NEW] Waktu kedaluwarsa pembayaran spesifik ini
    
    paid_at             TIMESTAMP,    -- Kapan user bayar (diisi saat Webhook)
    
    created_at          TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Index agar pencarian cepat saat Webhook masuk
CREATE INDEX idx_payments_order_id ON payments(order_id);
CREATE INDEX idx_payments_xendit_id ON payments(xendit_id);