package cmd

import (
	"context"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gitlab.com/olaris/olaris-server/metadata"
	"gitlab.com/olaris/olaris-server/metadata/app"
	"gitlab.com/olaris/olaris-server/react"
	"gitlab.com/olaris/olaris-server/streaming"
	"net/http"
	"os"
	"os/signal"
	"time"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the olaris server",
	Run: func(cmd *cobra.Command, args []string) {
		r := mux.NewRouter()

		r.PathPrefix("/s").Handler(http.StripPrefix("/s", streaming.GetHandler()))
		defer streaming.Cleanup()

		mctx := app.NewDefaultMDContext()
		defer mctx.Db.Close()

		r.PathPrefix("/app").Handler(http.StripPrefix("/app", react.GetHandler()))

		r.PathPrefix("/m").Handler(http.StripPrefix("/m", metadata.GetHandler(mctx)))

		handler := handlers.LoggingHandler(os.Stdout, r)

		var port = os.Getenv("PORT")
		// Set a default port if there is nothing in the environment
		if port == "" {
			port = "8080"
		}
		log.Infoln("binding on port", port)
		srv := &http.Server{Addr: ":" + port, Handler: handler}
		go func() {
			if err := srv.ListenAndServe(); err != nil {
				log.WithFields(log.Fields{"error": err}).Fatal("Error starting server.")
			}
		}()

		stopChan := make(chan os.Signal)
		signal.Notify(stopChan, os.Interrupt)

		// Wait for termination signal
		<-stopChan
		log.Println("Shutting down...")

		//	mctx.ExitChan <- 1
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		srv.Shutdown(ctx)
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
}