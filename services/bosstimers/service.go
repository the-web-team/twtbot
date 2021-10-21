package bosstimers

import (
	"context"
	"errors"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"time"
	"twtbot/db"
)

var Started = false

type Service struct {
	Session       *discordgo.Session
	BossChannelId string
	collection    *mongo.Collection
	errorChannel  chan error
}

type Boss struct {
	ID       primitive.ObjectID `bson:"_id,omitempty"`
	GuildId  string             `bson:"guildId,omitempty"`
	BossName string             `bson:"bossName,omitempty"`
}

type BossTimerModel struct {
	ID                primitive.ObjectID `bson:"_id,omitempty"`
	GuildId           string             `bson:"guildId,omitempty"`
	BossName          string             `bson:"bossName,omitempty"`
	KilledAt          *time.Time         `bson:"killedAt,omitempty"`
	DurationTilWindow time.Duration      `bson:"durationTilWindow,omitempty"`
	WindowDuration    time.Duration      `bson:"windowDuration,omitempty"`
}

func (s *Service) RegisterBoss(guildId string, bossName string, durationTilWindow time.Duration, windowDuration time.Duration) error {
	res, insertError := s.collection.InsertOne(context.Background(), bson.D{
		{"guildId", guildId},
		{"bossName", bossName},
		{"durationTilWindow", durationTilWindow},
		{"windowDuration", windowDuration},
	})
	if insertError != nil {
		return insertError
	}
	if res.InsertedID == nil {
		return errors.New("failed to register boss")
	}
	return nil
}

func (s *Service) getBossTimers() ([]BossTimerModel, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cursor, findError := s.collection.Find(ctx, bson.D{})
	if findError != nil {
		return nil, findError
	}

	defer cursor.Close(ctx)

	var bossTimers []BossTimerModel

	for cursor.Next(ctx) {
		var bossTimer BossTimerModel
		decodeError := cursor.Decode(&bossTimer)
		if decodeError != nil {
			return nil, decodeError
		}
		bossTimers = append(bossTimers, bossTimer)
	}

	return bossTimers, cursor.Err()
}

func (s *Service) checkBoss(bossTimer BossTimerModel) error {
	now := time.Now()
	if bossTimer.KilledAt != nil {
		killedAt := *bossTimer.KilledAt
		windowOpen := killedAt.Add(bossTimer.DurationTilWindow)
		windowClose := windowOpen.Add(bossTimer.WindowDuration)

		if windowOpen.After(now) {
			duration := windowOpen.Sub(now)
			if duration.Minutes() == 0 {
				actionMessage := "window opens"
				if windowOpen.Equal(windowClose) {
					actionMessage = "spawns"
				}
				message := fmt.Sprintf("%s %s in %d hours", bossTimer.BossName, actionMessage, duration.Hours())
				_, sendError := s.Session.ChannelMessageSend(s.BossChannelId, message)
				if sendError != nil {
					return sendError
				}
			}
		}
	}
	return nil
}

func (s *Service) StartService() error {
	fmt.Println("Starting bosstimers service...")
	if Started {
		return errors.New("bosstimers service already started")
	}

	s.errorChannel = make(chan error)

	// Create Index
	_, database := db.GetConnection()
	s.collection = database.Collection("bosstimers")

	index := mongo.IndexModel{
		Keys: bson.M{
			"guildId": 1,
		},
	}
	_, indexError := s.collection.Indexes().CreateOne(context.Background(), index)
	if indexError != nil {
		fmt.Println(indexError)
		return indexError
	}

	fmt.Println("Boss Timers Service started")
	Started = true
	for range time.Tick(time.Minute) {
		select {
		case err := <-s.errorChannel:
			return err
		default:
			go func() {
				timers, getError := s.getBossTimers()
				if getError != nil {
					log.Fatalln(getError)
				}

				for _, timer := range timers {
					checkError := s.checkBoss(timer)
					log.Fatalln(checkError)
				}
			}()
		}
	}

	return nil
}

func (s *Service) StopService() {
	close(s.errorChannel)
	Started = false
}
