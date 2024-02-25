package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
)

var (
	nodeID    = flag.Int("id", 0, "node ID")
	addr      = flag.String("addr", "http://127.0.0.1:2379", "etcd addresses")
	electName = flag.String("name", "my-test-elect", "election name")
)

func main() {
	flag.Parse()

	// etcd 的地址
	endpoints := strings.Split(*addr, ",")

	// 创建一个 etcd 客户端
	cli, err := clientv3.New(clientv3.Config{Endpoints: endpoints})
	if err != nil {
		log.Fatal(err)
	}
	defer cli.Close()

	// 创建一个并发 session
	session, err := concurrency.NewSession(cli)
	defer session.Close()

	// 得到选举同步原语
	e1 := concurrency.NewElection(session, *electName)

	consolescanner := bufio.NewScanner(os.Stdin)
	for consolescanner.Scan() {
		action := consolescanner.Text()
		switch action {
		case "elect": // 启动选举
			go elect(e1, *electName)
		case "proclaim": // 宣告，设置主节点的值
			proclaim(e1, *electName)
		case "resign": // 放弃主节点
			resign(e1, *electName)
		case "watch": // 监听主从变化事件
			go watch(e1, *electName)
		case "query": // 主动查询
			query(e1, *electName)
		case "rev": // 查询版本号
			rev(e1, *electName)
		default:
			fmt.Println("unknown action")
		}
	}
}

var count int

// 选主 主节点的值为 value-<主节点 ID>-<count>
func elect(e1 *concurrency.Election, electName string) {
	log.Println("acampaigning for ID:", *nodeID)
	// 把一个节点设为主节点，并且设置一个值
	// 这是一个阻塞方法，在调用它时会被阻塞，直到满足下面三个条件之一：
	// 1. 当前节点成为主节点
	// 2. 此方法返回错误
	// 3. ctx 被取消
	if err := e1.Campaign(context.Background(), fmt.Sprintf("value-%d-%d", *nodeID, count)); err != nil {
		log.Println(err)
	}
	log.Println("campaigned for ID:", *nodeID)
	count++
}

// 为主节点设置一个新值
func proclaim(e1 *concurrency.Election, electName string) {
	log.Println("proclaiming for ID:", *nodeID)
	// 重新设置主节点的值，但是不会重新选主
	if err := e1.Proclaim(context.Background(), fmt.Sprintf("value-%d-%d", *nodeID, count)); err != nil {
		log.Println(err)
	}
	log.Println("proclaimed for ID:", *nodeID)
	count++
}

// 重新选主，有可能另一个节点成为主节点
func resign(e1 *concurrency.Election, electName string) {
	log.Println("resigning for ID:", *nodeID)
	// 主节点主动放弃主节点身份，开始新一轮选举
	if err := e1.Resign(context.TODO()); err != nil {
		log.Println(err)
	}
	log.Println("resigned for ID:", *nodeID)
}

func watch(e1 *concurrency.Election, electName string) {
	// 通过 Observe 方法可以观察主节点的变化
	// 返回一个 channel，显示主节点的变化，只会返回最近一条变化以及之后的变化
	ch := e1.Observe(context.TODO())

	log.Println("start to watch for ID:", *nodeID)
	for i := 0; i < 10; i++ {
		resp := <-ch
		log.Println("leader changed to", string(resp.Kvs[0].Key), string(resp.Kvs[0].Value))
	}
}

// 查询当前主节点
func query(e1 *concurrency.Election, electName string) {
	// 调用 Leader 方法可以获取当前主节点的信息，包括主节点的键值对
	resp, err := e1.Leader(context.Background())
	if err != nil {
		log.Printf("failed to get the current leader: %v", err)
	}
	log.Println("current leader:", string(resp.Kvs[0].Key), string(resp.Kvs[0].Value))
}

// 直接查询主节点的版本号
func rev(e1 *concurrency.Election, electName string) {
	rev := e1.Rev()
	log.Println("current rev:", rev)
}
