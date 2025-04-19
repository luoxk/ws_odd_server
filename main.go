package main

import (
	bbinWails "bbinWails/src"
	"log"
	"net/http"
	"ws_odd_server/handler"
	"ws_odd_server/models"
)

func main() {
	defer func() {
		bbinWails.GetBrowserController().CloseAllBrowsers()
	}()
	instance := models.GetBotInstance()
	instance.LoginCheckIn()

	http.HandleFunc("/ws", handler.HandleWebSocket)
	http.HandleFunc("/debug/snapshot", handler.HandleDebugSnapshot)

	log.Println("WebSocket server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))

}
