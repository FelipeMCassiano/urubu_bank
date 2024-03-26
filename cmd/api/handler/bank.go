package handler

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/FelipeMCassiano/urubu_bank/internal/bank"
	"github.com/FelipeMCassiano/urubu_bank/internal/domain"
	"github.com/go-playground/validator/v10"
	"github.com/go-redis/redis"
	"github.com/gofiber/fiber/v2"
)

var (
	ErrInvalidJson  = errors.New("invalid json")
	ErrNotFound     = errors.New("client not found")
	InvalidToDToErr = errors.New("invalid request")
	LimitErr        = errors.New("limit error")
	BalanceErr      = errors.New("value bigger than balance")
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
type UserLoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}
type UrubuTradingRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
	Value    int    `json:"value" validate:"required,gt=0"`
}

type BankController struct {
	bankService bank.Service
}

type ErrorResponse struct {
	Error       bool
	FailedField string
	Tag         string
	Value       interface{}
}

func NewBank(s bank.Service) *BankController {
	return &BankController{
		bankService: s,
	}
}

var validate *validator.Validate

func init() {
	validate = validator.New()
}

func validateStruct(data interface{}) error {
	err := validate.Struct(data)
	if err != nil {
		return err
	}
	return nil
}

const sessionName = "session-name"

func (b *BankController) UrubuTrading() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		request := UrubuTradingRequest{}

		if err := ctx.BodyParser(request); err != nil {
			return ctx.Status(fiber.StatusUnprocessableEntity).JSON(err.Error())
		}

		if err := validateStruct(request); err != nil {
			return ctx.Status(fiber.StatusUnprocessableEntity).JSON(err.Error())
		}

		user, err := b.bankService.GetUsernameAndPassword(ctx.Context(), request.Username)
		if err != nil {
			return ctx.Status(fiber.StatusNotFound).SendString(err.Error())
		}

		if request.Password != user.Password {
			return ctx.Status(fiber.StatusUnauthorized).SendString("Password not correct")
		}

		result := make(chan domain.ValueTraded, 1)
		errChan := make(chan error, 1)

		go b.bankService.UrubuTrading(ctx.Context(), user, request.Value, result, errChan)

		select {
		case response := <-result:
			if response == 0 {
				return ctx.SendString("Thank you for money dumbass")
			}

			return ctx.JSON(fiber.Map{
				"Congratulations,you've earned": response,
			})

		case err := <-errChan:
			if err == BalanceErr {
				return ctx.Status(fiber.StatusUnprocessableEntity).JSON(err.Error())
			}
			return ctx.Status(fiber.StatusInternalServerError).JSON(err.Error())
		}
	}
}

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
		userLogin := UserLoginRequest{}

		if err := ctx.BodyParser(userLogin); err != nil {
			return ctx.Status(fiber.StatusUnprocessableEntity).JSON(err.Error())
		}

		if err := validateStruct(userLogin); err != nil {
			return ctx.Status(fiber.StatusUnprocessableEntity).JSON(err.Error())
		}
		user, err := b.bankService.GetUsernameAndPassword(ctx.Context(), userLogin.Username)
		if err != nil {
			return err
		}

		if user.Password != userLogin.Password {
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

		if err := validateStruct(input); err != nil {
			return ctx.Status(fiber.StatusUnprocessableEntity).JSON(err.Error())
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

		result := make(chan domain.TransactionResponseCredit, 1)
		errChan := make(chan error, 1)

		go b.bankService.DeposityMoney(stdctx, newtransaction, result, errChan)

		select {
		case response := <-result:

			return ctx.Status(fiber.StatusOK).JSON(response)
		case err := <-errChan:

			return ctx.Status(fiber.StatusUnprocessableEntity).JSON(err.Error())
		}
	}
}

func (b *BankController) CreateTransaction() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		input := &TransactionRequestDebit{}
		stdctx := ctx.Context()

		if err := ctx.BodyParser(input); err != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(ErrInvalidJson.Error())
		}

		if err := validateStruct(input); err != nil {
			return ctx.Status(fiber.StatusUnprocessableEntity).JSON(err.Error())
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
		result := make(chan domain.TransactionResponseDebit, 1)
		errChan := make(chan error, 1)

		go b.bankService.CreateTransaction(stdctx, newtransaction, result, errChan)

		select {
		case response := <-result:
			return ctx.Status(fiber.StatusCreated).JSON(response)
		case err := <-errChan:
			if err.Error() == ErrNotFound.Error() {
				return ctx.Status(fiber.StatusNotFound).JSON(ErrNotFound.Error())
			}
			if err.Error() == LimitErr.Error() {
				return ctx.Status(fiber.StatusUnprocessableEntity).JSON(err.Error())
			}
			return ctx.Status(fiber.StatusInternalServerError).JSON(err.Error())

		}
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

		if err := validateStruct(newcostumer); err != nil {
			return ctx.Status(fiber.StatusUnprocessableEntity).JSON(err.Error())
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
