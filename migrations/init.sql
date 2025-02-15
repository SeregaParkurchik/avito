-- Active: 1739191328165@@127.0.0.1@5432@shop
CREATE TABLE employees (
    id SERIAL PRIMARY KEY,
    username VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    coins INT DEFAULT 1000,
    token VARCHAR(255) NOT NULL
);

CREATE TABLE merchandise (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) UNIQUE NOT NULL,
    price INT NOT NULL
);

INSERT INTO merchandise (name, price) VALUES
('t-shirt', 80),
('cup', 20),
('book', 50),
('pen', 10),
('powerbank', 200),
('hoody', 300),
('umbrella', 200),
('socks', 10),
('wallet', 50),
('pink-hoody', 500);

CREATE TABLE transactions (
    id SERIAL PRIMARY KEY,
    employee_id INT NOT NULL,
    transaction_type VARCHAR(50) NOT NULL, -- 'received' или 'sent'
    amount INT NOT NULL,
    from_user VARCHAR(255), --  для отправителя
    to_user VARCHAR(255), --  для получателя
    FOREIGN KEY (employee_id) REFERENCES employees(id)
);

CREATE TABLE inventory (
    id SERIAL PRIMARY KEY,
    employee_id INT NOT NULL,
    merchandise_id INT NOT NULL,
    quantity INT DEFAULT 1,
    FOREIGN KEY (employee_id) REFERENCES employees(id),
    FOREIGN KEY (merchandise_id) REFERENCES merchandise(id),
    UNIQUE (employee_id, merchandise_id) -- гарантирует, что один и тот же товар не может быть добавлен несколько раз для одного пользователя
);
