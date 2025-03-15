package browserEvent

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	ws "github.com/devpablocristo/monorepo/pkg/websocket/gorilla"
	dto "github.com/devpablocristo/monorepo/projects/qh/internal/browser-events/websocket/dto"
)

// webSocket es el adaptador WS que utiliza el paquete ws para manejar conexiones WebSocket.
type webSocket struct {
	ucs UseCases
	upg ws.Upgrader
}

// NewWebSocket crea una nueva instancia de webSocket inyectando el Upgrader.
func NewWebSocket(ucs UseCases, upg ws.Upgrader) WebSocket {
	return &webSocket{
		ucs: ucs,
		upg: upg,
	}
}

// websocketConnection actualiza la conexión HTTP a una conexión WebSocket usando ws y delega el manejo.
func (h *webSocket) websocketConnection(w http.ResponseWriter, r *http.Request, handler func(ws.Conn)) {
	// Usa el upgrader inyectado.
	ws.HandleWebSocket(w, r, h.upg, func(conn ws.Conn) {
		defer conn.Close()
		handler(conn)
	})
}

// Ping es un endpoint de ejemplo que responde con "Pong!" en una conexión WebSocket.
func (h *webSocket) Ping(w http.ResponseWriter, r *http.Request) {
	h.websocketConnection(w, r, func(conn ws.Conn) {
		for {
			messageType, message, err := conn.ReadMessage()
			if err != nil {
				log.Printf("Error reading message: %v", err)
				break
			}

			log.Printf("Received message (Ping): %s", string(message))
			response := "Pong!"
			if err := conn.WriteMessage(messageType, []byte(response)); err != nil {
				log.Printf("Error writing message: %v", err)
				break
			}
		}
	})
}

// BrowserEvent procesa eventos en tiempo real desde el navegador.
func (h *webSocket) BrowserEvent(w http.ResponseWriter, r *http.Request) {
	h.websocketConnection(w, r, func(conn ws.Conn) {
		ctx := context.Background()
		for {
			messageType, message, err := conn.ReadMessage()
			if err != nil {
				log.Printf("Error reading message: %v", err)
				break
			}

			var event dto.BrowserEvent
			if err := json.Unmarshal(message, &event); err != nil {
				log.Printf("Error unmarshalling message to DTO: %v", err)
				sendError(conn, messageType, "Invalid event format")
				continue
			}

			if err := h.ucs.BrowserEvent(ctx, event.ToDomain()); err != nil {
				log.Printf("Error processing event: %v", err)
				sendError(conn, messageType, "Error processing event: "+err.Error())
				continue
			}

			ackMessage := []byte(`{"status": "Event processed successfully"}`)
			if err := conn.WriteMessage(messageType, ackMessage); err != nil {
				log.Printf("Error sending acknowledgement: %v", err)
				break
			}
		}
	})
}

// sendError es una función auxiliar para enviar un mensaje de error al cliente.
func sendError(conn ws.Conn, messageType int, errorMsg string) {
	errMsg := map[string]string{"error": errorMsg}
	if msg, err := json.Marshal(errMsg); err == nil {
		if err := conn.WriteMessage(messageType, msg); err != nil {
			log.Printf("Error sending error message: %v", err)
		}
	}
}
