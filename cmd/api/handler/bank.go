package handler

import (
	"errors"

	"github.com/FelipeMCassiano/urubu_bank/internal/bank"
	"github.com/FelipeMCassiano/urubu_bank/internal/domain"
	"github.com/gofiber/fiber/v2"
)

var (
	ErrInvalidJson  = errors.New("invalid json")
	ErrNotFound     = errors.New("client not found")
	InvalidToDToErr = errors.New("invalid request")
	LimitErr        = errors.New("limit error")
)

type TransactionRequest struct {
	Value         int    `json:"value" validate:"required,gt=0"`
	Kind          string `json:"kind" validate:"required,oneof=credit debit"`
	Description   string `json:"description" validate:"required, min=1, max=10"`
	Payor         string `json:"payor"  validate:"required"`
	PayeeUrubuKey string `json:"payeeurubukey" validate:"required"`
}

type BankController struct {
	bankService bank.Service
}

func NewBank(s bank.Service) *BankController {
	return &BankController{
		bankService: s,
	}
}

func (b *BankController) CreateTransaction() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		input := &TransactionRequest{}
		stdctx := ctx.Context()

		if err := ctx.BodyParser(input); err != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(ErrInvalidJson.Error())
		}

		id, _ := ctx.ParamsInt("id")
		payor, err := b.bankService.VerifyIfCostumerExists(stdctx, id)
		if err != nil {
			return ctx.Status(fiber.StatusNotFound).JSON(ErrNotFound.Error())
		}

		newtransaction := domain.Transaction{
			Client_Id:     id,
			Kind:          input.Kind,
			Description:   input.Description,
			Payor:         payor,
			PayeeUrubuKey: input.PayeeUrubuKey,
		}

		response, err := b.bankService.CreateTransaction(stdctx, newtransaction)
		if err != nil {
			if err.Error() == ErrNotFound.Error() {
				return ctx.Status(fiber.StatusNotFound).JSON(ErrNotFound.Error())
			}
			return ctx.Status(fiber.StatusInternalServerError).JSON(err.Error())
		}

		return ctx.Status(fiber.StatusCreated).JSON(response)
	}
}

func (b *BankController) SeachCostumerByName() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		name := ctx.Query("name")
		stdctx := ctx.Context()

		respose, err := b.bankService.SearchClientByName(stdctx, name)
		if err != nil {
			return ctx.Status(fiber.StatusNotFound).JSON(ErrNotFound.Error())
		}

		return ctx.Status(fiber.StatusOK).JSON(respose)
	}
}

func (b *BankController) CreateNewAccount() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		newcostumer := domain.CreateCostumer{}
		stdctx := ctx.Context()

		if err := ctx.BodyParser(&newcostumer); err != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(err.Error())
		}

		createdCostumer, err := b.bankService.CreateNewAccount(stdctx, newcostumer)
		if err != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(err.Error())
		}
		urubukey, err := b.bankService.GenerateUrubukey(stdctx, createdCostumer.ID)
		if err != nil {
			return ctx.Status(fiber.StatusUnprocessableEntity).JSON(err.Error())
		}
		createdCostumer.UrubuKey = urubukey

		return ctx.Status(fiber.StatusOK).JSON(createdCostumer)
	}
}

func (b *BankController) GetBankStatement() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		id, err := ctx.ParamsInt("id")
		if err != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(err.Error())
		}

		stdctx := ctx.Context()

		bankstatement, err := b.bankService.GetBankStatement(stdctx, id)
		if err != nil {
			return ctx.Status(fiber.StatusNoContent).JSON(err.Error())
		}

		return ctx.Status(fiber.StatusOK).JSON(bankstatement)
	}
}
