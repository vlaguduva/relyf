// app.go

package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	// _ IS BLANK IDENTIFIER IMPORTED SOLELY FOR ITS SIDE EFFECTS THAT IS THE INITIALIZATION TO HAPPEN
	_ "github.com/lib/pq"
)

type App struct {
	Router *mux.Router
	DB     *sql.DB
}

func (a *App) Initialize(host, user, password, dbname string) {
	connectionString := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable", host, user, password, dbname)
	log.Printf("Connection String: %s", connectionString)
	var err error

	a.DB, err = sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatal(err)
	}

	a.Router = mux.NewRouter()

	a.initializeRoutes()
}
func (a *App) Run(address string) {
	log.Fatal(http.ListenAndServe(":8080", a.Router))
}

func (a *App) initializeRoutes() {
	a.Router.HandleFunc("/products", a.getProducts).Methods("GET")
	a.Router.HandleFunc("/product/{id:[0-9]{1,3}}", a.getProduct).Methods("GET")
	a.Router.HandleFunc("/product/{id:[0-9]{1,3}}", a.updateProduct).Methods("PUT")
	a.Router.HandleFunc("/product/{id:[0-9]{1,3}}", a.deleteProduct).Methods("DELETE")
	a.Router.HandleFunc("/product", a.createProduct).Methods("POST")
}

func (a *App) getProducts(writer http.ResponseWriter, request *http.Request) {
	count, _ := strconv.Atoi(request.FormValue("count"))
	start, _ := strconv.Atoi(request.FormValue("start"))

	if count > 10 || count < 1 {
		count = 10
	}

	if start < 0 {
		start = 0
	}

	products, err := getProducts(a.DB, start, count)
	if err != nil {
		respondWithError(writer, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(writer, http.StatusOK, products)
}

func (a *App) getProduct(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	id, err := strconv.Atoi(vars["id"])

	if err != nil {
		respondWithError(writer, http.StatusBadRequest, "Invalid product id.")
		return
	}
	product := product{ID: id}

	err = product.getProduct(a.DB)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			respondWithError(writer, http.StatusNotFound, "Product not found.")
		default:
			respondWithError(writer, http.StatusInternalServerError, err.Error())
		}
		return
	}

	respondWithJSON(writer, http.StatusOK, product)
}

func (a *App) updateProduct(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	id, err := strconv.Atoi(vars["id"])

	if err != nil {
		respondWithError(writer, http.StatusBadRequest, "Invalid product id.")
		return
	}

	var p product
	decoder := json.NewDecoder(request.Body)
	if err := decoder.Decode(&p); err != nil {
		respondWithError(writer, http.StatusUnprocessableEntity, "Invalid request body.")
		return
	}

	p.ID = id

	err = p.updateProduct(a.DB)
	if err != nil {
		respondWithError(writer, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(writer, http.StatusOK, p)
}

func (a *App) createProduct(writer http.ResponseWriter, request *http.Request) {
	var p product
	decoder := json.NewDecoder(request.Body)
	if err := decoder.Decode(&p); err != nil {
		respondWithError(writer, http.StatusUnprocessableEntity, "Invalid request body.")
		return
	}

	if err := p.createProduct(a.DB); err != nil {
		respondWithError(writer, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(writer, http.StatusOK, p)
}

func (a *App) deleteProduct(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	id, err := strconv.Atoi(vars["id"])

	if err != nil {
		respondWithError(writer, http.StatusBadRequest, "Invalid product id.")
		return
	}
	product := product{ID: id}

	err = product.getProduct(a.DB)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			respondWithError(writer, http.StatusNotFound, "Product not found.")
		default:
			respondWithError(writer, http.StatusInternalServerError, err.Error())
		}
		return
	}

	respondWithJSON(writer, http.StatusOK, product)
}

func respondWithError(writer http.ResponseWriter, code int, message string) {
	respondWithJSON(writer, code, map[string]string{"error": message})
}

func respondWithJSON(writer http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(code)
	writer.Write(response)
}
