package main

import (
	"fmt"
	"time"
)

func main()  {
	c := make(chan int) //创建出用于通信的通道
	for i := 0; i < 5; i++ {
		go sleepyGopher(i, c)
	}
	for i := 0; i < 5; i++ {
		gopherID := <-c //从通道中接收值
		fmt.Println("gopher ", gopherID, " has finished sleeping")
	}
}

func sleepyGopher(id int, c chan int)  { //将通道声明为实参
	time.Sleep(3 * time.Second)
	fmt.Println("...", id, " snore ...")
	c <- id //将值回传至main函数
}