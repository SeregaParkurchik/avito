-- Active: 1739568803311@@127.0.0.1@5431@shop_test
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

CREATE TABLE employees (
    id SERIAL PRIMARY KEY,
    username VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    coins INT DEFAULT 1000,
    token VARCHAR(255) NOT NULL
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