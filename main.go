package main

import (
	"context"
	"flag"
	"log"

	"github.com/IvanOrsh/go-hot-res/api"
	"github.com/IvanOrsh/go-hot-res/db"
	"github.com/gofiber/fiber/v2"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const dburi = "mongodb://localhost:27017"
const dbname = "hot-res"
const userColl = "users"

var config = fiber.Config(fiber.Config{
	ErrorHandler: func(c *fiber.Ctx, err error) error {	
			return c.JSON(map[string]string{"error": err.Error()})
	},
})

func main() {
	listenAddr := flag.String("listenAddr", ":5001", "The listen address of the API server")
	flag.Parse()

	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(dburi))
	if err != nil {
		log.Fatal(err)
	}
	// handler initialization
	userHandler := api.NewUserHandler(db.NewMongoUserStore(client))


	app := fiber.New(config)
	apiv1 := app.Group("/api/v1")

	apiv1.Post("/users", userHandler.HandlePostUser)
	apiv1.Get("/users", userHandler.HandleGetUsers)
	apiv1.Get("/user/:id", userHandler.HandleGetUser)
	app.Listen(*listenAddr)
}
