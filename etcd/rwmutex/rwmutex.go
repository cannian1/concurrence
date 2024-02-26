package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
	recipe "go.etcd.io/etcd/client/v3/experimental/recipes"
)

// 当写锁被持有时，对读锁和写锁的请求都会被阻塞等待写锁的释放
// 当读锁被持有时，对写锁的请求会被阻塞等待读锁的释放，对读锁的请求可以直接获得锁

// 当读锁被持有时，如果有一个节点请求写锁，那么后续的读锁请求会被阻塞等待写锁的释放;
// 如果此时再有对读锁的请求，那么这个请求会被阻塞，直到前面的写锁被释放
// 这个阻塞行为与标准库 sync.RWMutex 的行为一致

var (
	addr     = flag.String("addr", "http://127.0.0.1:2379", "etcd addresses")
	lockName = flag.String("name", "my-test-lock", "lock name")
	action   = flag.String("rw", "w", "r means acquiring read lock, w means acquiring write lock")
)

func main() {
	flag.Parse()

	// 解析etcd地址
	endpoints := strings.Split(*addr, ",")

	// 创建etcd的client
	cli, err := clientv3.New(clientv3.Config{Endpoints: endpoints})
	if err != nil {
		log.Fatal(err)
	}
	defer cli.Close()
	// 创建session
	s1, err := concurrency.NewSession(cli)
	if err != nil {
		log.Fatal(err)
	}
	defer s1.Close()
	m1 := recipe.NewRWMutex(s1, *lockName)

	// 从命令行读取命令
	consolescanner := bufio.NewScanner(os.Stdin)
	for consolescanner.Scan() {
		action := consolescanner.Text()
		switch action {
		case "w": // 请求写锁
			testWriteLocker(m1)
		case "r": // 请求读锁
			testReadLocker(m1)
		default:
			fmt.Println("unknown action")
		}
	}
}

func testWriteLocker(m1 *recipe.RWMutex) {
	// 请求写锁
	log.Println("acquiring write lock")
	if err := m1.Lock(); err != nil {
		log.Fatal(err)
	}
	log.Println("acquired write lock")

	// 等待一段时间
	time.Sleep(30 * time.Second)

	// 释放写锁
	if err := m1.Unlock(); err != nil {
		log.Fatal(err)
	}
	log.Println("released write lock")
}

func testReadLocker(m1 *recipe.RWMutex) {
	// 请求读锁
	log.Println("acquiring read lock")
	if err := m1.RLock(); err != nil {
		log.Fatal(err)
	}
	log.Println("acquired read lock")

	// 等待一段时间
	time.Sleep(30 * time.Second)

	// 释放写锁
	if err := m1.RUnlock(); err != nil {
		log.Fatal(err)
	}
	log.Println("released read lock")
}
