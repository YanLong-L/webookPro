package channel

import (
	"runtime"
	"testing"
	"time"
)

func TestChannel(t *testing.T) {
	// 声明一个放int类型的channel
	//var ch chan int
	//ch <- 123
	//val := <-ch
	//print(val)
	// 以上这种写法，只是声明了，但是没有初始化，读写都会panic

	//ch1 := make(chan int) // 这种是不带容量的
	ch2 := make(chan int, 2) // 这种是带容量的，而且容量是不会变的
	ch2 <- 123               // 我放了一个数据
	close(ch2)               // 此时我关闭了ch2
	val, ok := <-ch2
	if !ok {
		t.Log("ch2被人关了")
	}
	println(val)
	val = <-ch2
	println(val)
}

func TestChannelClose(t *testing.T) {
	ch := make(chan int, 2)
	ch <- 123
	ch <- 234
	val, ok := <-ch
	t.Log(val, ok) //123 true
	close(ch)

	// 能不能把 234 读出来？
	val, ok = <-ch
	t.Log(val, ok) // 234 true 为啥是true呢，？ 因为false必须是channel被关闭且同时读不出数据时，才会为false

	val, ok = <-ch
	t.Log(val, ok) // 0 false
}

func TestChannelBlock(t *testing.T) {
	ch := make(chan int, 3)
	// 阻塞，不再占用 CPU 了
	val := <-ch
	// 意味着，这一句不会执行下来
	t.Log(val)
	// goroutine 数量
	runtime.NumGoroutine()
}

func TestSelect(t *testing.T) {
	ch1 := make(chan int, 1)
	ch2 := make(chan int, 1)
	//ch1 <- 123
	//ch2 <- 234
	go func() {
		time.Sleep(time.Millisecond * 100)
		ch1 <- 123
	}()
	go func() {
		time.Sleep(time.Millisecond * 100)
		ch2 <- 123
	}()
	select {
	case val := <-ch1:
		t.Log("ch1", val)
		val = <-ch2
		t.Log("ch2", val)
	case val := <-ch2:
		t.Log("ch2", val)
		val = <-ch1
		t.Log("ch1", val)
	}

	println("往后执行")
}

func TestGoroutineCh(t *testing.T) {
	ch := make(chan int)
	// 这一个就泄露掉了
	go func() {
		// 永久阻塞在这里
		ch <- 123
	}()

	// 这里后面没人往 ch 里面读数据
}

func TestLoopChannel(t *testing.T) {
	ch := make(chan int)
	go func() {
		for i := 0; i < 10; i++ {
			ch <- i
			time.Sleep(time.Millisecond * 100)
		}
		close(ch)
	}()
	for val := range ch {
		t.Log(val)
	}

	t.Log("channel 被关了")

	//for {
	//	val, ok := <-ch
	//	if !ok {
	//		break
	//	}
	//	t.Log(val)
	//}
}
