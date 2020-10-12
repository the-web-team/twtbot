package karma

import (
	"context"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"math/rand"
	"regexp"
	"strings"
	"twtbot/db"
)

const Increment = "Increment"
const Decrement = "Decrement"

type Model struct {
	UserId string `bson:"userId"`
	Karma int32 `bson:"karma"`
}

type Operation struct {
	UserId string
	KarmaDelta int32
}

func HandleMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	log.Println(m.Content)
	incRegex := regexp.MustCompile(`<@!(\d+)> ((--)|(\+\+))`)
	matches := incRegex.FindAllString(m.Content, -1)
	triedSelf := false

	if len(matches) > 0 {
		var userIds []string
		var updates []Operation
		for _, match := range matches {
			userIdRegex := regexp.MustCompile(`\d+`)
			userId := userIdRegex.FindString(match)
			if userId == m.Author.ID {
				// Cannot give/remove karma to yourself
				triedSelf = true
				continue
			}
			userIds = append(userIds, userId)
			if strings.HasSuffix(match, "++") {
				updates = append(updates, Operation{
					UserId: userId,
					KarmaDelta: 1,
				})
			} else if strings.HasSuffix(match, "--") {
				updates = append(updates, Operation{
					UserId: userId,
					KarmaDelta: -1,
				})
			}
		}

		if len(updates) > 0 {
			updateKarmaForMultipleUsers(updates)
			karmaRecords := getKarmaForMultipleUsers(userIds)

			var replies []string
			for _, update := range updates {
				user, userError := s.User(update.UserId)
				if userError != nil {
					log.Fatal(userError)
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
				_, sendError := s.ChannelMessageSend(m.ChannelID, strings.Join(replies, ", "))
				if sendError != nil {
					log.Fatal(sendError)
				}
			}
		}

		if triedSelf {
			_, sendError := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s... You cannot give or take karma from yourself!", m.Author.Mention()))
			if sendError != nil {
				log.Fatal(sendError)
			}
		}
	}
}

func getKarmaForMultipleUsers(userIds []string) map[string]int32 {
	filter := bson.D{{"userId", bson.D{{"$in", userIds}}}}

	collection := getCollection()
	ctx := context.Background()
	cursor, findErr := collection.Find(ctx, filter)
	if findErr != nil {
		log.Fatal(findErr)
	}
	defer cursor.Close(ctx)

	karmaRecords := make(map[string]int32)
	for cursor.Next(ctx) {
		var karmaRecord Model
		if cursorError := cursor.Decode(&karmaRecord); cursorError != nil {
			log.Fatal(cursorError)
		}
		karmaRecords[karmaRecord.UserId] = karmaRecord.Karma
	}
	fmt.Println(karmaRecords)
	return karmaRecords
}

func updateKarmaForMultipleUsers(karmaOperations []Operation) {
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
			log.Fatal(bulkWriteError)
		}
	}
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