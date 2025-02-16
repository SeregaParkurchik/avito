-- Active: 1739655257467@@127.0.0.1@5429@shop_test
CREATE TABLE employees (
    id SERIAL PRIMARY KEY,
    username VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    coins INT DEFAULT 1000,
    token VARCHAR(255) NOT NULL
);
CREATE TABLE transactions (
    id SERIAL PRIMARY KEY,
    employee_id INT NOT NULL,
    transaction_type VARCHAR(50) NOT NULL, -- 'received' или 'sent'
    amount INT NOT NULL,
    from_user VARCHAR(255), --  для отправителя
    to_user VARCHAR(255), --  для получателя
    FOREIGN KEY (employee_id) REFERENCES employees(id)
);