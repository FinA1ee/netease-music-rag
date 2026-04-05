package service

import (
	"fmt"
	"time"

	"github.com/skip2/go-qrcode"
)

func ShowQRCodeInTerminal(content string) {
	// 生成 ASCII 二维码
	qr, err := qrcode.New(content, qrcode.Medium)
	if err != nil {
		fmt.Println("生成二维码失败:", err)
		return
	}

	// 打印到控制台（黑白块，手机可扫）
	fmt.Println("\n========== 请用网易云APP扫码 ==========")
	fmt.Print(qr.ToSmallString(false))
	fmt.Println("\n=======================================")
}

// Poll 轮询检查
// interval: 轮询间隔 秒
// timeout: 超时时间 秒
// checkFunc: 你的检查逻辑，返回 true 表示成功，false 继续轮询
func Poll(interval int, timeout int, checkFunc func() bool) bool {
	// 计时器
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()

	// 超时计时器
	timeoutChan := time.After(time.Duration(timeout) * time.Second)

	fmt.Printf("开始轮询，间隔 %d秒，超时时间 %d秒\n", interval, timeout)

	// 循环监听
	for {
		select {
		case <-timeoutChan:
			fmt.Println("轮询超时！")
			return false

		case <-ticker.C:
			// 执行检查
			fmt.Println("执行检查...")
			if checkFunc() {
				fmt.Println("检查成功！退出轮询")
				return true
			}
		}
	}
}
