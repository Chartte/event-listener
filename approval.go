package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
)

var (
	mutex = &sync.Mutex{} // 防止并发修改
)

// EventInfo 包含事件信息和审核状态
type EventInfo struct {
	Name       string `json:"name"`
	Reason     string `json:"reason"`
	Message    string `json:"message"`
	IsReviewed bool   `json:"isReviewed"`
}

// 启动 HTTP 服务器
func startHTTPServer() {
	// 提供静态文件服务（HTML 文件）
	//http.Handle("/", http.FileServer(http.Dir("./static")))
	http.Handle("/", http.FileServer(http.Dir("./")))

	http.HandleFunc("/events", handleListEvents)
	http.HandleFunc("/approve", handleEventReview)

	log.Println("HTTP server started on :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatalf("HTTP server failed to start: %v", err)
	}
}

// handleListEvents - 显示当前缓存的事件列表，并提供审核按钮
func handleListEvents(w http.ResponseWriter, r *http.Request) {
	mutex.Lock()
	defer mutex.Unlock()

	var eventList []EventInfo

	for _, entry := range eventCache {
		eventList = append(eventList, EventInfo{
			Name:       entry.Event.Name,
			Reason:     entry.Event.Reason,
			Message:    entry.Event.Message,
			IsReviewed: entry.IsReviewed,
		})
	}

	// 返回 JSON 数据
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(eventList)
}

// handleEventReview - 处理审核事件的 HTTP 请求
func handleEventReview(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}
	mutex.Lock()
	defer mutex.Unlock()

	eventName := r.URL.Query().Get("eventName")
	if eventName == "" {
		http.Error(w, "Missing eventName", http.StatusBadRequest)
		return
	}

	// 查找并标记事件为已审核
	if entry, exists := eventCache[eventName]; exists {
		entry.IsReviewed = true
		fmt.Fprintf(w, "Event %v has been reviewed.", eventName)
	} else {
		http.Error(w, "Event not found", http.StatusNotFound)
	}
}
