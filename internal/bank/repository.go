package bank

import (
	"context"
	"errors"

	"github.com/FelipeMCassiano/urubu_bank/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sethvargo/go-password/password"
)

type Respository interface {
	CreateTransaction(ctx context.Context, t domain.Transaction) (domain.TransactionResponse, error)
	GetBankStatement(ctx context.Context, id int) (domain.BankStatemant, error)
	GenerateUrubuKey(ctx context.Context, id int) (domain.UrubuKey, error)
	SearchClientByName(ctx context.Context, name string) (domain.CostumerConsult, error)
	CreateNewAccount(ctx context.Context, client domain.CreateCostumer) (domain.CreatedCostumer, error)
	VerifyIfClientExists(ctx context.Context, id int) (string, error)
}

type repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) Respository {
	return &repository{
		db: db,
	}
}

var (
	ErrNotFound = errors.New("client not found")
	LimitErr    = errors.New("limit error")
)

func (r *repository) SearchClientByName(ctx context.Context, name string) (domain.CostumerConsult, error) {
	result, err := r.db.Query(ctx, "SElECT fullname, urubukey FROM client WHERE similarity(fullname, %s) > 0.6", name)
	if err != nil {
		return domain.CostumerConsult{}, err
	}

	client, err := pgx.CollectOneRow(result, pgx.RowToStructByPos[domain.CostumerConsult])
	if err != nil {
		return domain.CostumerConsult{}, err
	}

	return client, nil
}

func (r *repository) VerifyIfClientExists(ctx context.Context, id int) (string, error) {
	var clientxists string
	err := r.db.QueryRow(ctx, "SELECT fullname FROM clients WHERE id=$1", id).Scan(&clientxists)
	if err != nil {
		return "", err
	}

	return clientxists, nil
}

func (r *repository) CreateTransaction(ctx context.Context, t domain.Transaction) (domain.TransactionResponse, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return domain.TransactionResponse{}, err
	}
	defer tx.Rollback(ctx)
	var limit, balance, newbalance int

	err = tx.QueryRow(context.Background(), "SELECT limit, balance FROM clients WHERE name=$1", t.Payor).Scan(&limit, &balance)
	if err != nil {
		return domain.TransactionResponse{}, err
	}

	if t.Kind == "c" {
		newbalance = t.Value + balance
	} else {
		newbalance = t.Value - balance
	}

	if (newbalance + limit) < 0 {
		return domain.TransactionResponse{}, LimitErr
	}

	var Payee string

	err = tx.QueryRow(context.Background(), "SELECT fullname FROM clients WHERE urubukey=$1", t.PayeeUrubuKey).Scan(&Payee)
	if err != nil {
		return domain.TransactionResponse{}, ErrNotFound
	}

	batch := &pgx.Batch{}

	batch.Queue("INSERT INTO transactions (client_id, value, kind, description, payee, completed_at, payor) VALUES($1,$2,$3,$4,$5,$6, $7)",
		t.Client_Id, t.Value, t.Kind, t.Description, Payee, t.Completed_at, t.Payor)
	batch.Queue("UPDATE clients SET balance WHERE urubukey =$1", t.PayeeUrubuKey)
	sendBatch := tx.SendBatch(context.Background(), batch)
	_, err = sendBatch.Exec()
	if err != nil {
		return domain.TransactionResponse{}, err
	}
	if err := sendBatch.Close(); err != nil {
		return domain.TransactionResponse{}, err
	}

	err = tx.Commit(context.Background())
	if err != nil {
		if err.Error() == "no rows in result set" {
			return domain.TransactionResponse{}, ErrNotFound
		}
		return domain.TransactionResponse{}, err
	}

	response := domain.TransactionResponse{
		Value:        t.Value,
		Kind:         t.Kind,
		Payor:        t.Payor,
		Payee:        Payee,
		Completed_at: t.Completed_at,
	}

	return response, nil
}

func (r *repository) CreateNewAccount(ctx context.Context, client domain.CreateCostumer) (domain.CreatedCostumer, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return domain.CreatedCostumer{}, err
	}
	var fullaname, birth string
	var limit, id int

	err = tx.QueryRow(context.Background(), "INSERT INTO clients (fullname, birth, limit,password) VALUES ($1,$2,$3,$4) RETURING id, fullname, birth, limit",
		client.Fullname, client.Birth, client.Limit, client.Password).Scan(&id, &fullaname, &birth, &limit)
	if err != nil {
		return domain.CreatedCostumer{}, err
	}
	if err := tx.Commit(context.Background()); err != nil {
		return domain.CreatedCostumer{}, err
	}
	createdClient := domain.CreatedCostumer{
		ID:       id,
		Fullname: fullaname,
		Limit:    limit,
	}

	return createdClient, nil
}

func (r *repository) GenerateUrubuKey(ctx context.Context, id int) (domain.UrubuKey, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return "", err
	}
	defer tx.Rollback(ctx)

	var urubukeyexitst string

	err = tx.QueryRow(context.Background(),
		"SELECT urubukey FROM clients WHERE id=$1", id).Scan(&urubukeyexitst)
	if err != nil {
		return "", err
	}

	if urubukeyexitst != "" {
		return domain.UrubuKey(urubukeyexitst), errors.New("urubukey already exists")
	}

	urubukeygenerated, err := password.Generate(10, 3, 3, false, false)
	if err != nil {
		return "", err
	}

	batch := &pgx.Batch{}
	batch.Queue("UPDATE clients SET urubukey=$2 WHERE id =$1", id, urubukeygenerated)
	sendBatch := tx.SendBatch(context.Background(), batch)
	_, err = sendBatch.Exec()
	if err != nil {
		return "", err
	}

	if err := sendBatch.Close(); err != nil {
		return "", err
	}

	if err := tx.Commit(context.Background()); err != nil {
		return "", nil
	}

	return domain.UrubuKey(urubukeygenerated), nil
}

func (r *repository) GetBankStatement(ctx context.Context, id int) (domain.BankStatemant, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return domain.BankStatemant{}, err
	}
	defer tx.Rollback(ctx)

	row, err := tx.Query(context.Background(), "SELECT balance, now(), limit FROM clients WHERE id=$1", id)
	if err != nil {
		return domain.BankStatemant{}, err
	}
	balanceAccount, err := pgx.CollectOneRow(row, pgx.RowToStructByPos[domain.BalanceStatement])
	if err != nil {
		return domain.BankStatemant{}, err
	}

	rows, err := tx.Query(context.Background(), "SELECT id, value, kind, description, payor, payee, completed_at FROM transactions WHERE id=$1", id)
	if err != nil {
		return domain.BankStatemant{}, err
	}

	lastTransactions, err := pgx.CollectRows(rows, pgx.RowToStructByPos[domain.LastTransaction])
	if err != nil {
		return domain.BankStatemant{}, err
	}

	bankstatement := domain.BankStatemant{
		Balance:          balanceAccount,
		LastTransactions: []domain.LastTransaction{},
	}

	bankstatement.LastTransactions = append(bankstatement.LastTransactions, lastTransactions...)
	err = tx.Commit(context.Background())
	if err != nil {
		return domain.BankStatemant{}, err
	}

	return bankstatement, nil
}
