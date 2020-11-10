package givepoints

import (
	"errors"
	"log"
	"twtbot/interfaces"
	"twtbot/services/points"
)

type Handler struct {
	interfaces.MessageHandler
	PointsManager *points.Service
}

func (h *Handler) ShouldRun() bool {
	if h.PointsManager == nil {
		log.Println("PointsManager is nil")
		return false
	}

	return true
}

func (h *Handler) Run() error {
	if h.PointsManager == nil {
		return errors.New("pointsmanager must not be nil")
	}

	h.PointsManager.QueueUser(h.Message.Author.ID)
	return nil
}
