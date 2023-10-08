package monitor

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/FredyXue/go-utils/testdata"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"golang.org/x/sync/errgroup"
)

var array []string

// 测试任务重入
func TestMonitorReentry(t *testing.T) {
	redisClient := testdata.NewTestRedis()
	group := "test_monitor"
	var g errgroup.Group
	array = make([]string, 0) // 清空比对数据

	// m1 master
	g.Go(func() error {
		m1 := NewMonitor(redisClient,
			WithHeartbeatTime(time.Millisecond*10), // 10ms 心跳
			WithHeartbeatTimeout(time.Second*2),    // 2s 心跳超时
			WithWatchTimeout(time.Minute),
		)
		m1.Register("test_method", func(mctx MonitorContext) {
			// 获取上下文
			body, err := mctx.Get()
			assert.NoError(t, err)
			obj := DoObj{}
			json.Unmarshal(body, &obj)

			MethodDo(obj)
		})
		// 开启 m1
		m1.Start(group)

		time.Sleep(time.Millisecond * 5) // 等待 5ms 使 m1 选举完成
		assert.Equal(t, m1.IsMaster(), true)

		// exec
		mctx, err := m1.Watch("test_method", "1")
		assert.NoError(t, err)

		obj := DoObj{
			Title:  "m1",
			Number: 1,
		}
		data, _ := json.Marshal(obj)
		err = mctx.Set(data) // 设置上下文
		assert.NoError(t, err)

		MethodDo(obj)

		time.Sleep(time.Second * 3)          // 3s 等待 m3, m4 重入完成
		err = m1.Unwatch("test_method", "1") // 完成 mctx
		assert.NoError(t, err)

		return nil
	})

	time.Sleep(time.Millisecond * 10) // 等待 10ms 使 m1 成为 master

	// m2 正常执行完成，不重入
	g.Go(func() error {
		m2 := NewMonitor(redisClient,
			WithHeartbeatTime(time.Millisecond*10),
			WithHeartbeatTimeout(time.Second*2),
			WithWatchTimeout(time.Minute),
		)
		m2.Register("test_method", func(mctx MonitorContext) {
			body, err := mctx.Get()
			assert.NoError(t, err)
			obj := DoObj{}
			json.Unmarshal(body, &obj)

			MethodDo(obj)
		})
		m2.Start(group)

		time.Sleep(time.Millisecond * 5) // 等待 5ms 选举
		assert.Equal(t, m2.IsMaster(), false)

		// exec
		mctx, err := m2.Watch("test_method", "2")
		assert.NoError(t, err)

		obj := DoObj{
			Title:  "m2",
			Number: 2,
		}
		data, _ := json.Marshal(obj)
		err = mctx.Set(data)
		assert.NoError(t, err)

		MethodDo(obj)

		err = m2.Unwatch("test_method", "2") // 正常执行完成，不重入
		assert.NoError(t, err)

		m2.Stop()
		return nil
	})

	// m3 主动退出
	g.Go(func() error {
		m3 := NewMonitor(redisClient,
			WithHeartbeatTime(time.Millisecond*10),
			WithHeartbeatTimeout(time.Second*2),
			WithWatchTimeout(time.Minute),
		)
		m3.Register("test_method", func(mctx MonitorContext) {
			body, err := mctx.Get()
			assert.NoError(t, err)
			obj := DoObj{}
			json.Unmarshal(body, &obj)

			MethodDo(obj)
		})
		m3.Start(group)

		time.Sleep(time.Millisecond * 5) // 等待 5ms 选举
		assert.Equal(t, m3.IsMaster(), false)

		// exec
		obj := DoObj{
			Title:  "m3",
			Number: 3,
		}
		data, _ := json.Marshal(obj)
		_, err := m3.Watch("test_method", "3", data) // watch 时即设置上下文
		assert.NoError(t, err)

		m3.Stop() // 主动退出，不执行相关的 Unwatch
		return nil
	})

	// m4 模拟进程崩溃
	g.Go(func() error {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("[Panic] catch panic: %v", r)
			}
		}()

		m4 := NewMonitor(redisClient,
			WithHeartbeatTime(time.Second*5), // 延长心跳时间，避免定时器执行
			WithHeartbeatTimeout(time.Second*2),
			WithWatchTimeout(time.Minute),
		)
		m4.Register("test_method", func(mctx MonitorContext) {
			body, err := mctx.Get()
			assert.NoError(t, err)
			obj := DoObj{}
			json.Unmarshal(body, &obj)

			MethodDo(obj)
		})
		m4.Start(group)

		time.Sleep(time.Millisecond * 5) // 等待 5ms 选举
		assert.Equal(t, m4.IsMaster(), false)

		// exec
		mctx, err := m4.Watch("test_method", "4")
		assert.NoError(t, err)

		obj := DoObj{
			Title:  "m4",
			Number: 4,
		}
		data, _ := json.Marshal(obj)
		err = mctx.Set(data)
		assert.NoError(t, err)

		panic("exit m4") // 模拟进程崩溃
	})

	g.Wait()

	// 执行顺序 m1,m2,m3,m4
	assert.Equal(t, array, []string{"m1_1", "m2_2", "m3_3", "m4_4"})
}

type DoObj struct {
	Title  string
	Number int64
}

func (d *DoObj) String() string {
	return fmt.Sprintf("%s_%d", d.Title, d.Number)
}

func MethodDo(obj DoObj) {
	log.Println("MethodDo", obj)
	array = append(array, obj.String())
}

func TestMonitorWatchTimeout(t *testing.T) {

	redisClient := testdata.NewTestRedis()
	group := "test_monitor_watchtimeout"

	m1 := NewMonitor(redisClient,
		WithHeartbeatTime(time.Millisecond*100),    // 100ms 心跳
		WithHeartbeatTimeout(time.Second*10),       // 10s 心跳超时
		WithWatchTimeout(time.Second),              // 1s watch 最大时长
		WithWatchWarningTime(time.Millisecond*300), // 300ms 长耗时预警
	)
	m1.Register("test_method_1", func(mctx MonitorContext) {
		// 获取上下文
		body, err := mctx.Get()
		assert.NoError(t, err)
		obj := DoObj{}
		json.Unmarshal(body, &obj)

		MethodDo(obj)
	})
	m1.Register("test_method_2",
		func(mctx MonitorContext) {},
		CallOpt{WatchWarningTime: time.Millisecond * 600}) // 600ms 方法级长耗时
	// 开启 m1
	m1.Start(group)

	time.Sleep(time.Millisecond * 5) // 等待 5ms 使 m1 选举完成
	assert.Equal(t, m1.IsMaster(), true)

	// exec
	testKey := "test_monitor_watchtimeout|test_method_1|1"
	mctx, err := m1.Watch("test_method_1", "1")
	assert.NoError(t, err)

	keys, err := m1.WatchList()
	assert.NoError(t, err)
	assert.Equal(t, len(keys), 1)
	assert.Equal(t, keys[0], testKey) // 测试任务列表

	err = mctx.Set([]byte{}) // 设置空上下文
	assert.NoError(t, err)

	// 测试方法级长耗时，不设置 mctx
	_, err = m1.Watch("test_method_2", "2", nil)
	assert.NoError(t, err)

	// 3s 等待 mctx 超时。 不会发生重入
	// 预期出现长耗时预警, test_method_1 匹配全局配置, test_method_2 匹配方法级配置
	time.Sleep(time.Second * 3)

	_, err = mctx.Get()
	assert.Equal(t, errors.Is(err, redis.Nil), true) // mctx 已被 Close

	keys, err = m1.WatchList()
	assert.NoError(t, err)
	assert.Equal(t, len(keys), 0)

	err = m1.Unwatch("test_method_1", "1") // 手动完成 mctx, 期望此处不会报错
	assert.NoError(t, err)
	err = m1.Unwatch("test_method_2", "2")
	assert.NoError(t, err)
}
