CREATE TABLE payments (
    id SERIAL PRIMARY KEY,
    order_id VARCHAR(50) NOT NULL,
    transaction_id     VARCHAR(100),
    payment_type       VARCHAR(50),
    gross_amount       NUMERIC(15,2) NOT NULL,
    transaction_status VARCHAR(50) NOT NULL DEFAULT 'pending',
    bank        VARCHAR(20),
    va_number   VARCHAR(50), 
    bill_key    VARCHAR(50),
    biller_code VARCHAR(50),
    qr_url      TEXT,
    deeplink    TEXT,
    expiry_time TIMESTAMP,  
    paid_at     TIMESTAMP,    
    created_at  TIMESTAMP DEFAULT NOW(),
    updated_at  TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_payments_order_id ON payments(order_id);
CREATE INDEX idx_payments_transaction_id ON payments(transaction_id);