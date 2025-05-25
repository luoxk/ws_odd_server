package main

import (
	bbinWails "bbinWails/src"
	"fmt"
	"github.com/luoxk/ws_odd_server/handler"
	"github.com/luoxk/ws_odd_server/models"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func cleanup() {
	fmt.Println("执行清理逻辑...")
	bbinWails.GetBrowserController().CloseAllBrowsers()
	// 例如关闭数据库、保存状态、释放资源等
}

func main() {
	// 捕获信号
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// 启动 goroutine 等待信号
	go func() {
		<-sigs
		cleanup()
		os.Exit(0)
	}()

	instance := models.GetBotInstance()
	if instance != nil {
		if instance.LoginCheckIn() {

			instance.GetBrowser().Close()
		}
	}

	http.HandleFunc("/ws", handler.HandleWebSocket)
	http.HandleFunc("/debug/snapshot", handler.HandleDebugSnapshot)

	log.Println("WebSocket server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))

}
