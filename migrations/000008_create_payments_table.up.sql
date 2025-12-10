CREATE TABLE payments (
    id              SERIAL PRIMARY KEY,
    
    -- Terhubung ke Booking via Order ID
    order_id        VARCHAR(50) NOT NULL, 
    
    -- Xendit Reference
    xendit_id       VARCHAR(100), -- ID dari Xendit (misal: invoice_id)
    payment_method  VARCHAR(50) DEFAULT 'QRIS',
    payment_status  VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    
    -- Data Uang
    amount          NUMERIC(15,2) NOT NULL,
    currency        VARCHAR(3) DEFAULT 'IDR',
    
    -- Data untuk Frontend (QR Code / Redirect Link)
    payment_url     TEXT, 
    
    paid_at         TIMESTAMP,
    created_at      TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Index agar pencarian status pembayaran cepat
CREATE INDEX idx_payments_order_id ON payments(order_id);
CREATE INDEX idx_payments_xendit_id ON payments(xendit_id);