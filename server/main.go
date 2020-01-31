package main

import (
	"log"
	"net/http"

	"github.com/mooncascade/toy-banklink/config"
	"github.com/mooncascade/toy-banklink/dao"
	"github.com/mooncascade/toy-banklink/handlers"
	"github.com/mooncascade/toy-banklink/services"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	log.Println("Starting backend.")

	config, err := config.ReadConfiguration()
	if err != nil {
		log.Panicln("Unable to load configuration file", err)
	}

	paymentDAO, err := dao.NewDAO()
	if err != nil {
		log.Panicln("Unable to create new DAO")
	}

	if err = paymentDAO.CreateTable(); err != nil {
		log.Panicln("Unable to create database tables", err)
	}

	log.Println("Created tables.")

	r := mux.NewRouter()
	handler := handlers.GetHandler(services.GetService(paymentDAO, config.ClientID, config.ClientSecret))

	log.Println("Registering handlers.")
	r.HandleFunc("/api/callback", handler.BankCallback).Methods("GET")
	r.HandleFunc("/api/pay", handler.PayEndpoint).Methods("POST")
	r.HandleFunc("/api/banks", handler.GetBanks).Methods("GET")
	r.HandleFunc("/api/payment", handler.PreparePayment).Methods("POST")
	r.HandleFunc("/api/payment/{uuid}", handler.GetPaymentData).Methods("GET")
	r.PathPrefix("/api/").HandlerFunc(corsHandler).Methods("OPTIONS")

	r.Use(basicMiddleware)

	log.Println("Handlers registered.")
	log.Fatal(http.ListenAndServe(":3000", r))
}

func basicMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodOptions {
			w.Header().Add("Content-Type", "application/json")
			corsHandler(w, r)
		}
		next.ServeHTTP(w, r)
	})
}

func corsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Origin") != "" {
		w.Header().Add("Access-Control-Allow-Origin", "http://localhost")

		if r.Method == http.MethodOptions {
			w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
			w.WriteHeader(http.StatusOK)
		}
	}
}
