CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE  TABLE clients (
	id SERIAL PRIMARY KEY,
	fullname TEXT NOT NULL,
    birth VARCHAR(10),  
	limit INTEGER NOT NULL,
    balance INTEGER NOT NULL DEFAULT 0,
    urubukey VARCHAR(10)
);

CREATE  TABLE transactions (
	id SERIAL PRIMARY KEY,
	client_id INTEGER NOT NULL,
	value INTEGER NOT NULL,
	kind CHAR(1) NOT NULL,
	description VARCHAR(10) NOT NULL,
	payor TEXT NOT NULL,
	payee TEXT NOT NULL,
	completed_at TIMESTAMP NOT NULL DEFAULT NOW(),
	CONSTRAINT fk_clients_transactions_id
		FOREIGN KEY (client_id) REFERENCES clients(id)
);
CREATE INDEX idx_clients_fullname_trgm ON clients USING gin (fullname gin_trgm_ops);


DO $$
BEGIN
	INSERT INTO clientes (name, limit)
	VALUES
		('o barato sai caro', 1000 * 100),
		('zan corp ltda', 800 * 100),
		('les cruders', 10000 * 100),
		('padaria joia de cocaia', 100000 * 100),
		('kid mais', 5000 * 100);
	
END;
$$;
