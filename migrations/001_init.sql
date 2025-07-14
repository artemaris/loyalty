CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    login TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT now()
    );

CREATE TABLE IF NOT EXISTS orders (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id),
    number TEXT UNIQUE NOT NULL,
    status TEXT NOT NULL,
    accrual NUMERIC DEFAULT 0,
    uploaded_at TIMESTAMP DEFAULT now()
    );

CREATE TABLE IF NOT EXISTS withdrawals (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id),
    order_number TEXT NOT NULL,
    amount NUMERIC NOT NULL,
    processed_at TIMESTAMP DEFAULT now()
    );