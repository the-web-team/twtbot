package karma

import (
	"context"
	"github.com/bwmarrin/discordgo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"twtbot/db"
)

var compliments = []string{
	"you are a gift to those around you!",
	"is awesome!",
	"is a smart cookie!",
	"has got swag!",
	"you are appreciated!",
	"is string!",
	"is inspiring!",
	"on a scale from 1 to 10, you're an 11!",
	"is a ray of sunshine!",
	"is making a difference!",
	"brings out the best in us!",
	"thank you!",
}

var negativeComments = []string{
	"my middle finger salutes you!",
	"is disappointing!",
	"rip...",
	"nope...",
	"fail!",
	"gtfo!",
}

type Model struct {
	UserId string `bson:"userId"`
	Karma  int32  `bson:"karma"`
}

type Operation struct {
	UserId     string
	KarmaDelta int32
}

type Service struct {
	session *discordgo.Session
}

func (s *Service) getUsersKarma(userIds []string) (map[string]int32, error) {
	filter := bson.D{{"userId", bson.D{{"$in", userIds}}}}

	collection := s.getCollection()
	ctx := context.Background()
	cursor, findErr := collection.Find(ctx, filter)
	if findErr != nil {
		return nil, findErr
	}
	defer func() {
		_ = cursor.Close(ctx)
	}()

	karmaRecords := make(map[string]int32)
	for cursor.Next(ctx) {
		var karmaRecord Model
		if cursorError := cursor.Decode(&karmaRecord); cursorError != nil {
			return nil, cursorError
		}
		karmaRecords[karmaRecord.UserId] = karmaRecord.Karma
	}

	return karmaRecords, nil
}

func (s *Service) updateUsersKarma(karmaOperations []Operation) error {
	var operations []mongo.WriteModel
	bulkOptions := options.BulkWrite()

	for _, op := range karmaOperations {
		operation := mongo.NewUpdateOneModel()
		operation.SetUpsert(true)

		filter := bson.D{{"userId", op.UserId}}
		update := bson.D{
			{"$inc", bson.D{{"karma", op.KarmaDelta}}},
		}
		operation.SetFilter(filter)
		operation.SetUpdate(update)
		operations = append(operations, operation)
	}

	if len(operations) > 0 {
		collection := s.getCollection()
		_, bulkWriteError := collection.BulkWrite(context.TODO(), operations, bulkOptions)
		if bulkWriteError != nil {
			return bulkWriteError
		}
	}

	return nil
}

func (s *Service) getCollection() *mongo.Collection {
	_, database := db.Connect()
	return database.Collection("karma")
}
