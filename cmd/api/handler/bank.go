package handler

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/FelipeMCassiano/urubu_bank/internal/bank"
	"github.com/FelipeMCassiano/urubu_bank/internal/domain"
	"github.com/go-redis/redis"
	"github.com/gofiber/fiber/v2"
)

var (
	ErrInvalidJson  = errors.New("invalid json")
	ErrNotFound     = errors.New("client not found")
	InvalidToDToErr = errors.New("invalid request")
	LimitErr        = errors.New("limit error")
)

type TransactionRequestDebit struct {
	Value         int    `json:"value" validate:"required,gt=0"`
	Kind          string `json:"kind" validate:"required,oneof=debit"`
	Description   string `json:"description" validate:"required, min=1, max=10"`
	PayeeUrubuKey string `json:"payeeurubukey" validate:"required"`
}
type TransactionRequestCredit struct {
	Value       int    `json:"value" validate:"required,gt=0"`
	Kind        string `json:"kind" validate:"required,oneof=credit"`
	Description string `json:"description" validate:"required, min=1, max=10"`
}

type BankController struct {
	bankService bank.Service
}

func NewBank(s bank.Service) *BankController {
	return &BankController{
		bankService: s,
	}
}

const sessionName = "session-name"

func (b *BankController) IsAuthenticated() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		token, err := b.bankService.RetrieveCookies(sessionName)
		if err != nil {
			if err == redis.Nil {
				return ctx.Status(fiber.StatusUnauthorized).SendString("Unauthorized")
			}
			return ctx.Status(fiber.StatusInternalServerError).JSON(err.Error())
		}

		fmt.Printf("token: %v\n", token)
		if token == "" {
			return ctx.Status(fiber.StatusUnauthorized).SendString("Unauthorized")
		}

		if err := b.bankService.VerifyIfTokenExists(token); err != nil {
			return ctx.Status(fiber.StatusUnauthorized).SendString("Invalid Session token")
		}

		return ctx.Next()
	}
}

func (b *BankController) Login() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		username := ctx.FormValue("username")
		password := ctx.FormValue("password")

		user, err := b.bankService.GetUsernameAndPassword(ctx.Context(), username)
		if err != nil {
			return err
		}

		if user.Password != password {
			return ctx.Status(fiber.StatusUnauthorized).SendString("Invalid password or usesrname")
		}

		token, err := b.bankService.CreateSessionToken(sessionName)
		if err != nil {
			return err
		}
		log.Println(sessionName)

		log.Println(token)

		ctx.Cookie(&fiber.Cookie{
			Name:     sessionName,
			Value:    token,
			Expires:  time.Now().Add(24 * time.Hour),
			HTTPOnly: true,
		})

		return ctx.SendString("Login successful")
	}
}

func (b *BankController) Logout() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		err := b.bankService.DeleteSessionToken(sessionName)
		if err != nil {
			return ctx.Status(fiber.StatusInternalServerError).JSON(err.Error())
		}

		ctx.ClearCookie(sessionName)
		return ctx.SendString("Logout successful")
	}
}

func (b *BankController) DeposityMoney() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		input := &TransactionRequestCredit{}
		stdctx := ctx.Context()

		if err := ctx.BodyParser(input); err != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(ErrInvalidJson.Error())
		}

		id, _ := ctx.ParamsInt("id")
		_, err := b.bankService.VerifyIfCostumerExists(stdctx, id)
		if err != nil {
			return ctx.Status(fiber.StatusUnprocessableEntity).JSON(err.Error())
		}

		newtransaction := domain.TransactionCredit{
			Client_Id:    id,
			Value:        input.Value,
			Kind:         input.Kind,
			Description:  input.Description,
			Completed_at: time.Now(),
		}

		response, err := b.bankService.DeposityMoney(stdctx, newtransaction)
		if err != nil {
			return err
		}

		return ctx.Status(fiber.StatusOK).JSON(response)
	}
}

func (b *BankController) CreateTransaction() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		input := &TransactionRequestDebit{}
		stdctx := ctx.Context()

		if err := ctx.BodyParser(input); err != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(ErrInvalidJson.Error())
		}

		log.Println(input.Description)

		id, _ := ctx.ParamsInt("id")
		payor, err := b.bankService.VerifyIfCostumerExists(stdctx, id)
		if err != nil {
			return ctx.Status(fiber.StatusNotFound).JSON(ErrNotFound.Error())
		}

		newtransaction := domain.TransactionDebit{
			Value:         input.Value,
			Client_Id:     id,
			Kind:          input.Kind,
			Description:   input.Description,
			Payor:         payor,
			PayeeUrubuKey: input.PayeeUrubuKey,
			Completed_at:  time.Now(),
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

func (b *BankController) SearchCostumerByName() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		name := ctx.Query("name")
		stdctx := ctx.Context()

		log.Printf("catch a query with name %s", name)

		respose, err := b.bankService.SearchClientByName(stdctx, name)
		if err != nil {
			return ctx.Status(fiber.StatusNotFound).JSON(ErrNotFound.Error())

			// return ctx.Status(fiber.StatusNotFound).JSON(err.Error())
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
			log.Println(err.Error())
			return ctx.Status(fiber.StatusNoContent).JSON(err.Error())
		}

		return ctx.Status(fiber.StatusOK).JSON(bankstatement)
	}
}
