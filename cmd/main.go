package main

import (
	"context"
	"log"
	"os"

	"github.com/FelipeMCassiano/urubu_bank/cmd/api/routes"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	var psqlconn string = os.Getenv("DATABASE_URL")

	poolConfig, err := pgxpool.ParseConfig(psqlconn)
	if err != nil {
		log.Fatal(err)
	}

	db, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		log.Fatal(err)
	}

	err = db.Ping(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	log.Println("Connected!")

	eng := fiber.New()

	router := routes.NewRouter(eng, db)
	router.MapRoutes()

	if err := eng.Listen(":8080"); err != nil {
		panic(err)
	}
}
