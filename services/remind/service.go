package remind

import (
	"context"
	"errors"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
	"twtbot/db"
)

var Started = false

type Service struct {
	Session      *discordgo.Session
	reminders    []ReminderModel
	errorChannel chan error
	collection   *mongo.Collection
}

type Reminder struct {
	GuildId   string
	UserId    string
	ChannelId string
	When      time.Time
	What      string
}

type ReminderModel struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	GuildId   string             `bson:"guildId,omitempty"`
	UserId    string             `bson:"userId,omitempty"`
	ChannelId string             `bson:"channelId,omitempty"`
	When      time.Time          `bson:"when,omitempty"`
	What      string             `bson:"what,omitempty"`
}

func (r *Service) SetReminder(reminder Reminder) error {
	res, insertError := r.collection.InsertOne(context.Background(), bson.D{
		{"guildId", reminder.GuildId},
		{"userId", reminder.UserId},
		{"channelId", reminder.ChannelId},
		{"when", reminder.When},
		{"what", reminder.What},
	})
	if insertError != nil {
		return insertError
	}
	if res.InsertedID == nil {
		return errors.New("failed to insert reminder")
	}

	// Find by ID and store

	r.reminders = append(r.reminders, ReminderModel{
		GuildId:   reminder.GuildId,
		UserId:    reminder.UserId,
		ChannelId: reminder.ChannelId,
		When:      reminder.When,
		What:      reminder.What,
	})

	return nil
}

func (r *Service) getAllReminders() ([]ReminderModel, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cur, findError := r.collection.Find(ctx, bson.D{})
	if findError != nil {
		return nil, findError
	}

	defer cur.Close(ctx)

	var reminders []ReminderModel

	for cur.Next(ctx) {
		var reminder ReminderModel
		decodeError := cur.Decode(&reminder)
		if decodeError != nil {
			return nil, decodeError
		}
		reminders = append(reminders, reminder)
	}

	// Return whether there is an error or not
	return reminders, cur.Err()
}

func (r *Service) fireValidReminders() error {
	for _, reminder := range r.reminders {
		if reminder.When.Equal(time.Now()) || reminder.When.Before(time.Now()) {
			// TODO: remove from array?
			member, memberError := r.Session.GuildMember(reminder.GuildId, reminder.UserId)
			if memberError != nil {
				return memberError
			} else {
				mention := member.Mention()
				message := fmt.Sprintf("%s you have a reminder for \"%s\" at \"%s\".", mention, reminder.What, reminder.When.Format(time.RFC822))
				_, sendError := r.Session.ChannelMessageSend(reminder.ChannelId, message)
				if sendError != nil {
					return sendError
				}
			}
		}
	}
	return nil
}

func (r *Service) removeReminder(reminder ReminderModel) error {
	_, deleteError := r.collection.DeleteOne(context.Background(), bson.D{
		{"_id", reminder.ID},
	})
	return deleteError
}

func (r *Service) StartService() error {
	if Started {
		return errors.New("reminders service already started")
	}

	errorChannel := make(chan error)
	r.errorChannel = errorChannel

	// Create Index
	_, database := db.GetConnection()
	r.collection = database.Collection("reminders")

	index := mongo.IndexModel{
		Keys: bson.M{
			"when": 1,
		},
	}
	_, indexError := r.collection.Indexes().CreateOne(context.Background(), index)
	if indexError != nil {
		return indexError
	}

	// Retrieve stored reminders
	reminders, getError := r.getAllReminders()
	if getError != nil {
		return getError
	}

	r.reminders = reminders

	fmt.Println("Reminders Service started.")
	Started = true
	for range time.Tick(time.Second) {
		select {
		case err := <-r.errorChannel:
			return err
		default:
			go func() {
				if reminderError := r.fireValidReminders(); reminderError != nil {
					r.errorChannel <- reminderError
				}
			}()
		}
	}

	return nil
}

func (r *Service) StopService() {

}
