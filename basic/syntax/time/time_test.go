package time

import (
	"context"
	"testing"
	"time"
)

func TestTicker(t *testing.T) {
	// ticker相当于每隔一段时间就执行一次
	tk := time.NewTicker(time.Second)
	defer tk.Stop()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	for {
		select {
		case <-ctx.Done():
			t.Log("超时或被主动取消了")
			goto end
		case now := <-tk.C:
			t.Log(now.String())
		}
	}
end:
	t.Log("退出循环")
}

func TestTimer(t *testing.T) {
	// 精确到 12:00 怎么用 timer
	now := time.Now().Unix()
	t.Log(now)
	tm := time.NewTimer(time.Second * 10)
	defer tm.Stop()
	go func() {
		for now := range tm.C {
			t.Log(now.Unix())
		}
	}()

	time.Sleep(time.Second * 10)
}
