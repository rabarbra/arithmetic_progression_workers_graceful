package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"

	"workers_server/workerstore"
)

var ws *workerstore.WorkerStore
var v *validator.Validate
var shutdownTimeout = time.Second * 5

func main() {

	var maxParallelCount uint
	flag.UintVar(&maxParallelCount, "n", 2, "Max")
	flag.Parse()
	if maxParallelCount == 0 {
		log.Fatal("-n flag must be equal or grater then 1")
	}

	ws = workerstore.NewWorkerStore()
	v = validator.New()

	cancelCtx, cancel_workers := context.WithCancel(context.Background())
	go ws.StartWorkers(cancelCtx, maxParallelCount)

	router := mux.NewRouter()
	router.HandleFunc("/add", NewTask).Methods("POST")
	router.HandleFunc("/get", GetAllTasks).Methods("GET")

	srv := &http.Server{
		Handler:      router,
		Addr:         "127.0.0.1:8000",
		WriteTimeout: 5 * time.Second,
		ReadTimeout:  5 * time.Second,
	}
	log.Printf("Starting server on %s", srv.Addr)
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Error in ListenAndServe(): %v", err)
		}
	}()
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	log.Println("Shuting down")
	cancel_workers()
	time.Sleep(shutdownTimeout)
	timeoutCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(timeoutCtx); err != nil {
		log.Fatalf("Server Shutdown error: %v", err)
	}
	log.Println("Server shut down succesfully")
}

func NewTask(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var task workerstore.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		log.Println(err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	if err := v.Struct(task); err != nil {
		log.Println(err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	resp := ws.AddTask(task)
	b, err := json.Marshal(resp)
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
	w.Write(b)
}

func GetAllTasks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	workers := ws.GetSortedTasks()
	b, err := json.Marshal(workers)
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
	w.Write(b)
}
