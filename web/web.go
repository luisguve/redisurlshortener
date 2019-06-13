package web

import (
	"fmt"
	"net/http"
	"time"

	"github.com/villegasl/urlshortener.redis/web/api"
	"github.com/villegasl/urlshortener.redis/web/www"
	"github.com/villegasl/urlshortener.redis/models"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

func Start(DB_Handler *models.DBHandler, port string) {
	log.Info("web starts")
	router := mux.NewRouter()

	router.HandleFunc("/", www.Index).Methods(http.MethodGet)

	sub := router.PathPrefix("/api/shorturl").SubRouter()
	sub.Handle("/{url:[a-zA-Z0-9]{1,11}}",
		api.RedirectByShortURL(DB_Handler)).Methods(http.MethodGet)
	sub.Handle("/new",
		api.NewShortURL(DB_Handler)).Methods(http.MethodPost)

	//serve the static files (CSS)
	cssHandler := http.StripPrefix("/static", http.FileServer(http.Dir("./static")))
	router.PathPrefix("/static/").Handler(cssHandler)

	srv := http.Server {
		Addr: 			":" + port,
		Handler: 		router,
		ReadTimeout: 	10 * time.Second,
		WriteTimeout: 	10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	fmt.Println("visit localhost:8080")
	log.Fatal(srv.ListenAndServe())
}