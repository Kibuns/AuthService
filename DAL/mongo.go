package DAL

import (
	"context"
	"errors"
	"fmt"

	"github.com/Kibuns/AuthService/Models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// global variable mongodb connection client
var client mongo.Client = newClient()

// ----Create----
func RegisterUser(detailedUser Models.DetailedUser) error {
	//simplify detaileduser to user
	var user Models.User
	user.UserName = detailedUser.UserName
	user.Password = detailedUser.Password


	userCollection := client.Database("AuthDB").Collection("credentials")
	_, err := userCollection.InsertOne(context.TODO(), user)
	if err != nil {
		return err
	}

	fmt.Println("New user added called: " + user.UserName)
	return nil
}

//----Read----

func ReadAllUsers() ([]primitive.M, error) {
	twootCollection := client.Database("AuthDB").Collection("users")
	// retrieve all the documents (empty filter)
	cursor, err := twootCollection.Find(context.TODO(), bson.D{})
	if err != nil {
		return nil, err
	}

	var results []bson.M
	if err = cursor.All(context.TODO(), &results); err != nil {
		return nil, err
	}

	return results, nil
}



func SearchUser(username string) (primitive.M, error) {
	userCollection := client.Database("AuthDB").Collection("credentials")

	// Create a filter to search for the document with the specified username
	filter := bson.M{"username": username}

	fmt.Println(username)

	// Find the first document that matches the filter
	var result bson.M
	err := userCollection.FindOne(context.TODO(), filter).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	// Display the retrieved document
	fmt.Println("Displaying the result from the search query")
	fmt.Println(result)

	return result, nil
}


//----Update----

//----Delete----

// other
func newClient() (value mongo.Client) {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		panic(err)
	}
	if err := client.Ping(context.TODO(), readpref.Primary()); err != nil {
		panic(err)
	}
	value = *client

	return
}