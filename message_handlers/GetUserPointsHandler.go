package message_handlers

import (
	"fmt"
	"log"
	"twtbot/interfaces"
	"twtbot/points"
)

type GetUserPointsHandler struct {
	interfaces.MessageHandler
	PointsManager *points.Manager
}

func (h *GetUserPointsHandler) ShouldRun() bool {
	if h.PointsManager == nil {
		log.Println("PointsManager is nil")
		return false
	}

	return h.IsCommand("points")
}

func (h *GetUserPointsHandler) Run() error {
	numPoints := h.PointsManager.GetUserPoints(h.Message.Author.ID)
	reply := fmt.Sprintf("you have `%d` points!", numPoints)
	return h.ReplyWithMention(reply)
}
