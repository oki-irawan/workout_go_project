package main

import (
	"flag"
	"fmt"
	"github.com/oki-irawan/fem_project/internal/app"
	"github.com/oki-irawan/fem_project/internal/routes"
	"net/http"
	"time"
)

func main() {
	var port int
	flag.IntVar(&port, "port", 8080, "port to listen on")
	flag.Parse()

	app, err := app.NewApplication()
	if err != nil {
		panic(err)
	}
	defer app.DB.Close()

	r := routes.SetupRoutes(app)

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      r,
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	app.Logger.Printf("We are running on port %d\n", port)

	err = server.ListenAndServe()
	if err != nil {
		app.Logger.Fatal(err)
	}

}
