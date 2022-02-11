package server

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

// Options is a struct that contains all the options required to config
// the server
type Options struct {
	Logger  *log.Logger
	Addr    string
	Handler http.Handler
}

// serverCreated is used to guarantee that only one instance
// of the server is created.
var serverCreated = false

func Run(opts *Options) {
	// reject if server was already created
	if serverCreated {
		errMsg := "server instance already running"
		opts.Logger.Panicln(errMsg)
	}

	// Config server options
	server := &http.Server{
		Addr:         opts.Addr,
		Handler:      opts.Handler,
		WriteTimeout: 5 * time.Second,
		ReadTimeout:  10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// start server
	go func() {
		serverCreated = true

		err := server.ListenAndServe()
		if err != http.ErrServerClosed {
			opts.Logger.Panicln("failed to start server instance:", err.Error())
			serverCreated = false
		}
	}()

	// config server for listening for shutdown signals (kill or interrupt)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Kill)
	signal.Notify(sigChan, os.Interrupt)

	// waitisten for gracefull shutdown signals
	sig := <-sigChan
	opts.Logger.Println("received graceful shutdown - shuting down server:", sig)

	// forcefully shutdown server after 30 seconds if there are pending jobs
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	server.Shutdown(ctx)
}
