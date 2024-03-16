package routes

import (
	"database/sql"

	"github.com/FelipeMCassiano/urubu_bank/cmd/api/handler"
	"github.com/FelipeMCassiano/urubu_bank/internal/bank"
	"github.com/go-redis/redis"
	"github.com/gofiber/fiber/v2"
)

type Router interface {
	MapRoutes()
}

type router struct {
	eng   *fiber.App
	rg    fiber.Router
	db    *sql.DB
	redis *redis.Client
}

func NewRouter(eng *fiber.App, db *sql.DB, redis *redis.Client) Router {
	return &router{eng: eng, db: db, redis: redis}
}

func (r *router) MapRoutes() {
	r.setGroup()
	r.buildRoutes()
}

func (r *router) setGroup() {
	r.rg = r.eng.Group("")
}

func (r *router) buildRoutes() {
	repo := bank.NewRepository(r.db, r.redis)
	service := bank.NewService(repo)
	handler := handler.NewBank(service)

	r.rg.Post("/costumers/:id/transacoes", handler.IsAuthenticated(), handler.CreateTransaction())
	r.rg.Post("/costumers/create", handler.CreateNewAccount())
	r.rg.Post("/costumers/:id/depositymoney", handler.IsAuthenticated(), handler.DeposityMoney())
	r.rg.Get("/costumers/:id/bankstatement", handler.IsAuthenticated(), handler.GetBankStatement())
	r.rg.Get("/costumers/search", handler.IsAuthenticated(), handler.SearchCostumerByName())
	r.rg.Post("/costumers/login", handler.Login())
	r.rg.Get("/costumers/logout", handler.IsAuthenticated(), handler.Logout())
	r.rg.Post("/costumers/urubutrading", handler.UrubuTrading())
}
