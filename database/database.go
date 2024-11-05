package database

import (
	"context"
	"demo-gosnmp/config"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var SNMPCollection *mongo.Collection

type MongoInstance struct {
	Client *mongo.Client
	Db     *mongo.Database
}

var Mg MongoInstance

func Connect() error {

	config.LoadConfig()

	mongoURI := config.Get("MONGO_URL")
	dbName := config.Get("DB_NAME")

	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(mongoURI).SetMaxPoolSize(50).SetConnectTimeout(10*time.Second))

	if err != nil {
		return err
	}

	// Ping the database to verify connection
	if err := client.Ping(context.TODO(), nil); err != nil {
		return err
	}

	db := client.Database(dbName)

	Mg = MongoInstance{
		Client: client,
		Db:     db,
	}

	SNMPCollection = db.Collection("snpmcollection")

	fmt.Println("Connected to MongoDB!")
	return nil
}
