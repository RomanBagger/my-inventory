package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

type App struct {
	Router *mux.Router
	DB     *sql.DB
}

func (app *App) Initialize(DbUser string, DbPassword string, PublicIP string, Port int, DbName string) error {
	connectionString := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", DbUser, DbPassword, PublicIP, Port, DbName)
	var err error
	app.DB, err = sql.Open("mysql", connectionString)
	if err != nil {
		return err
	}

	err = app.DB.Ping()
	if err != nil {
		log.Printf("Error pinging database: %v", err)
		return err
	}

	app.Router = mux.NewRouter().StrictSlash(true)
	app.handlerRoutes()
	log.Println("Database connection successful and routes initialized.")
	return nil
}

func (app *App) Run(address string) {
	log.Fatal(http.ListenAndServe(address, app.Router))
}

func sendResponse(w http.ResponseWriter, statusCode int, payload interface{}) {
	responce, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(responce)
}

func sendError(w http.ResponseWriter, statusCode int, err error) {
	error_message := map[string]string{"error": err.Error()}
	sendResponse(w, statusCode, error_message)
}

func (app *App) getProducts(w http.ResponseWriter, r *http.Request) {
	products, err := getProducts(app.DB)
	if err != nil {
		sendError(w, http.StatusInternalServerError, err)
		return
	}
	sendResponse(w, http.StatusOK, products)
}

func (app *App) getProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key, err := strconv.Atoi(vars["id"])
	if err != nil {
		sendError(w, http.StatusBadRequest, fmt.Errorf("invalid product ID"))
		return
	}

	p := product{ID:key}
	err = p.getProduct(app.DB)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			sendError(w, http.StatusNotFound, fmt.Errorf("product not found"))
		default:
			sendError(w, http.StatusInternalServerError, err)
		}
		return
	}
	sendResponse(w, http.StatusOK, p)
}

func (app *App) createProduct(w http.ResponseWriter, r *http.Request) {
	var p product
	
	err := json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		sendError(w, http.StatusBadRequest, fmt.Errorf("invalid request payload"))
		return
	}
	err = p.createProduct(app.DB)
	if err != nil {
		sendError(w, http.StatusInternalServerError, err)
		return
	}
	sendResponse(w, http.StatusCreated, p)
}

func (app *App) updateProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key, err := strconv.Atoi(vars["id"])
	if err != nil {
		sendError(w, http.StatusBadRequest, fmt.Errorf("invalid product ID"))
		return
	}

	var p product
	err = json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		sendError(w, http.StatusBadRequest, fmt.Errorf("invalid request payload"))
		return
	}
	p.ID = key
	err = p.updateProduct(app.DB)
	if err != nil {
		sendError(w, http.StatusInternalServerError, err)
		return
	}
	sendResponse(w, http.StatusOK, p)
}

func (app *App) deleteProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key, err := strconv.Atoi(vars["id"])
	if err != nil {
		sendError(w, http.StatusBadRequest, fmt.Errorf("invalid product ID"))
		return
	}

	p := product{ID:key}
	err = p.deleteProduct(app.DB)
	if err != nil {
		sendError(w, http.StatusInternalServerError, err)
		return
	}
	sendResponse(w, http.StatusOK, map[string]string{"result": "success"})
}

func (app *App) handlerRoutes() {
	app.Router.HandleFunc("/products", app.getProducts).Methods("GET")
	app.Router.HandleFunc("/product/{id}", app.getProduct).Methods("GET")
	app.Router.HandleFunc("/product/", app.createProduct).Methods("POST")
	app.Router.HandleFunc("/product/{id}", app.updateProduct).Methods("PUT")
	app.Router.HandleFunc("/product/{id}", app.deleteProduct).Methods("DELETE")
}
