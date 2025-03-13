package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
)

type apiFunc func(http.ResponseWriter, *http.Request) error

type APIError struct {
	Error string
}

type APIServer struct {
	listenAddr string
}

func WriteJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

func makeHTTPHandleFunc(f apiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			WriteJSON(w, http.StatusBadRequest, APIError{Error: err.Error()})
		}
	}
}
func NewApiServ(listenAddr string) *APIServer {
	return &APIServer{
		listenAddr: listenAddr,
	}

}

func (s *APIServer) Run() {
	router := mux.NewRouter()

	router.HandleFunc("/account", makeHTTPHandleFunc(s.handleAccount))

	// http.ListenAndServe(s.listenAddr, router)
	server := &http.Server{
		Addr:    s.listenAddr,
		Handler: router,
	}

	// Запускаем сервер в отдельной горутине
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	log.Printf("Server started on %s", s.listenAddr)

	// Ожидаем сигналов для Graceful Shutdown
	waitForShutdown(server)

}

func waitForShutdown(server *http.Server) {
	// Канал для получения сигналов ОС
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	// Ожидаем сигнала
	<-stop

	// Создаем контекст с таймаутом для завершения работы
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	log.Println("Shutting down server...")

	// Останавливаем сервер
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown error: %v", err)
	}

	log.Println("Server gracefully stopped")
}

func (s *APIServer) handleAccount(w http.ResponseWriter, r *http.Request) error {
	switch r.Method {
	case "GET":
		return s.handleGetAcc(w, r)
	case "POST":
		return s.handleCreateAcc(w, r)
	case "DELETE":
		return s.handleDeleteAcc(w, r)
	}

	return fmt.Errorf("unsupported method %s", r.Method)

}
func (s *APIServer) handleGetAcc(w http.ResponseWriter, r *http.Request) error {
	account := NewAccount("Maks", "Kidrov")
	return WriteJSON(w, http.StatusOK, account)
}
func (s *APIServer) handleCreateAcc(w http.ResponseWriter, r *http.Request) error {
	return nil
}
func (s *APIServer) handleDeleteAcc(w http.ResponseWriter, r *http.Request) error {
	return nil
}
func (s *APIServer) handleTransfer(w http.ResponseWriter, r *http.Request) error {
	return nil
}
