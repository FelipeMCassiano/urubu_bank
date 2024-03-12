package bank

import (
	"context"

	"github.com/FelipeMCassiano/urubu_bank/internal/domain"
)

type Service interface {
	GenerateUrubukey(ctx context.Context, id int) (domain.UrubuKey, error)
	CreateTransaction(ctx context.Context, t domain.TransactionDebit) (domain.TransactionResponseDebit, error)
	SearchClientByName(ctx context.Context, name string) ([]domain.CostumerConsult, error)
	GetBankStatement(ctx context.Context, id int) (domain.BankStatemant, error)
	VerifyIfCostumerExists(ctx context.Context, id int) (string, error)
	CreateNewAccount(ctx context.Context, client domain.CreateCostumer) (domain.CreatedCostumer, error)
	DeposityMoney(ctx context.Context, t domain.TransactionCredit) (domain.TransactionResponseCredit, error)
}

type bankService struct {
	repository Respository
}

func NewService(r Respository) Service {
	return &bankService{
		repository: r,
	}
}

func (s *bankService) DeposityMoney(ctx context.Context, t domain.TransactionCredit) (domain.TransactionResponseCredit, error) {
	response, err := s.repository.DeposityMoney(ctx, t)

	return response, err
}

func (s *bankService) CreateTransaction(ctx context.Context, t domain.TransactionDebit) (domain.TransactionResponseDebit, error) {
	response, err := s.repository.CreateTransaction(ctx, t)
	return response, err
}

func (s *bankService) SearchClientByName(ctx context.Context, name string) ([]domain.CostumerConsult, error) {
	clients, err := s.repository.SearchClientByName(ctx, name)

	return clients, err
}

func (s *bankService) GetBankStatement(ctx context.Context, id int) (domain.BankStatemant, error) {
	bankStatemant, err := s.repository.GetBankStatement(ctx, id)

	return bankStatemant, err
}

func (s *bankService) GenerateUrubukey(ctx context.Context, id int) (domain.UrubuKey, error) {
	urubukey, err := s.repository.GenerateUrubuKey(ctx, id)

	return urubukey, err
}

func (s *bankService) CreateNewAccount(ctx context.Context, client domain.CreateCostumer) (domain.CreatedCostumer, error) {
	response, err := s.repository.CreateNewAccount(ctx, client)

	return response, err
}

func (s *bankService) VerifyIfCostumerExists(ctx context.Context, id int) (string, error) {
	clientname, err := s.repository.VerifyIfClientExists(ctx, id)

	return clientname, err
}
