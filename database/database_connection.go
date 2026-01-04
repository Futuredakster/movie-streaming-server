package database

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// DATABASE CONNECTION EXPLAINED (coming from Node.js):
// =====================================================
// In Node.js: mongoose.connect(process.env.MONGODB_URI) - done!
// In Go: Much more verbose due to:
// 1. No ORM like Mongoose - using direct MongoDB driver
// 2. Manual error handling required everywhere
// 3. Explicit connection configuration
// 4. Type safety - everything must be defined

// POINTER USAGE EXPLANATION (coming from JavaScript):
// ===================================================
// JavaScript: Everything is automatically a reference
// Go: You must choose - copy (expensive) or pointer (efficient)
//
// RULE: Database connections are HUGE objects containing:
// - Connection pools, network connections, auth data, config
// Copying would waste memory and break connections
//
// *mongo.Client = pointer to mongo.Client (memory address)
// mongo.Client = actual mongo.Client object (full copy)
//
// MongoDB functions return pointers because connections are expensive to copy

// dbInstance creates the MongoDB connection
// Returns: *mongo.Client (pointer/address, not copy)
func dbInstance() *mongo.Client {
	// Load environment variables from .env file
	// Node.js equivalent: require('dotenv').config()
	err := godotenv.Load(".env")

	if err != nil {
		log.Println("Warning: Cant find the .env file")
	}

	// Get connection string from environment
	// Node.js equivalent: process.env.MONGODB_URI
	MongoDb := os.Getenv("MONGODB_URI")

	if MongoDb == "" {
		log.Fatal("MONGODB_URI not retrived")
	}

	fmt.Println(MongoDb)

	// Configure connection options
	// Node.js equivalent: mongoose handles this automatically
	clientOptions := options.Client().ApplyURI(MongoDb)

	// Actually connect to MongoDB
	// Returns: *mongo.Client (pointer) - address of connection, not copy
	client, err := mongo.Connect(clientOptions)

	if err != nil {
		return nil
	}

	return client // Returns pointer (address) to avoid copying expensive connection
}

// Global client instance - stores POINTER to connection
// Why pointer? Sharing same connection across app (no copying)
// Node.js equivalent: mongoose handles this internally
var Client *mongo.Client = dbInstance()

// OpenCollection gets a specific collection
// Returns: *mongo.Collection (pointer) for same efficiency reasons
func OpenCollection(collectionName string) *mongo.Collection {
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("Warning: Cant find the .env file")
	}

	databaseName := os.Getenv("DATABASE_NAME")
	fmt.Println(databaseName)

	// Get collection from database
	// Client is pointer, so we use . (not ->) like JavaScript obj.method()
	// Returns pointer to collection (efficient)
	collection := Client.Database(databaseName).Collection(collectionName)

	if collection == nil {
		return nil
	}

	return collection // Return pointer to collection (address, not copy)
}
