package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
	recipe "go.etcd.io/etcd/client/v3/experimental/recipes"
)

// 计数型屏障（两阶段屏障）
// 两阶段屏障是一种分布式同步工具，它允许一组节点在屏障处等待，直到所有节点都到达屏障后，所有节点才能继续执行
// 两阶段屏障有两个阶段：进入和离开
// Enter 进入阶段：所有节点都到达屏障后，屏障才会打开，所有节点才能继续执行
// Leave 离开阶段：所有节点都离开屏障后，屏障才会关闭，所有节点才能继续执行

// 第一阶段：一群小学生去春游，等人来齐了大巴车开门
// 第二阶段：他们中午午休到餐馆吃饭，等人来齐了才能动筷子

// func NewDoubleBarrier(s *concurrency.Session, key string, count int) *DoubleBarrier
// 创建一个两阶段屏障，key 是屏障的名字，count 是屏障的大小

var (
	addr        = flag.String("addr", "http://127.0.0.1:2379", "etcd addresses")
	barrierName = flag.String("name", "my-test-doublebarrier", "barrier name")
	count       = flag.Int("c", 2, "")
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

	// 创建/获取 DoubleBarrier
	b := recipe.NewDoubleBarrier(s1, *barrierName, *count)

	// 从命令行读取命令
	consolescanner := bufio.NewScanner(os.Stdin)
	for consolescanner.Scan() {
		action := consolescanner.Text()
		items := strings.Split(action, " ")
		switch items[0] {
		case "enter": // 持有这个 DoubleBarrier
			b.Enter()
			fmt.Println("enter")
		case "leave": // 释放这个 DoubleBarrier
			b.Leave()
			fmt.Println("leave")
		case "quit", "exit": //退出
			return
		default:
			fmt.Println("unknown action")
		}
	}
}
