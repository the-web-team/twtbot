package points

import (
	"context"
	"errors"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"sync"
	"time"
	"twtbot/db"
)

type Manager struct {
	lock         *sync.Mutex
	queuedPoints map[string]int32
	errorChannel chan error
}

func (m *Manager) HandleMessageCreate(_ *discordgo.Session, msg *discordgo.MessageCreate) {
	go m.QueueUser(msg.Author.ID)

}

func (m *Manager) QueueUser(userId string) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.queuedPoints[userId]++
}

func (m *Manager) StopService() {
	m.errorChannel <- errors.New("points service stopped")
}

func (m *Manager) StartService() error {
	fmt.Println("Points Manager started.")
	for range time.Tick(10 * time.Second) {
		select {
		case err := <-m.errorChannel:
			return err
		default:
			go func() {
				if awardError := m.awardPoints(); awardError != nil {
					m.errorChannel <- awardError
				}
			}()
		}
	}

	return nil
}

func (m *Manager) awardPoints() error {
	m.lock.Lock()
	defer m.lock.Unlock()

	if len(m.queuedPoints) > 0 {
		var operations []mongo.WriteModel
		bulkOptions := options.BulkWrite()

		for userId, points := range m.queuedPoints {
			operation := mongo.NewUpdateOneModel()
			operation.SetUpsert(true)

			filter := bson.D{{"userId", userId}}
			update := bson.D{
				{"$inc", bson.D{{"points", points}}},
			}
			operation.SetFilter(filter)
			operation.SetUpdate(update)
			operations = append(operations, operation)
		}

		log.Println(fmt.Sprintf("Awarding points to %d users.", len(operations)))

		if len(operations) > 0 {
			_, database := db.Connect()
			collection := database.Collection("points")
			if _, bulkWriteError := collection.BulkWrite(context.TODO(), operations, bulkOptions); bulkWriteError != nil {
				return bulkWriteError
			}
		}

		m.queuedPoints = nil
	}

	return nil
}
