package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http/httptest"
	"testing"

	"github.com/IvanOrsh/go-hot-res/db"
	"github.com/IvanOrsh/go-hot-res/types"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	testdburi = "mongodb://localhost:27017"
	dbname = "hot-res-test"
)

type testdb struct {
	db.UserStore
}

func (tdb testdb) teardown(t *testing.T) {
	if err := tdb.UserStore.Drop(context.TODO()); err != nil {
		t.Fatal(err)
	}
}

func (tdb testdb) seedUsers(t *testing.T) *types.User {
	user := types.User{
		FirstName: "James",
		LastName: "St. James",
		Email: "valid_email1@email.com",
		EncryptedPassword: "encrypted",
	}

	resp, err := tdb.UserStore.InsertUser(context.TODO(), &user)
	if err != nil {
		t.Fatal(err)
	}

	return resp
}

func setup(t *testing.T) *testdb {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(testdburi))
	if err != nil {
		log.Fatal(err)
	}

	return &testdb{
		UserStore: db.NewMongoUserStore(client, dbname),
	}
}

func TestPostUser(t *testing.T) {
	tdb := setup(t)
	defer tdb.teardown(t)

	app := fiber.New()
	userHandler := NewUserHandler(tdb.UserStore)
	app.Post("/", userHandler.HandlePostUser)

	params := types.CreateUserParams{
		Email: "valid_email@email.com",
		FirstName: "James",
		LastName: "Foo",
		Password: "valid_password123",
	}

	b, _ := json.Marshal(params)
	req := httptest.NewRequest("POST", "/", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Error(err)
	}

	var user types.User

	json.NewDecoder(resp.Body).Decode(&user)

	if resp.StatusCode != 200 {
		t.Errorf("expected status code to be 200")
	}

	if len(user.ID) == 0 {
		t.Errorf("expected a user id to be set")
	}

	if len(user.EncryptedPassword) > 0 {
		t.Errorf("expected the EncryptedPassword not to be included in the json response")
	}

	if user.FirstName != params.FirstName {
		t.Errorf("expected firstName %s but got %s", params.FirstName, user.FirstName)
	}

	if user.LastName != params.LastName {
		t.Errorf("expected lastName %s but got %s", params.LastName, user.LastName)
	}
	
	if user.Email != params.Email {
		t.Errorf("expected email %s but got %s", params.Email, user.Email)
	}
}

func TestGetUser(t *testing.T) {
	tdb := setup(t)
	insertedUser := tdb.seedUsers(t)
	defer tdb.teardown(t)

	app := fiber.New()
	userHandler := NewUserHandler(tdb.UserStore)
	app.Get("/:id", userHandler.HandleGetUser)

	stringObjectID := primitive.ObjectID.Hex(insertedUser.ID)

	req := httptest.NewRequest(
		"GET",
		fmt.Sprintf("/%s", stringObjectID),
		nil,
	)
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Error(err)
	}
	
	var user types.User

	json.NewDecoder(resp.Body).Decode(&user)

	if resp.StatusCode != 200 {
		t.Errorf("expected status code to be 200")
	}	

	if user.FirstName != insertedUser.FirstName {
		t.Errorf("expected firstName %s but got %s", user.FirstName, insertedUser.FirstName)
	}

	if user.LastName != insertedUser.LastName {
		t.Errorf("expected lastName %s but got %s", user.LastName, insertedUser.LastName)
	}
	
	if user.Email != insertedUser.Email {
		t.Errorf("expected email %s but got %s", user.Email, insertedUser.Email)
	}
}

func TestPutUser(t *testing.T) {
	tdb := setup(t)
	insertedUser := tdb.seedUsers(t)
	defer tdb.teardown(t)

	app := fiber.New()
	userHandler := NewUserHandler(tdb.UserStore)
	app.Put("/:id", userHandler.HandlePutUser)
	app.Get("/:id", userHandler.HandleGetUser)

	stringObjectID := primitive.ObjectID.Hex(insertedUser.ID)

	params := types.UpdateUserParams{
		FirstName: "NewName",
		LastName: "NewLastName",
	}

	b, _ := json.Marshal(params)
	req := httptest.NewRequest(
		"PUT",
		fmt.Sprintf("/%s", stringObjectID),
		bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	_, err := app.Test(req)
	if err != nil {
		t.Error(err)
	}

	req = httptest.NewRequest(
		"GET",
		fmt.Sprintf("/%s", stringObjectID),
		nil,
	)
	if err != nil {
		t.Error(err)
	}
	resp, err := app.Test(req)
	if err != nil {
		t.Error(err)
	}

	var user types.User

	json.NewDecoder(resp.Body).Decode(&user)

	if resp.StatusCode != 200 {
		t.Errorf("expected status code to be 200")
	}

	if len(user.ID) == 0 {
		t.Errorf("expected a user id to be set")
	}


	if len(user.EncryptedPassword) > 0 {
		t.Errorf("expected the EncryptedPassword not to be included in the json response")
	}

	if user.FirstName != params.FirstName {
		t.Errorf("expected firstName %s but got %s", params.FirstName, user.FirstName)
	}

	if user.LastName != params.LastName {
		t.Errorf("expected lastName %s but got %s", params.LastName, user.LastName)
	}
}

func TestGetUsers(t *testing.T) {
	tdb := setup(t)
	insertedUser := tdb.seedUsers(t)
	defer tdb.teardown(t)

	app := fiber.New()
	userHandler := NewUserHandler(tdb.UserStore)
	app.Get("/", userHandler.HandleGetUsers)	

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Error(err)
	}
	
	var users []types.User

	json.NewDecoder(resp.Body).Decode(&users)

	if resp.StatusCode != 200 {
		t.Errorf("expected status code to be 200")
	}

	if len(users) != 1 {
		t.Errorf("expected length of users %d but got %d", 1, len(users))
	}

	if users[0].FirstName != insertedUser.FirstName {
		t.Errorf("expected firstName %s but got %s", users[0].FirstName, insertedUser.FirstName)
	}

	if users[0].LastName != insertedUser.LastName {
		t.Errorf("expected lastName %s but got %s", users[0].LastName, insertedUser.LastName)
	}
	
	if users[0].Email != insertedUser.Email {
		t.Errorf("expected email %s but got %s", users[0].Email, insertedUser.Email)
	}
}