package sessionscleanerservice

import (
	"time"

	"com.code.vidmicro/com.code.vidmicro/httpHandler/basecontrollers"
	"com.code.vidmicro/com.code.vidmicro/httpHandler/basecontrollers/baseconst"
	"com.code.vidmicro/com.code.vidmicro/settings/serviceutils"
)

type SessionsCleanerService struct {
}

func (u *SessionsCleanerService) Run() error {

	u.AssignSubscriber()
	serviceutils.GetInstance().RunService()
	u.RunServer()

	return nil
}

func (u *SessionsCleanerService) RunServer() {

	u.cleanSessions()

	// Create a ticker that ticks every 24 hours
	ticker := time.NewTicker(24 * time.Hour)

	// Run the task at the specified intervals
	for range ticker.C {
		u.cleanSessions()
	}

}

func (u *SessionsCleanerService) cleanSessions() {
	controller, _ := basecontrollers.GetInstance().GetController(baseconst.UsersSessions)

	now := time.Now().Unix()
	past24Hours := now - (24 * 60 * 60)

	query := `DELETE FROM users_sessions WHERE expiring_at >= $1 AND expiring_at <= $2;`
	controller.RawQuery(controller.GetDBName(), controller.GetCollectionName(), query, []interface{}{past24Hours, now})
}

func (u *SessionsCleanerService) AssignSubscriber() error {
	return nil
}

func (u *SessionsCleanerService) Stop() {
	serviceutils.GetInstance().Shutdown()
}
