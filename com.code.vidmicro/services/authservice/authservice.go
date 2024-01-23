package authservice

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"com.code.vidmicro/com.code.vidmicro/httpHandler"
	"com.code.vidmicro/com.code.vidmicro/httpHandler/basecontrollers"
	"com.code.vidmicro/com.code.vidmicro/settings/configmanager"
	"com.code.vidmicro/com.code.vidmicro/settings/serviceutils"
)

type AuthService struct {
	sb serviceutils.SubscriptionInterface
}

func (u *AuthService) Run() error {

	u.AssignSubscriber()
	serviceutils.GetInstance().RunService()
	u.RunServer()

	return nil
}

func (u *AuthService) RunServer() {

	srv := &http.Server{
		Addr:    configmanager.GetInstance().Address,
		Handler: httpHandler.GetInstance().GetEngine(),
	}
	basecontrollers.GetInstance().RegisterControllers()

	// Initializing the server in a goroutine so that
	// it won't block the graceful shutdown handling below
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal, 1)
	// kill (no param) default send syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can't be catch, so don't need add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown: ", err)
	}

	log.Println("Server exiting")
}

func (u *AuthService) AssignSubscriber() error {
	return nil
}

func (u *AuthService) Stop() {
	serviceutils.GetInstance().Shutdown()
}
