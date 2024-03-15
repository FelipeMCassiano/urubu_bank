package bank

import (
	"context"

	"github.com/FelipeMCassiano/urubu_bank/internal/domain"
)

type Service interface {
	GenerateUrubukey(ctx context.Context, id int) (domain.UrubuKey, error)
	SearchClientByName(ctx context.Context, name string) ([]domain.CostumerConsult, error)
	GetBankStatement(ctx context.Context, id int) (domain.BankStatemant, error)
	VerifyIfCostumerExists(ctx context.Context, id int) (string, error)
	CreateNewAccount(ctx context.Context, client domain.CreateCostumer) (domain.CreatedCostumer, error)
	GetUsernameAndPassword(ctx context.Context, name string) (domain.User, error)
	CreateSessionToken(sessionName string) (string, error)
	DeleteSessionToken(sessionName string) error
	VerifyIfTokenExists(token string) error
	RetrieveCookies(sessionName string) (string, error)
	DeposityMoney(ctx context.Context, t domain.TransactionCredit, result chan domain.TransactionResponseCredit, errChan chan error)
	CreateTransaction(ctx context.Context, t domain.TransactionDebit, result chan domain.TransactionResponseDebit, errChan chan error)
}

type bankService struct {
	repository Respository
}

func NewService(r Respository) Service {
	return &bankService{
		repository: r,
	}
}

func (s *bankService) RetrieveCookies(sessionName string) (string, error) {
	token, err := s.repository.RetrieveCookies(sessionName)

	return token, err
}

func (s *bankService) VerifyIfTokenExists(token string) error {
	err := s.repository.VerifyIfTokenExists(token)

	return err
}

func (s *bankService) DeleteSessionToken(sessionName string) error {
	err := s.repository.DeleteSessionToken(sessionName)

	return err
}

func (s *bankService) CreateSessionToken(sessionName string) (string, error) {
	token, err := s.repository.CreateSessionToken(sessionName)

	return token, err
}

func (s *bankService) GetUsernameAndPassword(ctx context.Context, name string) (domain.User, error) {
	response, err := s.repository.GetUsernameAndPassword(ctx, name)
	return response, err
}

func (s *bankService) DeposityMoney(ctx context.Context, t domain.TransactionCredit, result chan domain.TransactionResponseCredit, errChan chan error) {
	s.repository.DeposityMoney(ctx, t, result, errChan)

	return
}

func (s *bankService) CreateTransaction(ctx context.Context, t domain.TransactionDebit, result chan domain.TransactionResponseDebit, errChan chan error) {
	s.repository.CreateTransaction(ctx, t, result, errChan)
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
