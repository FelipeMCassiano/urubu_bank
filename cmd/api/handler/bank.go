package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/FelipeMCassiano/urubu_bank/internal/bank"
	"github.com/FelipeMCassiano/urubu_bank/internal/domain"
	"github.com/labstack/echo/v4"
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

func (b *BankController) CreateTransaction() echo.HandlerFunc {
	return func(ctx echo.Context) error {
		input := &TransactionRequest{}
		stdctx := ctx.Request().Context()

		if err := ctx.Bind(input); err != nil {
			return ctx.JSON(http.StatusUnprocessableEntity, ErrInvalidJson.Error())
		}

		id, _ := strconv.Atoi(ctx.Param("id"))
		payor, err := b.bankService.VerifyIfClientExists(stdctx, id)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, ErrNotFound.Error())
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
				return ctx.JSON(http.StatusNotFound, ErrNotFound.Error())
			}
			return ctx.JSON(http.StatusInternalServerError, err.Error())
		}

		return ctx.JSON(http.StatusCreated, response)
	}
}
