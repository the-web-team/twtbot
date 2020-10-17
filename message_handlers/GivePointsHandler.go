package message_handlers

import (
	"errors"
	"log"
	"twtbot/interfaces"
	"twtbot/points"
)

type GivePointsHandler struct {
	interfaces.MessageHandler
	PointsManager *points.Manager
}

func (h *GivePointsHandler) ShouldRun() bool {
	if h.PointsManager == nil {
		log.Println("PointsManager is nil")
		return false
	}

	return true
}

func (h *GivePointsHandler) Run() error {
	if h.PointsManager == nil {
		return errors.New("pointsmanager must not be nil")
	}

	h.PointsManager.QueueUser(h.Message.Author.ID)
	return nil
}
