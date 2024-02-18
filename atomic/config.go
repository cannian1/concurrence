package atomic

import (
	"fmt"
	"math/rand/v2"
	"sync/atomic"
	"time"
)

type Config struct {
	NodeName string
	Addr     string
	Count    int32
}

func loadNewConfig() Config {
	return Config{
		NodeName: "北京",
		Addr:     "10.77.95.27",
		Count:    rand.Int32(),
	}
}

func Demo() {
	// atomic.Value 原子地存取对象
	var config atomic.Value
	config.Store(loadNewConfig()) // 保存一个新的配置

	// 设置新的配置值
	go func() {
		for {
			time.Sleep(time.Duration(1+rand.Int64N(5)) * time.Second)
			config.Store(loadNewConfig())
		}
	}()

	go func() {
		for {
			cfg := config.Load().(Config)
			fmt.Printf("新的配置:%+v\n", cfg)
			time.Sleep(6 * time.Second)
		}
	}()

	select {}
}
