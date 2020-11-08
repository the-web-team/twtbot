package getuserpointshandler

import (
	"fmt"
	"log"
	"twtbot/interfaces"
	"twtbot/services/points"
)

type Handler struct {
	interfaces.MessageHandler
	PointsManager *points.Manager
}

func (h *Handler) ShouldRun() bool {
	if h.PointsManager == nil {
		log.Println("PointsManager is nil")
		return false
	}

	return h.IsCommand("points")
}

func (h *Handler) Run() error {
	numPoints := h.PointsManager.GetUserPoints(h.Message.Author.ID)
	reply := fmt.Sprintf("you have `%d` points!", numPoints)
	return h.ReplyWithMention(reply)
}
