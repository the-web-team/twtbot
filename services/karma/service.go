package karma

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"twtbot/db"
	"twtbot/lib"
)

type Model struct {
	UserId string `bson:"userId"`
	Karma  int32  `bson:"karma"`
}

type UserOperation struct {
	UserId     string
	KarmaDelta int32
}

type GroupOperation struct {
	GroupId    string
	KarmaDelta int32
}

type Service struct{}

func (s *Service) GetUsersKarma(userIds []string) (map[string]int32, error) {
	userIdSet := &lib.Set{}
	for _, userId := range userIds {
		userIdSet.Add(userId)
	}
	var userIdList []string
	for userId, _ := range userIdSet.GetItems() {
		userIdList = append(userIdList, userId.(string))
	}

	filter := bson.D{{"userId", bson.D{{"$in", userIdList}}}}

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

func (s *Service) UpdateUsersKarma(karmaOperations map[string]int32) error {
	if len(karmaOperations) > 0 {
		usersToBeUpdated := &lib.Set{}

		for userId := range karmaOperations {
			usersToBeUpdated.Add(userId)
		}

		var operations []mongo.WriteModel
		bulkOptions := options.BulkWrite()

		for userId := range usersToBeUpdated.GetItems() {
			userIdString := userId.(string)
			operation := mongo.NewUpdateManyModel()
			operation.SetUpsert(true)
			operation.SetFilter(bson.D{{"userId", userIdString}})
			operation.SetUpdate(bson.D{
				{"$inc", bson.D{{"karma", karmaOperations[userIdString]}}},
			})
			operations = append(operations, operation)
		}

		if len(operations) > 0 {
			_, database := db.GetConnection()
			collection := database.Collection("karma")
			if _, bulkWriteError := collection.BulkWrite(context.TODO(), operations, bulkOptions); bulkWriteError != nil {
				return bulkWriteError
			}
		}
	}
	return nil
}

func (s *Service) getCollection() *mongo.Collection {
	_, database := db.GetConnection()
	return database.Collection("karma")
}
