package bank

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"time"

	"github.com/FelipeMCassiano/urubu_bank/internal/domain"
	"github.com/sethvargo/go-password/password"
)

type Respository interface {
	CreateTransaction(ctx context.Context, t domain.TransactionDebit) (domain.TransactionResponseDebit, error)
	GetBankStatement(ctx context.Context, id int) (domain.BankStatemant, error)
	GenerateUrubuKey(ctx context.Context, id int) (domain.UrubuKey, error)
	SearchClientByName(ctx context.Context, name string) ([]domain.CostumerConsult, error)
	CreateNewAccount(ctx context.Context, client domain.CreateCostumer) (domain.CreatedCostumer, error)
	VerifyIfClientExists(ctx context.Context, id int) (string, error)
	DeposityMoney(ctx context.Context, t domain.TransactionCredit) (domain.TransactionResponseCredit, error)
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Respository {
	return &repository{
		db: db,
	}
}

var (
	ErrNotFound = errors.New("client not found")
	LimitErr    = errors.New("limit error")
)

func (r *repository) SearchClientByName(ctx context.Context, name string) ([]domain.CostumerConsult, error) {
	result, err := r.db.QueryContext(ctx, "SElECT fullname, urubukey FROM clients WHERE fullname ILIKE '%' || $1 || '%'", name)
	if err != nil {
		return []domain.CostumerConsult{}, err
	}
	var SliceOfClients []domain.CostumerConsult

	if result != nil {
		for result.Next() {
			var cname, urubukey string
			err := result.Scan(&cname, &urubukey)
			if err != nil {
				return []domain.CostumerConsult{}, err
			}
			client := domain.CostumerConsult{
				Fullname: cname,
				UrubuKey: domain.UrubuKey(urubukey),
			}

			SliceOfClients = append(SliceOfClients, client)
		}
	}

	return SliceOfClients, nil
}

func (r *repository) VerifyIfClientExists(ctx context.Context, id int) (string, error) {
	var clientxists string
	err := r.db.QueryRowContext(ctx, "SELECT fullname FROM clients WHERE id=$1", id).Scan(&clientxists)
	if err != nil {
		return "", err
	}

	return clientxists, nil
}

func (r *repository) DeposityMoney(ctx context.Context, t domain.TransactionCredit) (domain.TransactionResponseCredit, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.TransactionResponseCredit{}, err
	}
	defer tx.Rollback()
	var limit, balance int

	err = tx.QueryRowContext(context.Background(), "SELECT credit_limit, balance FROM clients WHERE id=$1", t.Client_Id).Scan(&limit, &balance)
	if err != nil {
		return domain.TransactionResponseCredit{}, err
	}
	newbalance := balance + t.Value
	stmt1, err := tx.PrepareContext(context.Background(), "INSERT INTO transactions (client_id, value, kind, description, payee, completed_at) VALUES($1,$2,$3,$4,$5,$6)")
	if err != nil {
		return domain.TransactionResponseCredit{}, err
	}
	defer stmt1.Close()

	log.Println(newbalance)

	_, err = stmt1.ExecContext(context.Background(), t.Client_Id, t.Value, t.Kind, t.Description, "self", t.Completed_at)
	if err != nil {
		return domain.TransactionResponseCredit{}, err
	}

	stmt2, err := tx.PrepareContext(context.Background(), "UPDATE clients SET balance=$2 WHERE id=$1")
	if err != nil {
		return domain.TransactionResponseCredit{}, err
	}

	_, err = stmt2.ExecContext(context.Background(), t.Client_Id, newbalance)
	if err != nil {
		return domain.TransactionResponseCredit{}, err
	}

	response := domain.TransactionResponseCredit{
		Newbalance:   newbalance,
		Completed_at: t.Completed_at,
	}

	if err := tx.Commit(); err != nil {
		return domain.TransactionResponseCredit{}, err
	}
	return response, nil
}

func (r *repository) CreateTransaction(ctx context.Context, t domain.TransactionDebit) (domain.TransactionResponseDebit, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.TransactionResponseDebit{}, err
	}
	defer tx.Rollback()
	var limit, balance int

	err = tx.QueryRowContext(context.Background(), "SELECT credit_limit, balance FROM clients WHERE fullname=$1", t.Payor).Scan(&limit, &balance)
	if err != nil {
		return domain.TransactionResponseDebit{}, err
	}

	log.Println(balance)

	newbalance := balance - t.Value
	log.Println(newbalance)

	if (newbalance + limit) < 0 {
		return domain.TransactionResponseDebit{}, LimitErr
	}

	var Payee string

	err = tx.QueryRowContext(context.Background(), "SELECT fullname FROM clients WHERE urubukey=$1", t.PayeeUrubuKey).Scan(&Payee)
	if err != nil {
		return domain.TransactionResponseDebit{}, ErrNotFound
	}

	stmt1, err := tx.PrepareContext(context.Background(), "INSERT INTO transactions (client_id, value, kind, description, payee, completed_at) VALUES($1,$2,$3,$4,$5,$6)")
	if err != nil {
		return domain.TransactionResponseDebit{}, err
	}

	_, err = stmt1.ExecContext(context.Background(), t.Client_Id, t.Value, t.Kind, t.Description, Payee, t.Completed_at)
	if err != nil {
		return domain.TransactionResponseDebit{}, err
	}

	stmt2, err := tx.PrepareContext(context.Background(), "UPDATE clients SET balance=$2 WHERE fullname=$1")
	if err != nil {
		return domain.TransactionResponseDebit{}, err
	}

	_, err = stmt2.ExecContext(context.Background(), t.Payor, newbalance)
	if err != nil {
		return domain.TransactionResponseDebit{}, err
	}

	stmt3, err := tx.PrepareContext(context.Background(), "UPDATE clients SET balance = balance + $2 WHERE urubukey=$1")
	if err != nil {
		return domain.TransactionResponseDebit{}, err
	}

	_, err = stmt3.ExecContext(context.Background(), t.PayeeUrubuKey, t.Value)
	if err != nil {
		return domain.TransactionResponseDebit{}, err
	}

	defer stmt1.Close()
	defer stmt2.Close()
	defer stmt3.Close()

	err = tx.Commit()
	if err != nil {
		if err.Error() == "no rows in result set" {
			return domain.TransactionResponseDebit{}, ErrNotFound
		}
		return domain.TransactionResponseDebit{}, err
	}

	response := domain.TransactionResponseDebit{
		Description:  t.Description,
		Value:        t.Value,
		Kind:         t.Kind,
		Payor:        t.Payor,
		Payee:        Payee,
		Completed_at: t.Completed_at,
		Balance:      newbalance,
	}

	return response, nil
}

func (r *repository) CreateNewAccount(ctx context.Context, client domain.CreateCostumer) (domain.CreatedCostumer, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.CreatedCostumer{}, err
	}

	createdClient := domain.CreatedCostumer{
		Fullname: client.Fullname,
		Limit:    client.Limit,
	}

	log.Println("Inserting client:", client)
	var id int

	err = tx.QueryRowContext(context.Background(), "INSERT INTO clients (fullname, birth, credit_limit, password) VALUES ($1, $2, $3, $4) RETURNING id",
		client.Fullname, client.Birth, client.Limit, client.Password).Scan(&id)
	if err != nil {
		return domain.CreatedCostumer{}, err
	}

	if err := tx.Commit(); err != nil {
		return domain.CreatedCostumer{}, err
	}

	createdClient.ID = id

	return createdClient, nil
}

func (r *repository) GenerateUrubuKey(ctx context.Context, id int) (domain.UrubuKey, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()
	log.Println("comecei GenerateUrubuKey")

	urubukeygenerated, err := password.Generate(20, 10, 5, false, false)
	if err != nil {
		return "", err
	}

	stmt, err := tx.PrepareContext(context.Background(), "UPDATE clients SET urubukey=$2 WHERE id =$1")
	if err != nil {
		return "", err
	}

	defer stmt.Close()

	_, err = stmt.Exec(id, urubukeygenerated)
	if err != nil {
		return "", err
	}

	if err := tx.Commit(); err != nil {
		return "", nil
	}

	return domain.UrubuKey(urubukeygenerated), nil
}

func (r *repository) GetBankStatement(ctx context.Context, id int) (domain.BankStatemant, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.BankStatemant{}, err
	}
	defer tx.Rollback()

	var balance, credit_limit int

	err = tx.QueryRowContext(context.Background(), "SELECT balance ,credit_limit FROM clients WHERE id=$1", id).Scan(&balance, &credit_limit)
	if err != nil {
		return domain.BankStatemant{}, err
	}

	balanceAccount := domain.BalanceStatement{
		Balance:      balance,
		Limit:        credit_limit,
		Completed_at: time.Now(),
	}

	rows, err := tx.QueryContext(context.Background(), "SELECT id, value, kind, description, payee, completed_at FROM transactions WHERE client_id=$1", id)
	if err != nil {
		return domain.BankStatemant{}, err
	}
	bankstatement := domain.BankStatemant{
		Balance:          balanceAccount,
		LastTransactions: []domain.LastTransaction{},
	}

	if rows != nil {
		for rows.Next() {
			var Transaction domain.LastTransaction
			err := rows.Scan(&Transaction.ID, &Transaction.Value, &Transaction.Kind, &Transaction.Description, &Transaction.Payee, &Transaction.Completed_at)
			if err != nil {
				return domain.BankStatemant{}, err
			}
			log.Println(Transaction)

			bankstatement.LastTransactions = append(bankstatement.LastTransactions, Transaction)

		}
	}

	err = tx.Commit()
	if err != nil {
		return domain.BankStatemant{}, err
	}

	return bankstatement, nil
}
