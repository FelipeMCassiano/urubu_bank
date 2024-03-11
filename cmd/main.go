package main

import (
	"context"
	"database/sql"
	"log"

	"github.com/FelipeMCassiano/urubu_bank/cmd/api/routes"
	"github.com/gofiber/fiber/v2"
	_ "github.com/lib/pq"
)

func main() {
	db, err := sql.Open("postgres", "host=localhost user=urubu password=urubu dbname=urubu sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}

	err = db.PingContext(context.Background())
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
