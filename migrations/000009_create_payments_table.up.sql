CREATE TABLE payments (
    id SERIAL PRIMARY KEY,
    
    -- [RELATION] Terhubung ke Booking
    order_id VARCHAR(50) NOT NULL, 
    
    -- [MIDTRANS CORE INFO]
    transaction_id     VARCHAR(100), -- ID Unik dari Midtrans (UUID)
    payment_type       VARCHAR(50),  -- 'bank_transfer', 'echannel', 'qris', 'gopay'
    gross_amount       NUMERIC(15,2) NOT NULL,
    transaction_status VARCHAR(50) NOT NULL DEFAULT 'pending', -- pending, settlement, expire, deny, cancel
    
    -- [BANK TRANSFER INFO] (BCA, BNI, BRI, Permata)
    bank        VARCHAR(20), -- 'bca', 'bni', 'bri'
    va_number   VARCHAR(50), 
    
    -- [MANDIRI BILL INFO] (Khusus 'echannel')
    bill_key    VARCHAR(50),
    biller_code VARCHAR(50),
    
    -- [QRIS & E-WALLET INFO]
    qr_url      TEXT, -- URL Gambar QR dari Midtrans
    deeplink    TEXT, -- Link redirect ke aplikasi (Gojek/Shopee)
    
    -- [TIMESTAMPS]
    expiry_time TIMESTAMP,    -- Waktu kedaluwarsa dari Midtrans
    paid_at     TIMESTAMP,    -- Diisi saat Webhook 'settlement' masuk
    created_at  TIMESTAMP DEFAULT NOW(),
    updated_at  TIMESTAMP DEFAULT NOW()
);

-- Indexes untuk mempercepat pencarian (terutama saat Webhook & Polling)
CREATE INDEX idx_payments_order_id ON payments(order_id);
CREATE INDEX idx_payments_transaction_id ON payments(transaction_id);