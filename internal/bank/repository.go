package bank

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/FelipeMCassiano/urubu_bank/internal/domain"
	"github.com/go-redis/redis"
	"github.com/gofrs/uuid"
)

type Respository interface {
	GetBankStatement(ctx context.Context, id int) (domain.BankStatemant, error)
	GenerateUrubuKey(ctx context.Context, id int) (domain.UrubuKey, error)
	SearchClientByName(ctx context.Context, name string) ([]domain.CostumerConsult, error)
	CreateNewAccount(ctx context.Context, client domain.CreateCostumer) (domain.CreatedCostumer, error)
	VerifyIfClientExists(ctx context.Context, id int) (string, error)
	GetUsernameAndPassword(ctx context.Context, name string) (domain.User, error)
	CreateSessionToken(sessionName string) (string, error)
	DeleteSessionToken(sessionName string) error
	VerifyIfTokenExists(token string) error
	RetrieveCookies(sessionName string) (string, error)
	DeposityMoney(ctx context.Context, t domain.TransactionCredit, result chan domain.TransactionResponseCredit, errChan chan error)
	CreateTransaction(ctx context.Context, t domain.TransactionDebit, result chan domain.TransactionResponseDebit, errChan chan error)
	UrubuTrading(ctx context.Context, user domain.User, value int, result chan domain.ValueTraded, errChan chan error)
}

type repository struct {
	db    *sql.DB
	redis *redis.Client
}

func NewRepository(db *sql.DB, redis *redis.Client) Respository {
	return &repository{
		db:    db,
		redis: redis,
	}
}

var (
	ErrNotFound = errors.New("client not found")
	LimitErr    = errors.New("limit error")
	BalanceErr  = errors.New("value bigger than balance")
)

func (r *repository) VerifyIfTokenExists(token string) error {
	err := r.redis.Exists(token).Err()
	if err != nil {
		return err
	}

	return nil
}

func (r *repository) RetrieveCookies(sessionName string) (string, error) {
	token, err := r.redis.Get(sessionName).Result()
	if err != nil {
		return "", err
	}
	return token, err
}

func (r *repository) CreateSessionToken(sessionName string) (string, error) {
	uuiD, _ := uuid.NewV4()
	token := base64.URLEncoding.EncodeToString([]byte(uuiD.String()))
	err := r.redis.Set(sessionName, token, 24*time.Hour).Err()
	if err != nil {
		return "", err
	}

	return token, nil
}

func (r *repository) DeleteSessionToken(sessionName string) error {
	err := r.redis.Del(sessionName).Err()
	if err != nil {
		return err
	}
	return nil
}

func (r *repository) UrubuTrading(ctx context.Context, user domain.User, value int, result chan domain.ValueTraded, errChan chan error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		errChan <- err
		return
	}
	defer tx.Rollback()
	var balance int

	err = r.db.QueryRowContext(context.Background(), "SELECT balance FROM clients WHERE fullname=$1", user.Username).Scan(&balance)
	if err != nil {
		_ = tx.Rollback()
		errChan <- err
		return
	}

	if balance < value {
		_ = tx.Rollback()
		errChan <- BalanceErr
		return
	}

	key1, _ := uuid.NewV4()
	key2, _ := uuid.NewV4()

	var newbalance int

	if key1.String() != key2.String() {
		newbalance = balance - value

		stmt, err := r.db.PrepareContext(context.Background(), "UPDATE clients SET balance= $1 WHERE fullname=$2")
		if err != nil {
			_ = tx.Rollback()
			errChan <- err
			return
		}
		if _, err := stmt.ExecContext(ctx, newbalance, user.Password); err != nil {
			_ = tx.Rollback()
			errChan <- err
			return

		}

		result <- 0
		return
	}

	if key1.String() == key2.String() {
		newbalance = value * 10

		stmt, err := r.db.PrepareContext(context.Background(), "UPDATE clients SET balance= $1 WHERE fullname=$2")
		if err != nil {
			_ = tx.Rollback()
			errChan <- err
			return
		}
		if _, err := stmt.ExecContext(ctx, newbalance, user.Password); err != nil {
			_ = tx.Rollback()
			errChan <- err
			return

		}
		result <- domain.ValueTraded(newbalance)
		return

	}

	if err := tx.Commit(); err != nil {
		_ = tx.Rollback()
		errChan <- err
		return
	}
}

func (r *repository) GetUsernameAndPassword(ctx context.Context, name string) (domain.User, error) {
	var user domain.User

	// TODO: fix pointer error

	userRedis, err := r.redis.Get(name).Result()
	if err.Error() != "redis: nil" {
		return domain.User{}, err
	}

	log.Println("pass 1")

	if userRedis != "" {
		var userCached domain.User
		if err := json.Unmarshal([]byte(userRedis), &userCached); err != nil {
			return domain.User{}, err
		}
		return userCached, nil
	}

	log.Println("pass 2")

	err = r.db.QueryRowContext(ctx, "SELECT fullname, password FROM clients WHERE fullname=$1", name).Scan(&user.Username, &user.Password)
	if err != nil {
		return domain.User{}, err
	}

	log.Println("pass 3")

	userJson, err := json.Marshal(user)
	if err != nil {
		return domain.User{}, err
	}
	log.Println("pass 4")

	if err := r.redis.Set(user.Username, userJson, 72*time.Hour).Err(); err != nil {
		return domain.User{}, err
	}

	log.Println("pass 5")

	return user, nil
}

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

func (r *repository) DeposityMoney(ctx context.Context, t domain.TransactionCredit, result chan domain.TransactionResponseCredit, errChan chan error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		_ = tx.Rollback()
		errChan <- err
		return
	}
	defer tx.Rollback()
	var limit, balance int

	err = tx.QueryRowContext(context.Background(), "SELECT credit_limit, balance FROM clients WHERE id=$1 FOR UPDATE", t.Client_Id).Scan(&limit, &balance)
	if err != nil {
		_ = tx.Rollback()
		errChan <- err
		return
	}
	newbalance := balance + t.Value
	stmt1, err := tx.PrepareContext(context.Background(), "INSERT INTO transactions (client_id, value, kind, description, payee, completed_at) VALUES($1,$2,$3,$4,$5,$6)")
	if err != nil {
		_ = tx.Rollback()
		errChan <- err
		return
	}
	defer stmt1.Close()

	log.Println(newbalance)

	_, err = stmt1.ExecContext(context.Background(), t.Client_Id, t.Value, t.Kind, t.Description, "self", t.Completed_at)
	if err != nil {
		_ = tx.Rollback()
		errChan <- err

		return
	}

	stmt2, err := tx.PrepareContext(context.Background(), "UPDATE clients SET balance=$2 WHERE id=$1")
	if err != nil {
		_ = tx.Rollback()
		errChan <- err
		return
	}

	_, err = stmt2.ExecContext(context.Background(), t.Client_Id, newbalance)
	if err != nil {
		_ = tx.Rollback()
		errChan <- err

		return
	}

	response := domain.TransactionResponseCredit{
		Newbalance:   newbalance,
		Completed_at: t.Completed_at,
	}

	if err := tx.Commit(); err != nil {
		_ = tx.Rollback()
		errChan <- err
		return
	}

	result <- response
	return
}

func (r *repository) CreateTransaction(ctx context.Context, t domain.TransactionDebit, result chan domain.TransactionResponseDebit, errChan chan error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		errChan <- err
		return
	}
	defer tx.Rollback()
	var limit, balance int

	err = tx.QueryRowContext(context.Background(), "SELECT credit_limit, balance FROM clients WHERE fullname=$1", t.Payor).Scan(&limit, &balance)
	if err != nil {
		_ = tx.Rollback()
		errChan <- err

		return
	}

	log.Println(balance)

	newbalance := balance - t.Value
	log.Println(newbalance)

	if (newbalance + limit) < 0 {
		_ = tx.Rollback()
		errChan <- LimitErr

		return
	}

	var Payee string

	err = tx.QueryRowContext(context.Background(), "SELECT fullname FROM clients WHERE urubukey=$1", t.PayeeUrubuKey).Scan(&Payee)
	if err != nil {
		_ = tx.Rollback()
		errChan <- err

		return
	}

	stmt1, err := tx.PrepareContext(context.Background(), "INSERT INTO transactions (client_id, value, kind, description, payee, completed_at) VALUES($1,$2,$3,$4,$5,$6)")
	if err != nil {
		_ = tx.Rollback()
		errChan <- err

		return
	}

	_, err = stmt1.ExecContext(context.Background(), t.Client_Id, t.Value, t.Kind, t.Description, Payee, t.Completed_at)
	if err != nil {
		_ = tx.Rollback()
		errChan <- err

		return
	}

	stmt2, err := tx.PrepareContext(context.Background(), "UPDATE clients SET balance=$2 WHERE fullname=$1")
	if err != nil {
		_ = tx.Rollback()
		errChan <- err

		return
	}

	_, err = stmt2.ExecContext(context.Background(), t.Payor, newbalance)
	if err != nil {
		_ = tx.Rollback()
		errChan <- err

		return
	}

	stmt3, err := tx.PrepareContext(context.Background(), "UPDATE clients SET balance = balance + $2 WHERE urubukey=$1")
	if err != nil {
		_ = tx.Rollback()
		errChan <- err

		return
	}

	_, err = stmt3.ExecContext(context.Background(), t.PayeeUrubuKey, t.Value)
	if err != nil {
		_ = tx.Rollback()
		errChan <- err

		return
	}

	defer stmt1.Close()
	defer stmt2.Close()
	defer stmt3.Close()

	err = tx.Commit()
	if err != nil {
		if err.Error() == "no rows in result set" {
			_ = tx.Rollback()
			errChan <- ErrNotFound

			return
		}
		errChan <- err

		return
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

	result <- response

	return
}

func (r *repository) CreateNewAccount(ctx context.Context, client domain.CreateCostumer) (domain.CreatedCostumer, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		_ = tx.Rollback()
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
		_ = tx.Rollback()
		return domain.CreatedCostumer{}, err
	}

	if err := tx.Commit(); err != nil {
		_ = tx.Rollback()
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

	urubukeygeneratedU, err := uuid.NewV4()
	if err != nil {
		return "", err
	}

	urubukeygenerated := urubukeygeneratedU.String()

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
