package monitor

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/FredyXue/go-utils"
	"github.com/FredyXue/go-utils/monitor/locker"
	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
)

type Monitor interface {
	Start(group string)                                                                  // 开启 monitor
	Stop()                                                                               // 主动退出 monitor，否则等进程心跳超时才空出位置。
	Register(method string, fn Callback, copt ...CallOpt)                                // 注册 callback
	Deregister(method string)                                                            // 注销 callback
	Watch(method string, tag string, ctxData ...[]byte) (mctx MonitorContext, err error) // 任务监控
	Unwatch(method string, tag string) error                                             // 任务完成，解除监控
	WatchList() (list []string, err error)                                               // 存活任务列表
	IsMaster() bool
}

// CallOpt
type CallOpt struct {
	WatchWarningTime time.Duration // watch 长耗时任务预警
}

// MOpt
type MOpt func(*monitorImpl)

func WithHeartbeatTime(heartbeatTime time.Duration) MOpt {
	return func(r *monitorImpl) {
		r.heartbeatTime = heartbeatTime
	}
}

func WithHeartbeatTimeout(heartbeatTimeout time.Duration) MOpt {
	return func(r *monitorImpl) {
		r.heartbeatTimeout = heartbeatTimeout
	}
}

func WithWatchTimeout(watchTimeout time.Duration) MOpt {
	return func(r *monitorImpl) {
		r.watchTimeout = watchTimeout
	}
}

func WithWatchWarningTime(watchWarningTime time.Duration) MOpt {
	return func(r *monitorImpl) {
		r.watchWarningTime = watchWarningTime
	}
}

func WithAlertFunc(alertFunc AlertFunc) MOpt {
	return func(r *monitorImpl) {
		r.alertFunc = alertFunc
	}
}

type Callback func(mctx MonitorContext)
type AlertFunc func(msg string)

type localWatch struct {
	mctx      MonitorContext
	startAt   time.Time // 任务开始时间
	warnCount int       // 预警次数
	method    string
}

// monitorImpl .
type monitorImpl struct {
	once             sync.Once
	ctx              context.Context
	cancelCtx        context.Context
	cancel           context.CancelFunc
	lock             locker.Locker
	cli              *redis.Client
	alertFunc        AlertFunc                 // 预警方法
	uid              string                    // 节点 id
	role             int32                     // 角色：0 worker 节点，1 master 节点
	callbackMap      map[string]Callback       // method -> callback
	watchWarningMap  map[string]time.Duration  // method -> WatchWarningTime 方法级长耗时预警
	nodeMap          map[string]int64          // uid -> timestamp;   master 进程维护的节点列表
	watchMap         map[string]MonitorContext // key -> mctx;   key = group|method|tag;  master 进程维护的 mctx 列表
	localWatchMap    map[string]*localWatch    // key -> mctx;   本地进程维护的 mctx 列表;
	group            string                    // 业务分组   STR
	groupList        string                    // 节点列表   ZSET  uid -> timestamp
	groupWatchList   string                    // 任务列表   HASH  key -> uid
	heartbeatTime    time.Duration             // 心跳轮询时间
	heartbeatTimeout time.Duration             // 心跳超时时间
	watchTimeout     time.Duration             // watch 最大超时时间
	watchWarningTime time.Duration             // watch 全局长耗时任务预警
}

func NewMonitor(cli *redis.Client, opts ...MOpt) Monitor {
	m := &monitorImpl{
		ctx:              context.Background(),
		cli:              cli,
		role:             0,
		callbackMap:      make(map[string]Callback),
		watchWarningMap:  make(map[string]time.Duration),
		nodeMap:          make(map[string]int64),
		watchMap:         make(map[string]MonitorContext),
		localWatchMap:    make(map[string]*localWatch),
		heartbeatTime:    time.Minute,     // 默认心跳 1 分钟轮询
		heartbeatTimeout: time.Minute * 3, // 默认心跳超时 3 分钟
		watchTimeout:     time.Hour * 2,   // 默认 watch 最大超时 2h
		watchWarningTime: 0,               // 默认全局不预警长耗时任务
	}
	for _, o := range opts {
		o(m)
	}
	m.uid = strings.ReplaceAll(uuid.NewV4().String(), "-", "")

	// 使用 RedisLocker 作为心跳工具
	m.lock = locker.NewRedisLocker(cli,
		locker.WithRefreshTime(m.heartbeatTime),
		locker.WithLockTime(m.heartbeatTimeout),
		locker.WithExpiredTime(time.Duration(1<<63-1)), // maxDuration
	)
	return m
}

// Start 开启 monitor，使用 group 来业务分组
// 先注册 callback, 然后 Start
func (m *monitorImpl) Start(group string) {
	tickerRun := func() {
		m.heartbeat()           // 心跳
		m.checkLocalWatchList() // 检测本地任务
		if m.role == 0 {
			m.election() // worker 角色，选举
		}
		if m.role == 1 {
			m.checkNodeList()  // master 检测节点列表
			m.checkWatchList() // master 检测任务列表
		}
	}

	m.once.Do(func() {
		m.group = group
		m.groupList = group + ":List"
		m.groupWatchList = group + ":WatchList"

		go utils.Protect(func() {
			m.cancelCtx, m.cancel = context.WithCancel(context.TODO()) // 建立 cancel ctx
			ticker := time.NewTicker(m.heartbeatTime)
			defer ticker.Stop()

			tickerRun() // 首次立刻执行
			log.Printf("[Monitor] Start. group: %s, uid: %s", m.group, m.uid)

		tickerLabel:
			for {
				select {
				case <-ticker.C:
					tickerRun()
				case <-m.cancelCtx.Done():
					break tickerLabel // ctx canceled 结束循环
				}
			}
		})
	})
}

// Stop 主动退出 monitor，否则等进程心跳超时才空出位置。
func (m *monitorImpl) Stop() {
	m.cancel()      // 取消定时器
	m.lock.Unlock() // 即使未加锁，解锁也不会报错
	m.role = 0      // 恢复为 worker

	// 从节点列表移除
	if err := m.cli.ZRem(m.ctx, m.groupList, m.uid).Err(); err != nil {
		log.Println("[Monitor] Logout ZRem Error:", err)
	}

	log.Printf("[Monitor] Stop. group: %s, uid: %s", m.group, m.uid)
}

// heartbeat 维护节点心跳
func (m *monitorImpl) heartbeat() {
	z := &redis.Z{
		Score:  float64(time.Now().Unix()),
		Member: m.uid,
	}
	if err := m.cli.ZAdd(m.ctx, m.groupList, z).Err(); err != nil {
		log.Println("[Monitor] heartbeat Error:", err)
	}
}

// election 抢占式选举
func (m *monitorImpl) election() {
	success, err := m.lock.Lock(m.group)
	if err != nil {
		log.Println("[Monitor] election Lock Error:", err)
		return
	}
	if success {
		m.role = 1 // 升级为 master
		log.Println("[Monitor] election master success:", m.group, m.uid)
	}
}

// Register 注册 callback
// callback 在异常情况下，可能会被调用多次。业务层需要做好相应的幂等操作。
// callback 执行失败，不会再次重入。业务层需要自己处理好执行失败的逻辑，例如重试。
// 重入只解决进程崩溃导致的任务丢失。
func (m *monitorImpl) Register(method string, fn Callback, copt ...CallOpt) {
	m.callbackMap[method] = fn

	// 方法级配置
	if len(copt) == 0 {
		return
	}
	opt := copt[0]

	// 长耗时预警
	if opt.WatchWarningTime > 0 {
		m.watchWarningMap[method] = opt.WatchWarningTime
	}
}

// Deregister 注销 callback
func (m *monitorImpl) Deregister(method string) {
	delete(m.callbackMap, method)
}

// Watch 任务监控
// 存储 key = group|method|tag
// return ctx 上下文
func (m *monitorImpl) Watch(method string, tag string, ctxData ...[]byte) (mctx MonitorContext, err error) {
	_, has := m.callbackMap[method]
	if !has {
		err = errors.Errorf("[Monitor] Watch Error: method %s is unregistered", method)
		return
	}

	key := m.group + "|" + method + "|" + tag

	// add to watchList
	if err = m.cli.HSet(m.ctx, m.groupWatchList, key, m.uid).Err(); err != nil {
		err = errors.Wrap(err, "[Monitor] Watch HSet Error")
		return
	}
	// new 上下文。原始 mctx, 任务首次 watch 时使用。
	mctx = NewMonitorContext(m.cli, key, m.watchTimeout)
	// for unwatch close
	m.localWatchMap[key] = &localWatch{
		mctx:    mctx,
		startAt: time.Now(),
		method:  method,
	}

	// 设置上下文数据
	if len(ctxData) > 0 {
		err = mctx.Set(ctxData[0])
		err = errors.Wrap(err, "[Monitor] Watch mctx.Set Error")
		return
	}

	return
}

// Unwatch 任务完成，解除监控
func (m *monitorImpl) Unwatch(method string, tag string) (err error) {
	key := m.group + "|" + method + "|" + tag
	// remove from watchList
	if err = m.cli.HDel(m.ctx, m.groupWatchList, key).Err(); err != nil {
		return errors.Wrap(err, "[Monitor] Unwatch HDel Error")
	}
	// 清理上下文
	lw, has := m.localWatchMap[key]
	if !has {
		return
	}
	if err = lw.mctx.Close(); err != nil {
		return errors.Wrap(err, "[Monitor] Unwatch Close Error")
	}
	delete(m.localWatchMap, key)
	return
}

// checkLocalWatchList 检测本地任务状态
func (m *monitorImpl) checkLocalWatchList() {
	// 未设置长耗时任务预警，直接跳过
	if m.watchWarningTime == 0 && len(m.watchWarningMap) == 0 {
		return
	}

	for key, lw := range m.localWatchMap {
		warningTime := m.watchWarningTime
		methodWaringTime, has := m.watchWarningMap[lw.method]
		if has {
			warningTime = methodWaringTime
		}

		// 超时 且 未预警的任务
		if time.Since(lw.startAt) > warningTime && lw.warnCount == 0 {
			msg := fmt.Sprintf("[Monitor] find executed too long MonitorContext cost: %v, key: %s", time.Since(lw.startAt), key)
			log.Println(msg)
			m.alert(msg) // 预警长耗时任务
			lw.warnCount++
		}
	}
}

// WatchList 存活任务列表
// return group|method|tag
func (m *monitorImpl) WatchList() (list []string, err error) {
	list, err = m.cli.HKeys(m.ctx, m.groupWatchList).Result()
	if err != nil {
		err = errors.Wrap(err, "[Monitor] WatchList Error")
		return
	}
	return
}

// checkNodeList 检测节点列表
func (m *monitorImpl) checkNodeList() {
	listZ, err := m.cli.ZRangeWithScores(m.ctx, m.groupList, 0, -1).Result()
	if err != nil {
		log.Println("[Monitor] checkNodeList ZRangeWithScores Error:", err)
		return
	}

	m.nodeMap = make(map[string]int64) // 重新初始化
	for _, z := range listZ {
		timestamp := int64(z.Score)
		uid := z.Member.(string)

		// now - heartbeatTimeout < timestamp 心跳未超时
		if uid != "" && time.Now().Add(-m.heartbeatTimeout).Unix() < timestamp {
			m.nodeMap[uid] = timestamp
		} else {
			// 从节点列表移除
			if err := m.cli.ZRem(m.ctx, m.groupList, uid).Err(); err != nil {
				log.Println("[Monitor] checkNodeList ZRem Error:", err)
			}
		}
	}
}

// checkWatchList 检测任务列表
// 只有 master 会执行该方法，以便获取重入的 mctx
// 构建 watchMap, 超时的 mctx，会 close
// 判断节点丢失的 mctx，会发起任务重入
func (m *monitorImpl) checkWatchList() {
	maps, err := m.cli.HGetAll(m.ctx, m.groupWatchList).Result()
	if err != nil {
		log.Println("[Monitor] checkWatchList HGetAll Error:", err)
		return
	}

	oldWatchMap := m.watchMap
	m.watchMap = make(map[string]MonitorContext) // 重新初始化
	for key, uid := range maps {
		mctx, has := oldWatchMap[key]
		if !has {
			// 构建新的 mctx
			mctx = NewMonitorContext(m.cli, key, m.watchTimeout)
		}

		valid, err := mctx.Check()
		if err != nil {
			log.Printf("[Monitor] checkWatchList Check key: %s, Error: %v", key, err)
			continue
		}
		// 无效 mctx，移除
		if !valid {
			msg := fmt.Sprintf("[Monitor] remove invalid MonitorContext key: %s", key)
			log.Println(msg)
			m.alert(msg) // 触发移除 mctx 时，预警

			if err = mctx.Close(); err != nil {
				log.Println(err)
			}
			if err = m.cli.HDel(m.ctx, m.groupWatchList, key).Err(); err != nil {
				log.Println("[Monitor] checkWatchList HDel Error:", err)
			}
			continue
		}

		m.watchMap[key] = mctx // 存储有效的 mctx

		if _, has = m.nodeMap[uid]; !has {
			// 节点丢失，触发任务重入
			m.reentry(mctx, key)
		}
	}
}

// reentry 任务重入
func (m *monitorImpl) reentry(mctx MonitorContext, key string) {
	// 判断是否正在重入
	_, has := m.localWatchMap[key]
	if has {
		return
	}

	msg := fmt.Sprintf("[Monitor] execute reentry MonitorContext key: %s", key)
	log.Println(msg)
	m.alert(msg) // 触发重入时，预警

	arr := strings.Split(key, "|")
	if len(arr) <= 2 {
		log.Printf("[Monitor] checkWatchList invalide key: %s", key)
		return
	}
	method, tag := arr[1], arr[2]
	callback, has := m.callbackMap[method]
	if !has {
		log.Printf("[Monitor] checkWatchList callback method not found: %s", method)
		return
	}

	// 执行任务重入
	go utils.Protect(func() {
		// 加入 localWatch
		m.localWatchMap[key] = &localWatch{
			mctx:    mctx,
			startAt: time.Now(),
			method:  method,
		}
		defer delete(m.localWatchMap, key) // finally 从 localWatch 移除。

		// callback 在异常情况下，可能会被调用多次。业务层需要做好相应的幂等操作。
		// callback 执行失败，不会再次重入。业务层需要自己处理好执行失败的逻辑，例如重试。
		// 重入只解决进程崩溃导致的任务丢失。
		callback(mctx)

		if err := m.Unwatch(method, tag); err != nil {
			log.Printf("[Monitor] reentry Error: %v", err) // 报错直接返回，等待下一次重入
			return
		}
	})
}

func (m *monitorImpl) IsMaster() bool {
	return m.role == 1
}

func (m *monitorImpl) alert(msg string) {
	if m.alertFunc != nil {
		m.alertFunc(msg)
	}
}

// ....
// --------------------------------------------- MonitorContext
// ....

// MonitorContext 上下文信息载体
type MonitorContext interface {
	Get() ([]byte, error)
	Set([]byte) error
	Check() (valid bool, err error)
	Close() error
}

type monitorContext struct {
	ctx         context.Context
	cli         *redis.Client
	key         string
	closed      bool
	expiredDur  time.Duration
	expiredTime time.Time
}

func NewMonitorContext(cli *redis.Client, key string, expiredDur time.Duration) MonitorContext {
	return &monitorContext{
		ctx:         context.TODO(),
		cli:         cli,
		key:         key,
		expiredDur:  expiredDur,
		expiredTime: time.Now().Add(expiredDur),
	}
}

// Get .
func (c *monitorContext) Get() (body []byte, err error) {
	rlt, err := c.cli.Get(c.ctx, c.key).Result()
	if err != nil {
		err = errors.Wrap(err, "[MonitorContext] Get Error")
		return
	}
	body = []byte(rlt)
	return
}

// Set 支持中途多次更新上下文
// 为了保留 key, 允许设置空 value
func (c *monitorContext) Set(body []byte) error {
	// if len(body) == 0 {
	// 	return nil
	// }
	c.expiredTime = time.Now().Add(c.expiredDur) // 更新超时时间
	err := c.cli.Set(c.ctx, c.key, body, c.expiredDur).Err()
	return errors.Wrap(err, "[MonitorContext] Set Error")
}

// check 检验 mctx 是否有效
func (c *monitorContext) Check() (valid bool, err error) {
	if c.closed {
		return
	}

	// 超时直接清理
	if time.Now().After(c.expiredTime) {
		err = c.Close()
		return
	}

	ttl, err := c.cli.TTL(c.ctx, c.key).Result()
	if err != nil {
		err = errors.Wrap(err, "[MonitorContext] check TTL Error")
		return
	}
	if ttl > 0 || ttl == -1 {
		valid = true
	}
	return
}

// Close 清理上下文
func (c *monitorContext) Close() (err error) {
	if err = c.cli.Del(c.ctx, c.key).Err(); err != nil {
		return errors.Wrap(err, "[MonitorContext] Close Error")
	}
	c.closed = true
	return
}
