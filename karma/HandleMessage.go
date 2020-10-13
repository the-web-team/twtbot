package karma

import (
	"context"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"math/rand"
	"regexp"
	"strings"
	"twtbot/db"
)

type Model struct {
	UserId string `bson:"userId"`
	Karma  int32  `bson:"karma"`
}

type Operation struct {
	UserId     string
	KarmaDelta int32
}

func HandleMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) error {
	matches := regexp.MustCompile(`<@!(\d+)> ((--)|(\+\+))`).FindAllString(m.Content, -1)
	triedSelf := false

	if len(matches) > 0 {
		var userIds []string
		var updates []Operation
		for _, match := range matches {
			userId := regexp.MustCompile(`\d+`).FindString(match)
			if userId == m.Author.ID {
				// Cannot give/remove karma to yourself
				triedSelf = true
				continue
			}
			userIds = append(userIds, userId)
			if strings.HasSuffix(match, "++") {
				updates = append(updates, Operation{
					UserId:     userId,
					KarmaDelta: 1,
				})
			} else if strings.HasSuffix(match, "--") {
				updates = append(updates, Operation{
					UserId:     userId,
					KarmaDelta: -1,
				})
			}
		}

		if len(updates) > 0 {
			if updateError := updateKarmaForMultipleUsers(updates); updateError != nil {
				return updateError
			}

			karmaRecords, getError := getKarmaForMultipleUsers(userIds)
			if getError != nil {
				return getError
			}

			var replies []string
			for _, update := range updates {
				user, userError := s.User(update.UserId)
				if userError != nil {
					return userError
				}

				newKarma := karmaRecords[update.UserId]
				if update.KarmaDelta == 1 {
					compliment := compliments[rand.Intn(len(compliments))]
					replies = append(replies, fmt.Sprintf("%s %s (Karma: %d)", user.Mention(), compliment, newKarma))
				} else if update.KarmaDelta == -1 {
					negativeComment := negativeComments[rand.Intn(len(negativeComments))]
					replies = append(replies, fmt.Sprintf("%s %s (Karma: %d)", user.Mention(), negativeComment, newKarma))
				}
			}

			if len(replies) > 0 {
				_, sendError := s.ChannelMessageSend(m.ChannelID, strings.Join(replies, "\n"))
				if sendError != nil {
					return sendError
				}
			}
		}

		if triedSelf {
			if typingError := s.ChannelTyping(m.ChannelID); typingError != nil {
				return typingError
			}
			if _, sendError := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s... You cannot give or take karma from yourself!", m.Author.Mention())); sendError != nil {
				return sendError
			}
		}
	}

	return nil
}

func getKarmaForMultipleUsers(userIds []string) (map[string]int32, error) {
	filter := bson.D{{"userId", bson.D{{"$in", userIds}}}}

	collection := getCollection()
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

func updateKarmaForMultipleUsers(karmaOperations []Operation) error {
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
		collection := getCollection()
		_, bulkWriteError := collection.BulkWrite(context.TODO(), operations, bulkOptions)
		if bulkWriteError != nil {
			return bulkWriteError
		}
	}

	return nil
}

func getCollection() *mongo.Collection {
	_, database := db.Connect()
	return database.Collection("karma")
}

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
