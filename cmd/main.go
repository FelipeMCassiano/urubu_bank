package main

import (
	"context"
	"database/sql"
	"log"

	"github.com/FelipeMCassiano/urubu_bank/cmd/api/routes"
	"github.com/go-redis/redis"
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

	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	if err := redisClient.Ping().Err(); err != nil {
		log.Fatal(err)
	}

	eng := fiber.New()

	router := routes.NewRouter(eng, db, redisClient)
	router.MapRoutes()

	if err := eng.Listen(":8080"); err != nil {
		panic(err)
	}
}
