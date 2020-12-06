# 第30章-goroutine和并发

## 启动 goroutine

启动 goroutine 就像调用函数一样简单，唯一要做的就是在调用前面写下一个关键词 go。

[代码 30-1](../code/chap30/30-1/sleepygopher.go)

## 不止一个 goroutine

每次使用关键字 go，都会产生一个新的 goroutine。

[代码 30-2](../code/chap30/30-2/sleepygophers.go)

[代码 30-3](../code/chap30/30-3/identifiedgophers.go)

## 通道

通道（channel）可以在多个 goroutine 之间安全的传递值。
      
跟 Go 中的其他类型一样，你可以将通道用作变量、传递至函数、存储在结构中，或者做你想让它做的几乎任何事情。


### 创建
使用内置函数 make 函数创建通道，并且为其指定相应的类型：

```go
c := make(chan int)
```

有了通道之后，通过左箭头操作符（<-）向它发送值或者从它那里接收值。

> 在向通道发送值时，将通道表达式放在左箭头操作符的左边，而待发送的值则放在左箭头操作符的右边。


### 发送
发送操作会等待直到有另一个 goroutine 尝试对相同的通道执行接收操作为止。

执行发送操作的 goroutine 在等待期间将无法执行其他操作，但是其他未在等待通道操作的 goroutine 仍然可以继续自由的运行。

比如，将值 99 发送至通道c：
```go
c <- 99
```

### 接收

在通过通道接收值的时候，将左箭头操作符放在通道的左边，让箭头指向通道之外的地方。

比如，从通道 c 中接收了一个值，并将它赋值给变量 r：
```go
r := <-c
```

[代码 30-4](../code/chap30/30-4/simplechan.go)

跟执行发送操作时一样，执行接收操作的 goroutine 将等待直到有另一个 goroutine 尝试向相同的通道执行发送操作为止。

## 使用 select 处理多个通道

有些情况下，当我们等待通道中的某些值时，可能并不愿意等待太久。比如，想要在网络请求发生数秒之后将判断其是否超时。

Go 标准库中提供了 time.After 函数，这个函数会返回一个通道，该通道会在经过待定时间之后接收到一个值（发送该值的 goroutine 是 Go 运行时的其中一部分）。


代码 30-5 select1.go
```go
func main(){
    timeout := time.After(2 * time.Second)
    for i := 0; i < 5; i++ {
        select { //select 语句
            case gopherID := <- c: //等待接收 goroutine 那里接收值
                fmt.Println("gopher ", gopherID, " has finished sleeping")
            case <- timeout: // 等待直到时间耗尽
                fmt.Println("my patinece ran out")
                return //放弃等待然后返回
        }
    }
}
```
> select 语句在不包含任何分支的情况下将永远地等待下去。
>
> 当启动多个 goroutine 并且打算让它们无线期地运行下去，就可以用这个方法阻止 main 函数返回。



代码 30-6 select2.go
```go
func sleepyGopher(id int, c chan int) {
    time.Sleep(time.Duration(rand.Intn(4000)) * time.Millisecond)
    c <- id
}
```
> 这个模式适用于任何想要控制事件完成时间的场景。通过将动作放入 goroutine 并在动作完成时向通道执行发送操作，我们可以为 Go 中的任何动作都设置超时。

**注意：**
> 即使程序已经停止等待 goroutine，但只要 main 函数还没返回，仍在运行的 goroutine 就会继续占用内存。所以在情况运行的情况下，我们还是应该尽量结束无用的 goroutine。


**什么都不做的 nil 通道**

如果创建通道的时候，没有显式的使用 make 函数，那么通道的值是默认的零值，也就是 nil。

对值为 nil 的通道执行发送或者接收操作并不会引发 panic，但是会导致操作永久阻塞，就好像遇到了一个从来没有接收或者发送过任何值的通道一样。但是你如果尝试对值为 nil 的通道执行稍后将要介绍的 close 函数，那么该函数将引发 panic。

初看上去，值为 nil 的通道似乎没有什么用处，但事实恰恰相反。例如，对于一个包含 select 语句的循环，如果我们不希望程序在每次循环的时候都等待 select 语句涉及的所有通道，那么可以先将某些通道设置为 nil，等到待发送的值准备就绪之后，再为通道变量赋予一个非 nil 值并执行实际的发送操作。


> select 语句的每个分支可以包含一个通道操作。

## 阻塞和死锁

### 阻塞
当 goroutine 在等待通道的发送或者接收操作的时候，我们就说它被**阻塞**了。

发生阻塞之后，除 goroutine 本身占用的少量内存之外，被阻塞的 goroutine 并不消耗任何资源。goroutine 会静静地停在那里，等待导致它阻塞的事情发生，然后接触阻塞。

### 死锁

当一个或者多个 goroutine 因为某些永远无法发生的事情而被阻塞时，我们称这种情况为**死锁**。

而出现死锁的程序通常都会崩溃或者被挂起。

比如，下面的代码会引发死锁：
```go
func main(){
    c := make(chan int)
    <-c
}
```

> 被阻塞的 goroutine 会做什么？ 答案：什么都不做。

## 一个流水线的例子

整条流水线共分为源头、过滤和打印3个阶段。

### 源头

流水线的起始端，也就是源头，它只会发送值而不会读取任何值。

其他程序的流水线起始端通常会从文件、数据库或者网络中读取数据，但本例子只会发送几个任意的值。

为了在所有值均已发送完成时通知下游程序，本程序使用了空字符串作为**哨兵值**，并将其用作于标识发送已经完成。

代码清单 30-7 源头 pipeline1.go
```go
func sourceGopher(downstream chan string) {
    for _, v := range []string{"hello world", "a bad apple", "goodbye all"} 
    {
        downstream <- v
    }
    downstream <- ""
}
```

### 过滤

本程序会过滤上游所有不好的东西。

具体来说，这个函数会从上游通道中读取值，并在字符串值不为 "bad" 的情况下将其发至下游通道。当函数见到结尾的空字符串时，它会停止筛选工作，并确保空字符串也发送给下游的程序。

代码清单 30-8 过滤 pipeline1.go
```go
func filterGopher(upstream, downstream, chan string) {
    for {
        item := <- upstream
        if item == "" {
            downstream <- ""
            return
        }
        if !strings.Contains(item, "bad") {
            downstream <- item
        }
    }
}
```

### 打印

位于流水线最末端的是打印程序，它没有任何下游。

在其他程序中，位于流水线末端的函数通常会将结果存储到文件或者数据库里面，或者将这些结果打印出来。

代码清单 30-9 打印 pipeline1.go
```go
func printGopher(uostream chan string) {
    for {
        v := <- upstream
        if v == "" {
            return
        }
        fmt.Println(v)
    }
}
```

### 组装

从上面可以看到，整个流水线包含源头、过滤和打印3个阶段，但是只用到两个通道。

因为我们希望可以在整条流水线都被处理完成之后再退出程序，所有没有为最后一个程序创建新的 goroutine。当 printGopher 函数返回的时候，我们可以确认其他两个 goroutine 已经完成了工作，而且 printGopher 也可以顺利的返回至 main 函数，然后完成整个程序。

代码清单 30-10 组装 pipeline1.go
```go
func main() {
    c0 := make(chan string)
    c1 := make(chan string)
    go sourceGopher(c0)
    go filterGopher(c0, c1)
    printGopher(c1)
}
```

到目前为止，整个流水线程序虽然可以正常运行，但是它有一个问题：

程序使用了空字符串来表示所有值均已发送完毕，但是当它需要像处理其他值一样处理空字符串的时候，该怎么办呢？为此，我们可以使用结构值来代替单纯的字符串值，在结构里面分别包含一个字符串和一个布尔值，并使用布尔值来表示当前字符串是否是最后一个值。

但事实上，还有一个更好的办法。

Go 允许在没有值可供发送的情况下通过 close 函数关闭通道：
```go
close(c)
```

通道被关闭之后将无法写入任何值，如果尝试写入值将会引发 panic。尝试读取已被关闭的通道将会获得一个与通道类型对应的零值，而这个零值就可以代替上述程序中的空字符串。

> 注意！
> 
> 如果在循环里面读取一个已关闭的通道，并且没有检查该通道是否已经关闭，那么这个循环将一直运转下去，并消耗大量的处理器时间。
>
> 为避免这种情况发生，务必对那些可能会被关闭的通道做相应的检查。

执行以下代码可以获取通道是否已经被关闭：
```go
v, ok := <- c
```

通过将接收操作的执行结果赋值给两个变量，可以根据第二个变量的值来判断此次通道读取操作是否成功。
如果改变了的值是false，那么说明通道已被关闭。

### 优化

代码清单 30-11 修改后的源头 pipelien2.go
```go
func sourceGopher(downstream chan string) {
    for _, v := range []string{"hello world", "a bad apple", "goodbye all"} 
    {
        downstream <- v
    }
    close(downstream)
}
```

代码清单 30-12 修改后的过滤 pipeline2.go
```go
func filterGopher(upstream, downstream, chan string) {
    for {
        item, ok := <- upstream
        if !ok {
            close(downstream)
            return
        }
        if !strings.Contains(item, "bad") {
            downstream <- item
        }
    }
}
```

因为"从通道里面读取值，直到它被关闭为止"这种模式实在太常用了， 所以 Go 提供了更快捷的方式。

通过在 range 语句里面使用通道，程序可以在通道被关闭之前，一直从通道里面读取值。


代码清单 30-13 使用 range 实现的过滤 pipeline2.go
```go
func filterGopher(upstream, downstream, chan string) {
    for item := range upstream {
        if !strings.Contains(item, "bad") {
            downstream <- item
        }
    }
    close(downstream)
}
```

代码清单 30-14 使用 range 实现的打印 pipeline2.go
```go
func printGopher(uostream chan string) {
    for v:= range upstream {
        fmt.Println(v)
    }
}
```

## 小结

- 使用 go 语句可以启动一个新的 goroutine，并且这个 goroutine 将以并发方式运行。
- 通道（channel）用于在多个 goroutine 之间传递值。
- 创建通道需要使用内置函数 make 函数。
- 从通道里面接收值，程序需要将 <- 操作符放在通道值的前面。
- 将值发送至通道，程序需要将 <- 操作符放在通道值和待发送值的中间。
- close 函数可以关闭一个通道。
- range 语句可以从通道中读取所有值，直到通道关闭为止。