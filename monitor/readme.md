# monitor
monitor 主要用于监控异步任务。支持以下功能：
1. 任务重入：当某一个进程节点崩溃后，其未完成的任务，会被选举出的主节点重新执行。
2. 长耗时任务预警：当一些任务运行时长超出预警设置时，会触发预警。



### monitor init
``` go
// 全局仅需初始化一次
m1 := NewMonitor(redisClient,
  WithHeartbeatTime(time.Minute),         // 设置心跳, 默认 1 分钟
  WithHeartbeatTimeout(time.Minute*3),    // 设置心跳超时, 默认 3 分钟
  WithWatchTimeout(time.Hour * 2),        // watch 最大超时, 默认 2h
  WithWatchWarningTime(time.Minute*10),   // watch 长耗时任务预警, 默认不预警
  WithAlertFunc(...),    // 设置预警 func  
)

// 注册 monitor func
m1.Register("test_method", func(mctx MonitorContext) {
  // 获取上下文
  body, err := mctx.Get()
  obj := DoObj{}
  json.Unmarshal(body, &obj)

  // 任务重入
  MethodDo(obj)
})

// 开启 monitor, 指定一个 group 业务分组
m1.Start("test_group")



// 测试方法
func MethodDo(obj DoObj) {
  log.Println("MethodDo", obj)
}
```


### monitor exec
``` go
// 开始执行任务

// 准备上下文数据
obj := DoObj{
  Title:  "m1",
  Number: 1,
}
data, _ := json.Marshal(obj)

// watch 并设置上下文
mctx, err := m1.Watch("test_method", "1", data)

// 支持中途更新上下文
err = mctx.Set(data)

// 执行具体任务
MethodDo(obj)

// 任务执行完成, unwatch
err = m1.Unwatch("test_method", "1") 
```