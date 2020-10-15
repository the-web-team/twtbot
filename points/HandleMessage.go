package points

import (
	"context"
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

var pointsManager *Manager

func HandleMessageCreate(_ *discordgo.Session, m *discordgo.MessageCreate) error {
	pointsManager.QueueUser(m.Author.ID)
	return nil
}

type Manager struct {
	sync.Mutex
	queuedPoints []string
}

func StartService() {
	if pointsManager == nil {
		pointsManager = new(Manager)
	}

	go pointsManager.RunManager()
}

func (p *Manager) QueueUser(userId string) {
	p.Lock()
	defer p.Unlock()
	p.queuedPoints = append(p.queuedPoints, userId)
}

func (p *Manager) RunManager() {
	for range time.Tick(10 * time.Second) {
		go func() {
			if awardError := p.AwardPoints(); awardError != nil {
				log.Fatal(awardError)
			}
		}()
	}
}

func (p *Manager) AwardPoints() error {
	p.Lock()
	defer p.Unlock()
	if len(p.queuedPoints) > 0 {
		var operations []mongo.WriteModel
		bulkOptions := options.BulkWrite()

		for _, userId := range p.queuedPoints {
			operation := mongo.NewUpdateOneModel()
			operation.SetUpsert(true)

			filter := bson.D{{"userId", userId}}
			update := bson.D{
				{"$inc", bson.D{{"points", 1}}},
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
		p.queuedPoints = nil
	}

	return nil
}
