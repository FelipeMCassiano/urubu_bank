package routes

import (
	"database/sql"

	"github.com/FelipeMCassiano/urubu_bank/cmd/api/handler"
	"github.com/FelipeMCassiano/urubu_bank/internal/bank"
	"github.com/gofiber/fiber/v2"
)

type Router interface {
	MapRoutes()
}

type router struct {
	eng *fiber.App
	rg  fiber.Router
	db  *sql.DB
}

func NewRouter(eng *fiber.App, db *sql.DB) Router {
	return &router{eng: eng, db: db}
}

func (r *router) MapRoutes() {
	r.setGroup()
	r.buildRoutes()
}

func (r *router) setGroup() {
	r.rg = r.eng.Group("")
}

func (r *router) buildRoutes() {
	repo := bank.NewRepository(r.db)
	service := bank.NewService(repo)
	handler := handler.NewBank(service)

	r.rg.Post("/costumers/:id/transacoes", handler.CreateTransaction())
	r.rg.Post("/costumers/create", handler.CreateNewAccount())
	r.rg.Post("/costumers/:id/depositymoney", handler.DeposityMoney())
	r.rg.Get("/costumers/:id/bankstatement", handler.GetBankStatement())
	r.rg.Get("/costumers", handler.SeachCostumerByName())
}
