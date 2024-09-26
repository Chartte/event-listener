package main

import "time"

// AlertRequest 定义了发送给报警接收器的请求体
type AlertRequest struct {
	Category  string `json:"category"`
	Project   string `json:"project"`
	State     string `json:"state"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
}

// NewHealthyRequest - 创建健康请求
func NewHealthyRequest() AlertRequest {
	return AlertRequest{
		Category:  "k8s-event",
		Project:   project,
		Message:   "success",
		State:     "Healthy",
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
	}
}

// NewUnHealthyRequest - 创建非健康请求
func NewUnHealthyRequest(message string) AlertRequest {
	return AlertRequest{
		Category:  "k8s-event",
		Project:   project,
		Message:   message,
		State:     "UnHealthy",
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
	}
}

func NewPendingReviewRequest(message string) AlertRequest {
	return AlertRequest{
		Category:  "k8s-event",
		Project:   project,
		Message:   message,
		State:     "UnHealthy",
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
	}
}
