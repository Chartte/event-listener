package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

// SendDingMsg - Deprecated
//func SendDingMsg(eventType, name, reason, message, time string) {
//	token = "https://oapi.dingtalk.com/robot/send?access_token=" + token
//	log.Printf("debug: dingtalk_url - %s", token)
//	// 检查环境变量是否存在
//	if project == "" || token == "" {
//		fmt.Println("Please set [project] and [token] environment variables")
//		os.Exit(1)
//	}
//
//	content := fmt.Sprintf("{\"msgtype\": \"markdown\",\"markdown\": {\"title\":\"集群Event告警\",\"text\":\"project:%v\n,type:%v\n,name:%v\n,reason:%v\n,message:%v\n,time:%v\"},\"at\":{\"isAtAll\":true}}", project, eventType, name, reason, message, time)
//
//	//创建一个请求
//	req, err := http.NewRequest("POST", token, strings.NewReader(content))
//	if err != nil {
//		log.Fatal("create request failed")
//	}
//
//	client := &http.Client{}
//	//设置请求头
//	req.Header.Set("Content-Type", "application/json; charset=utf-8")
//	//发送请求
//	resp, err := client.Do(req)
//	//关闭请求
//	defer resp.Body.Close()
//
//	if err != nil {
//		log.Fatal("Send request failed")
//	}
//}

// sendMessage - Send Message to Alert-Receiver
func sendMessage(alert AlertRequest) error {
	// 将请求结构体编码为 JSON
	jsonData, err := json.Marshal(alert)
	if err != nil {
		//log.Printf("JSON 编码失败: %v", err)
		return fmt.Errorf("JSON 编码失败: %v", err)
	}

	url := fmt.Sprintf("http://%s/alerts", alertSocket)
	// 发送 POST 请求
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		//log.Printf("发送报警请求失败: %v", err)
		return fmt.Errorf("发送报警请求失败: %v", err)
	}
	// 确保关闭响应体
	defer resp.Body.Close()

	// 读取响应体
	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		//log.Printf("读取响应体失败: %v", readErr)
		return fmt.Errorf("读取响应体失败: %v", readErr)
	}

	// 如果状态码不是 200，返回错误
	if resp.StatusCode != http.StatusOK {
		//log.Printf("报警请求失败，状态码: %d，响应内容: %s", resp.StatusCode, string(body))
		return fmt.Errorf("报警请求失败，状态码: %d，响应内容: %s", resp.StatusCode, string(body))
	}
	return nil
}

// sendUnhealthyMsg 发送不健康状态消息
func sendUnhealthyMsg(message string) {
	err := sendMessage(NewUnHealthyRequest(message))
	if err != nil {
		log.Println(err)
	}
}

// sendHealthyMsg 发送健康状态消息
func sendHealthyMsg() {
	err := sendMessage(NewHealthyRequest())
	if err != nil {
		log.Println(err)
	}
}

func sendPendingReviewMsg(message string) {
	err := sendMessage(NewPendingReviewRequest(message))
	if err != nil {
		log.Println(err)
	}
}
