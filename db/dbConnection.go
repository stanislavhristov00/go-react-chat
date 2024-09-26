package db

import (
	"chat-module/models"
	"chat-module/util"
	"context"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoRepo struct {
	MongoClient *mongo.Client
}

func newMongoRepo() *MongoRepo {
	err := util.LoadEnvFile()

	if err != nil {
		log.Fatalf("On initializing the .env file")
	}

	MongoDb := os.Getenv("MONGODB_URL")

	// Give it max 10 seconds to connect
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(MongoDb))

	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	err = client.Ping(ctx, nil /* rp */)

	if err != nil {
		log.Fatalf("Failed to ping the client: %v", err)
	}

	log.Println("Successfully established connection to database")

	return &MongoRepo{
		MongoClient: client,
	}
}

/*
 *	Singleton db instance. Outside packages should access the database through
 *	this object. MongoDB client is thread-safe so we can use it across multiple
 * 	threads without worries.
 */
var Client *MongoRepo = newMongoRepo()

func indexExists(collection *mongo.Collection, indexName string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	cursor, err := collection.Indexes().List(ctx)

	if err != nil {
		return false, err
	}

	defer cursor.Close(ctx)

	var index bson.M
	for cursor.Next(ctx) {
		if err := cursor.Decode(&index); err != nil {
			return false, err
		}

		if index["name"] == indexName {
			return true, nil
		}
	}

	return false, nil
}

/*
 * 	Function that opens/creates a collection in our database.
 */

func openCollection(client *mongo.Client, collectionName string) *mongo.Collection {
	if collectionName == "" {
		log.Fatal("Tried to create an empty collection")
	}

	// DB_NAME should be loaded at Init time, so we don't have to check it here.
	collection := client.Database(os.Getenv("DB_NAME")).Collection(collectionName)

	return collection
}

/*
 *	Creates a unique index with a name for a given field. Assumes that the collection
 *	already exists.
 */

func createUniqueIndex(collection *mongo.Collection, field, name string) error {
	exists, err := indexExists(collection, name)

	if err != nil {
		log.Printf("Failed to query for index %s: %v", name, err)
		return err
	}

	if exists {
		return nil
	}

	options := options.Index()
	options.SetUnique(true)
	options.SetName(name)

	// Define the index model
	indexModel := mongo.IndexModel{
		Keys: bson.M{
			field: 1, // Create an ascending index on the "email" field
		},
		Options: options, // Optional: Set the index to be unique
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	// Create the index
	indexName, err := collection.Indexes().CreateOne(ctx, indexModel)

	if err != nil {
		log.Printf("Error creating index: %v", err)
		return err
	}

	log.Printf("Index created with name: %v", indexName)

	return nil
}

/*
 *	Initializes the database. For now this is just opening the collections, so they
 *	are initialized. If this is the first time the collections are
 *	created, creates unique indexes also.
 */

func Init() error {
	collection := openCollection(Client.MongoClient, os.Getenv("USER_DOCUMENT"))

	err := createUniqueIndex(collection, "email", "users-email-index")

	if err != nil {
		log.Printf("Failed to create unique index for email: %v", err)
		return err
	}

	err = createUniqueIndex(collection, "username", "users-username-index")

	if err != nil {
		log.Printf("Failed to create unique index for username: %v", err)
		return err
	}

	return nil
}

// INTERFACE METHODS

func (repo *MongoRepo) CheckUserExists(username, email string) (bool, error) {
	collection := openCollection(repo.MongoClient, os.Getenv("USER_DOCUMENT"))

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	filter := bson.M{
		"$or": []bson.M{
			{"username": username},
			{"email": email},
		},
	}

	// Check if we already have an existing user with these credentials
	count, err := collection.CountDocuments(ctx, filter)

	if err != nil {
		return false, err
	}

	if count > 0 {
		return true, nil
	}

	return false, nil
}

func (repo *MongoRepo) AddUser(user models.User) error {
	collection := openCollection(repo.MongoClient, os.Getenv("USER_DOCUMENT"))

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	_, insertErr := collection.InsertOne(ctx, user)

	if insertErr != nil {
		return insertErr
	}

	return nil
}

func (repo *MongoRepo) GetUser(usernameOrEmail string) (*models.User, error) {
	collection := openCollection(repo.MongoClient, os.Getenv("USER_DOCUMENT"))
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	filter := bson.M{
		"$or": []bson.M{
			{"username": usernameOrEmail},
			{"email": usernameOrEmail},
		},
	}

	var user models.User
	err := collection.FindOne(ctx, filter).Decode(&user)

	if err != nil {
		return nil, err
	}

	return &user, nil
}
