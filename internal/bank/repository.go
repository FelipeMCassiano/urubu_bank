package bank

import (
	"context"

	"github.com/FelipeMCassiano/urubu_bank/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Respository interface {
	MakeTransaction(ctx context.Context, t domain.Transaction) (domain.TransactionResponse, error)
	GenerateUrubuKey() (int, error)
	GetBankStatement(ctx context.Context, id int) (domain.BankStatemant, error)
}

type repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) Respository {
	return &repository{
		db: db,
	}
}

func (r *repository) MakeTransaction(ctx context.Context, t domain.Transaction) (domain.TransactionResponse, error) {
	return domain.TransactionResponse{}, nil
}

func (r *repository) GenerateUrubuKey() (int, error) {
	return 0, nil
}

func (r *repository) GetBankStatement(ctx context.Context, id int) (domain.BankStatemant, error) {
	return domain.BankStatemant{}, nil
}
