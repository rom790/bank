package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gorilla/mux"
)

type apiFunc func(http.ResponseWriter, *http.Request) error

type APIError struct {
	Error string `json:"error"`
}

type APIServer struct {
	listenAddr string
	store      Storage
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
func NewApiServ(listenAddr string, store Storage) *APIServer {
	return &APIServer{
		listenAddr: listenAddr,
		store:      store,
	}

}

func (s *APIServer) Run() {
	router := mux.NewRouter()

	router.HandleFunc("/account", makeHTTPHandleFunc(s.handleAccount))
	router.HandleFunc("/account/{id}", makeHTTPHandleFunc(s.handleAccByID))
	router.HandleFunc("/transfer", makeHTTPHandleFunc(s.handleTransfer))
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

func (s *APIServer) handleAccByID(w http.ResponseWriter, r *http.Request) error {
	switch r.Method {
	case "GET":
		return s.handleGetAccByID(w, r)
	case "DELETE":
		return s.handleDeleteAcc(w, r)
	}

	return fmt.Errorf("ussupported id method %s", r.Method)
}

func (s *APIServer) handleGetAcc(w http.ResponseWriter, r *http.Request) error {
	accounts, err := s.store.GetAccounts()
	if err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, accounts)
}
func (s *APIServer) handleGetAccByID(w http.ResponseWriter, r *http.Request) error {
	id, err := getIDFromRequest(r)
	if err != nil {
		return err
	}
	account, err := s.store.GetAccountByID(id)
	if err != nil {
		return err
	}
	return WriteJSON(w, http.StatusOK, account)
}
func (s *APIServer) handleCreateAcc(w http.ResponseWriter, r *http.Request) error {
	createAccReq := new(CreateAccountRequest)
	if err := json.NewDecoder(r.Body).Decode(createAccReq); err != nil {
		return err
	}

	defer r.Body.Close()
	account := NewAccount(createAccReq.FirstName, createAccReq.LastName)
	if err := s.store.CreateAccount(account); err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, createAccReq)
}
func (s *APIServer) handleDeleteAcc(w http.ResponseWriter, r *http.Request) error {
	id, err := getIDFromRequest(r)
	if err != nil {
		return err
	}

	if err := s.store.DeleteAccount(id); err != nil {
		return err
	}
	return WriteJSON(w, http.StatusOK, map[string]int{"deleted": id})
}
func (s *APIServer) handleTransfer(w http.ResponseWriter, r *http.Request) error {
	transferReq := new(TransferRequest)
	if err := json.NewDecoder(r.Body).Decode(transferReq); err != nil {
		return err
	}

	defer r.Body.Close()
	return WriteJSON(w, http.StatusOK, transferReq)
}

func getIDFromRequest(r *http.Request) (int, error) {
	idStr := mux.Vars(r)["id"]
	// fmt.Println(idStr)
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return id, fmt.Errorf("invalid id %s", idStr)
	}

	return id, nil
}
