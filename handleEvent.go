package main

import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"log"
	"time"
)

// 处理 Added 事件
func handleEventAdded(event *corev1.Event) {
	//eventType := event.Object.(*corev1.Event).Type

	// 缓存事件，后续用来跟踪状态
	//eventCache[event.Name] = event

	eventTime := event.CreationTimestamp
	currentTime := time.Now()

	// 判断事件中的时间是否与实际时间一致
	if eventTime.Year() == currentTime.Year() &&
		eventTime.Month() == currentTime.Month() &&
		eventTime.Day() == currentTime.Day() &&
		eventTime.Hour() == currentTime.Hour() &&
		eventTime.Minute() == currentTime.Minute() {

		// 如果不存在Cache中 则新增至内map中
		if _, exists := eventCache[event.Name]; !exists {
			eventCache[event.Name] = &EventCache{event, false}
		}

		//fmt.Printf("Type: %#v\n", eventType)
		content := fmt.Sprintf("name:%v, reason:%v, message:%v\n", event.Name, event.Reason, event.Message)
		log.Println("检测到 Added Warning 事件:", content)

		// 发送不健康消息
		sendUnhealthyMsg(content)
	}
}

// checkEventResolve - 通过 Kubernetes API 查询 Warning Event是否解决
func checkEventResolve(clientset *kubernetes.Clientset, namespace string) bool {
	// 查询当前的事件列表
	events, _ := clientset.CoreV1().Events(namespace).List(context.Background(), v1.ListOptions{})

	// 创建一个 map 记录当前存在的 Warning 事件
	currentWarningEvents := make(map[string]bool)
	hasUnresolvedWarnings := false

	// 遍历事件列表，筛选出 Warning 类型事件
	for _, event := range events.Items {
		if event.Type == "Warning" {
			//log.Printf("DEBUG: %v", event.Name)
			currentWarningEvents[event.Name] = true

			// 检查事件是否在缓存中
			if cachedEvent, exists := eventCache[event.Name]; exists {
				// 检查 event.Count 是否大于缓存中的值
				if event.Count > cachedEvent.Event.Count {
					// 事件的 count 增加，说明事件未解决且在重试，发送不健康消息
					content := fmt.Sprintf("Warning 事件未解决，name:%v, reason:%v, message:%v\n", event.Name, event.Reason, event.Message)
					log.Println(content)
					sendUnhealthyMsg(content)

					// 更新缓存中的事件
					eventCache[event.Name].Event = event.DeepCopy()
					hasUnresolvedWarnings = true
				} else {
					// 重试值不再增加，记录下来等待下一次检测
					content := fmt.Sprintf("事件 %v 的重试值没有增加，正在等待人工审核是否解决..\n", event.Name)
					//log.Println(content)
					// 判断事件是否已经被审核
					if !eventCache[event.Name].IsReviewed {
						// 事件尚未被审核，发送等待审核的告警消息
						log.Printf("事件%v 重试次数不再增加,需要人工审核 ", event.Name)
						sendPendingReviewMsg(content)
						hasUnresolvedWarnings = true
					} else {
						// 事件已审核，不再发送告警，记录日志
						log.Printf("事件 %v 已审核，跳过告警，认为该告警已无威胁, 等待周期结束\n", event.Name)
					}
				}
			} else {
				// 新的事件，加入缓存
				eventCache[event.Name] = &EventCache{
					Event:      event.DeepCopy(),
					IsReviewed: false, // 默认设置为未审核
				}
				log.Printf("遗漏的 Warning 事件加入缓存: %v\n", event.Name)
				hasUnresolvedWarnings = true // 只要有新的 Warning 事件，视为未解决
			}
		}
	}

	// 遍历缓存中的事件，移除 Kubernetes 中已消失的事件
	for cachedEventName := range eventCache {
		if _, exists := currentWarningEvents[cachedEventName]; !exists {
			// 缓存中的事件已经从 Kubernetes 中消失，说明已经解决
			log.Printf("事件 %v 已解决，移除缓存\n", cachedEventName)
			delete(eventCache, cachedEventName)
		}
	}

	if len(eventCache) > 0 {
		log.Println("当前缓存中仍存在事件列表:")
		i := 1 // 初始化事件编号
		log.Println("--------------------------------------------------------")
		for eventName := range eventCache {
			log.Printf("[%d] 事件名: %v", i, eventName)
			i++ // 编号递增
		}
		log.Println("--------------------------------------------------------")
	}

	// 如果没有未解决的 Warning 事件，返回 true 表示所有事件已解决
	return !hasUnresolvedWarnings
}

// checkAllEventsHealthy - 检查所有事件是否健康，发送健康消息
func checkAllEventsHealthy(clientset *kubernetes.Clientset, namespace string) {
	// 检查事件是否都已解决
	if allResolved := checkEventResolve(clientset, namespace); allResolved {
		// 只有当所有 Warning 事件都解决时，发送健康消息
		log.Println("所有 Warning 事件已解决，发送健康消息")
		sendHealthyMsg()
	}
}

// 处理 Modified 事件
//func handleEventModified(event *corev1.Event) {
//	// 判断事件是否在缓存中
//	if cachedEvent, exists := eventCache[event.Name]; exists {
//		// 如果事件依然存在，更新缓存并继续告警
//		cachedEvent = event
//		eventCache[event.Name] = cachedEvent
//
//		// 持续告警
//		content := fmt.Sprintf("Warning 事件未解决，name:%v, reason:%v, message:%v\n", event.Name, event.Reason, event.Message)
//		log.Println("持续检测到 Warning 事件:", content)
//		sendUnhealthyMsg(content)
//	}
//}

//
//// 处理 Deleted 事件
//func handleEventDeleted(event *corev1.Event) {
//	// 从缓存中删除事件，表示告警解除
//	if _, exists := eventCache[event.Name]; exists {
//		delete(eventCache, event.Name)
//		log.Println("Warning 事件已解决:", event.Name)
//		//sendHealthyMsg()
//	}
//}
//
//// 定期检查缓存中的事件状态，如果存在则继续发送告警
//func checkEventStatus() {
//	for _, event := range eventCache {
//		content := fmt.Sprintf("持续监控 Warning 事件，name:%v, reason:%v, message:%v\n", event.Name, event.Reason, event.Message)
//		log.Println("监视仪：正在监测未处理 Warning 事件:", content)
//		sendUnhealthyMsg(content)
//	}
//}
