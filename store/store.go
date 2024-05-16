// filePath:  store/store.go
package store

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"satmine/satmine"

	"github.com/gorilla/websocket"
)

// Store is a struct for storing global data and managing WebSocket connections
type Store struct {
	OrdIdx  *satmine.BTOrdIdx        // Index for orders
	RecIdx  *satmine.BTRecIdx        // Index for records
	clients map[*websocket.Conn]bool // Active WebSocket connections

	// Upgrader for WebSocket connections, allows for custom configurations
	upgrader websocket.Upgrader
}

// instance is the private instance of Store
var instance *Store

// once ensures the singleton instance is created only once
var once sync.Once

// Instance returns the single instance of Store
func Instance() *Store {
	once.Do(func() {
		instance = &Store{
			clients: make(map[*websocket.Conn]bool),
			upgrader: websocket.Upgrader{
				ReadBufferSize:  1024,
				WriteBufferSize: 1024,
				CheckOrigin: func(r *http.Request) bool {
					return true // Allow all CORS requests, be stricter in production environments
				},
			},
		}
	})
	return instance
}

// Init initializes the data of the Store and starts the WebSocket server on the specified port
func (s *Store) Init(OrdIdx *satmine.BTOrdIdx, RecIdx *satmine.BTRecIdx, port string) {
	s.OrdIdx = OrdIdx
	s.RecIdx = RecIdx
	go s.startWebSocketServer(port) // Start the WebSocket server in a new goroutine
	fmt.Println("Listening Socket Port =", port)
}

// AllSocketPush pushes a string message to all connected WebSocket clients
func (s *Store) AllSocketPush(message string) {
	msg := []byte(message)
	for client := range s.clients {
		if err := client.WriteMessage(websocket.TextMessage, msg); err != nil {
			log.Printf("WebSocket error: %v", err)
			client.Close()
			delete(s.clients, client)
		}
	}
}

// startWebSocketServer starts the WebSocket server on a specified port
func (s *Store) startWebSocketServer(port string) {
	http.HandleFunc("/ws", s.handleConnections)
	log.Printf("WebSocket server starting on :%s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("ListenAndServe: %v", err)
	}
}

// handleConnections upgrades HTTP to WebSocket and manages connection lifecycle
func (s *Store) handleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer ws.Close()
	s.clients[ws] = true // Register new client

	for {
		_, _, err := ws.ReadMessage()
		if err != nil {
			log.Printf("WebSocket read error: %v", err)
			delete(s.clients, ws)
			break
		}
	}
}
