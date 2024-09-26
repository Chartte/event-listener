package main

import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const DEFAULT_NAMESPACE = "default"

// EventCacheEntry 用于存储事件及其审核状态
type EventCache struct {
	Event      *corev1.Event
	IsReviewed bool
}

var (
	project     string
	alertSocket string
	eventCache  = make(map[string]*EventCache) // 存储正在监控的事件及其 retry count
)

func init() {
	// 在 init 函数中初始化全局变量
	project = os.Getenv("PROJECT")
	alertSocket = os.Getenv("ALERT_SOCKET")
}

func main() {
	// 从集群获取config
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}

	// 创建 Kubernetes 客户端
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	fmt.Println("成功获取clientset")

	// 启动 HTTP 服务
	go startHTTPServer()

	// 创建信号处理程序，以便在收到终止信号时停止事件监听
	stopCh := make(chan os.Signal, 1)
	signal.Notify(stopCh, syscall.SIGINT, syscall.SIGTERM)

	// 创建一个定时器，每隔一分钟检测一次
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	// 启动事件监听循环
	for {
		err := watchEvents(clientset, stopCh, ticker)
		if err != nil {
			fmt.Printf("Error watching events: %v. Reconnecting...\n", err)
			time.Sleep(5 * time.Second) // 等待 5 秒后重连
		}
	}
}

func watchEvents(clientset *kubernetes.Clientset, stopCh <-chan os.Signal, ticker *time.Ticker) error {
	// 创建事件监听器
	watcher, err := clientset.CoreV1().Events(DEFAULT_NAMESPACE).Watch(context.Background(), v1.ListOptions{})
	if err != nil {
		return err
	}
	defer watcher.Stop()

	for {
		select {
		case event, ok := <-watcher.ResultChan():
			if !ok {
				fmt.Println("Event watcher channel closed, reconnecting...")
				return fmt.Errorf("watcher channel closed")
			}

			eventObj := event.Object.(*corev1.Event)

			// 只处理 Warning 类型的事件，其他类型忽略
			if eventObj.Type != "Warning" {
				continue
			}

			// 处理事件
			switch event.Type {
			case watch.Added:
				handleEventAdded(eventObj)
			}

		case <-ticker.C:
			// 每分钟查询未处理的 Warning 事件
			checkAllEventsHealthy(clientset, DEFAULT_NAMESPACE)

		case <-stopCh:
			// 收到终止信号，停止事件监听
			fmt.Println("Received termination signal. Stopping event watcher.")
			return nil
		}
	}
}
