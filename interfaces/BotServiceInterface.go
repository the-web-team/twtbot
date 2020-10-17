package interfaces

type BotService interface {
	StartService() error
	StopService()
}
