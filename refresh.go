package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"shuaitesteel.com/cms/config"
	"shuaitesteel.com/cms/handlers"
)

func refreshSignalPath() string {
	return filepath.Join(config.BaseDir(), ".cms-refresh")
}

func requestRefresh() {
	path := refreshSignalPath()
	content := []byte(time.Now().Format(time.RFC3339Nano))
	if err := os.WriteFile(path, content, 0644); err != nil {
		fail("发送刷新请求失败: %v", err)
	}
	fmt.Printf("已发送刷新请求: %s\n", path)
}

func startRefreshWatcher() {
	path := refreshSignalPath()
	go func() {
		var lastValue string
		for {
			content, err := os.ReadFile(path)
			if err == nil {
				value := string(content)
				if value != "" && value != lastValue {
					lastValue = value
					log.Printf("收到静态刷新请求: %s", path)
					handlers.InvalidateCache()
					handlers.RefreshStatic()
				}
			} else if !os.IsNotExist(err) {
				log.Printf("读取刷新请求失败: %v", err)
			}
			time.Sleep(time.Second)
		}
	}()
}
