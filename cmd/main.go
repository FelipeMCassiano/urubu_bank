package main

import (
	"context"
	"database/sql"
	"log"
	"os"

	"github.com/FelipeMCassiano/urubu_bank/cmd/api/routes"
	"github.com/go-redis/redis"
	"github.com/gofiber/fiber/v2"
	_ "github.com/lib/pq"
)

func main() {
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
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
		Addr:     "cache:6379",
		Password: "",
		DB:       0,
	})
	if err := redisClient.Ping().Err(); err != nil {
		log.Fatal(err)
	}

	eng := fiber.New()

	router := routes.NewRouter(eng, db, redisClient)
	router.MapRoutes()

	if err := eng.Listen(":" + os.Getenv("APP_PORT")); err != nil {
		panic(err)
	}
}
