package points

import (
	"context"
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"strings"
	"sync"
	"time"
	"twtbot/db"
)

var Started = false

type Service struct {
	sync.Mutex
	queuedPoints map[string]int32
	errorChannel chan error
}

type Model struct {
	userId string
	points int32
}

func (m *Service) QueueUser(userId string) {
	m.Lock()
	defer m.Unlock()
	if m.queuedPoints == nil {
		m.queuedPoints = make(map[string]int32)
	}
	m.queuedPoints[userId]++
}

func (m *Service) StopService() {
	if awardError := m.awardPoints(); awardError != nil {
		log.Println(awardError)
	}
	m.errorChannel <- errors.New("points service stopped")
}

func (m *Service) StartService() error {
	if Started {
		return errors.New("points service already started")
	}

	errorChannel := make(chan error)
	m.errorChannel = errorChannel

	fmt.Println("Points Service started.")
	for range time.Tick(10 * time.Second) {
		select {
		case err := <-m.errorChannel:
			return err
		default:
			go func() {
				Started = true
				if awardError := m.awardPoints(); awardError != nil {
					m.errorChannel <- awardError
				}
			}()
		}
	}

	return nil
}

func (m *Service) GetUserPoints(userId string) int32 {
	_, database := db.GetConnection()
	collection := database.Collection("points")

	var result Model
	filter := bson.D{{"userId", userId}}
	findError := collection.FindOne(context.TODO(), filter).Decode(&result)
	if findError != nil {
		fmt.Println(findError)
		return 0
	}

	fmt.Println(result)

	return result.points
}

func (m *Service) awardPoints() error {
	m.Lock()
	defer m.Unlock()

	if len(m.queuedPoints) > 0 {
		var operations []mongo.WriteModel
		bulkOptions := options.BulkWrite()

		var rewardees []string

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
			rewardees = append(rewardees, userId)
		}

		numOperations := len(operations)
		stringRewardees := strings.Join(rewardees, ", ")
		log.Println(fmt.Sprintf("Awarding points to %d users. %s", numOperations, stringRewardees))

		if numOperations > 0 {
			_, database := db.GetConnection()
			collection := database.Collection("points")
			if _, bulkWriteError := collection.BulkWrite(context.TODO(), operations, bulkOptions); bulkWriteError != nil {
				return bulkWriteError
			}
		}

		m.queuedPoints = nil
	}

	return nil
}

func (m *Service) reply(message string) {

}
