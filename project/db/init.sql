-- User table
CREATE TABLE users
(
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) NOT NULL,
    first_name VARCHAR(255),
    last_name VARCHAR(255),
    password VARCHAR(255) NOT NULL,
    active INT NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);


-- Billing Address Table
CREATE TABLE billing_address
(
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    address VARCHAR(100) NOT NULL,
    address_2 VARCHAR(100),
    postal_code VARCHAR(10),
    city VARCHAR(25),
    country VARCHAR(25),
    FOREIGN KEY (user_id) REFERENCES users(id)
);


-- Plan table
CREATE TABLE plans
(
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    duration INT NOT NULL,
    duration_units VARCHAR(10) NOT NULL,
    billing_frequency INT NOT NULL,
    billing_frequency_units VARCHAR(20) NOT NULL,
    currency VARCHAR(5),
    price NUMERIC(10,2) NOT NULL
);

-- Subscription table
CREATE TABLE subscriptions
(
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    plan_id INT NOT NULL,
    contract_start_date TIMESTAMP NOT NULL,
    duration INT NOT NULL,
    duration_units VARCHAR(10) NOT NULL,
    billing_frequency INT NOT NULL,
    billing_frequency_units VARCHAR(20) NOT NULL,
    price NUMERIC(10,2) NOT NULL,
    currency VARCHAR(3) NOT NULL,
    product_code VARCHAR(20) NOT NULL,
    status varchar(20) NOT NULL,
    billed_cycles INT DEFAULT 0,
    next_billing_date TIMESTAMP NOT NULL,
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (plan_id) REFERENCES plans(id),
    UNIQUE (user_id, plan_id, product_code)
);


--Failed Invoices table
CREATE TABLE failed_invoices
(
    id SERIAL PRIMARY KEY,
    subscription_id INT NOT NULL,
    invoice_id VARCHAR(25) NOT NULL,
    invoice_date TIMESTAMP NOT NULL,
    email_retry INT NOT NULL,
    FOREIGN KEY (subscription_id) REFERENCES subscriptions(id)

);
