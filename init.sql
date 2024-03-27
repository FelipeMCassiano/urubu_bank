CjEATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE  TABLE clients (
	id SERIAL PRIMARY KEY,
	fullname TEXT NOT NULL UNIQUE,
    birth VARCHAR(10) NOT NULL,  
    password TEXT NOT NULL,
	credit_limit INTEGER NOT NULL,
    balance INTEGER DEFAULT 0,
    urubukey  TEXT 
);

CREATE  TABLE transactions (
	id SERIAL PRIMARY KEY,
	client_id INTEGER NOT NULL,
	value INTEGER NOT NULL,
	kind CHAR(1) NOT NULL,
	description VARCHAR(10) NOT NULL,
	payee TEXT NOT NULL,
	completed_at TIMESTAMP NOT NULL DEFAULT NOW(),
	CONSTRAINT fk_clients_transactions_id
		FOREIGN KEY (client_id) REFERENCES clients(id)
);
-- CREATE INDEX idx_clients_fullname_trgm ON clients USING gin (fullname gin_trgm_ops);

CREATE INDEX idx_fullname_trgm ON clients USING gin (fullname gin_trgm_ops);
