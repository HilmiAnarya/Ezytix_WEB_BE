CREATE TABLE payments (
    id              SERIAL PRIMARY KEY,
    
    -- [THE GLUE] Terhubung ke Booking via Order ID
    order_id        VARCHAR(50) NOT NULL, 
    
    -- [XENDIT INFO]
    xendit_id       VARCHAR(100), -- Invoice ID dari Xendit (Penting untuk Webhook)
    payment_method  VARCHAR(50) DEFAULT 'QRIS',
    payment_status  VARCHAR(20) NOT NULL DEFAULT 'PENDING', -- PENDING, PAID, EXPIRED, FAILED
    
    -- [TRANSACTION INFO]
    amount          NUMERIC(15,2) NOT NULL,
    currency        VARCHAR(3) DEFAULT 'IDR',
    
    -- [FRONTEND INFO] QR String / Redirect URL
    payment_url     TEXT, 
    
    paid_at         TIMESTAMP, -- Kapan user bayar (diisi saat Webhook)
    
    created_at      TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Index agar pencarian cepat saat Webhook masuk
CREATE INDEX idx_payments_order_id ON payments(order_id);
CREATE INDEX idx_payments_xendit_id ON payments(xendit_id);