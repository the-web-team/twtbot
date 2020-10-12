package db

import (
	"context"
	"errors"
	"flag"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"log"
	"os"
)

var dbConn *mongo.Client

func Connect() (*mongo.Client, *mongo.Database) {
	if dbConn == nil {
		uri := getUri()
		connOpts := options.Client().ApplyURI(uri).SetReadPreference(readpref.SecondaryPreferred())
		session, sessionError := mongo.NewClient(connOpts)
		if sessionError != nil {
			log.Fatal(sessionError)
		}

		if connectError := session.Connect(context.Background()); connectError != nil {
			log.Fatal(connectError)
		}
		if pingErr := session.Ping(context.TODO(), readpref.SecondaryPreferred()); pingErr != nil {
			log.Fatal(pingErr)
		}

		dbConn = session
	}

	return dbConn, dbConn.Database(os.Getenv("MONGO_DATABASE_NAME"))
}

func getUri() string {
	var uri string
	flag.StringVar(&uri, "m", "", "Mongo URI")
	flag.Parse()

	if uri == "" {
		uri = os.Getenv("MONGO_URI")
	}

	if uri == "" {
		log.Fatal(errors.New("invalid mongo uri"))
	}

	return uri
}
