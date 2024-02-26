package main

import (
	"context"
	"flag"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
)

// 提供查询 key 信息的功能和支持 ctx 撤销的互斥锁

var (
	addr     = flag.String("addr", "http://127.0.0.1:2379", "etcd addresses")
	lockName = flag.String("name", "my-test-lock", "lock name")
	crash    = flag.Bool("crash", false, "crash after acquiring lock")
)

func main() {
	flag.Parse()

	rand.Seed(time.Now().UnixNano())

	// etcd地址
	endpoints := strings.Split(*addr, ",")
	cli, err := clientv3.New(clientv3.Config{Endpoints: endpoints})
	if err != nil {
		log.Fatal(err)
	}
	defer cli.Close()

	useMutex(cli)
}

func useMutex(cli *clientv3.Client) {
	// 为锁生成 session，30s 后自动过期释放锁
	s1, err := concurrency.NewSession(cli, concurrency.WithTTL(30))
	if err != nil {
		log.Fatal(err)
	}
	defer s1.Close()
	// mutex 有名字，区分不同的 mutex
	m1 := concurrency.NewMutex(s1, *lockName)

	// 在请求锁之前查询 key
	log.Printf("before acquiring. key: %s", m1.Key())
	// 请求锁
	log.Println("acquiring lock")
	if err := m1.Lock(context.TODO()); err != nil {
		log.Fatal(err)
	}
	log.Printf("acquired lock. key: %s", m1.Key())

	time.Sleep(time.Duration(rand.Intn(30)) * time.Second)
	if *crash { // 如果节点崩溃，程序直接退出，虽然还持有锁
		log.Println("crashing")
		os.Exit(1)
	}

	if err := m1.Unlock(context.TODO()); err != nil {
		log.Fatal(err)
	}
	log.Println("released lock")
}
