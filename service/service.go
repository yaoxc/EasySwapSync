// service 包，定义服务相关的结构体和方法
package service

// 导入标准库和依赖包
import (
	"context" // 上下文管理
	"fmt"     // 格式化输出
	"sync"    // 并发同步

	"github.com/pkg/errors"                           // 错误处理
	"github.com/yaoxc/EasySwapBase/chain"             // 链相关常量
	"github.com/yaoxc/EasySwapBase/chain/chainclient" // 区块链客户端
	"github.com/yaoxc/EasySwapBase/ordermanager"      // 订单管理器
	"github.com/yaoxc/EasySwapBase/stores/xkv"        // KV 存储
	"github.com/zeromicro/go-zero/core/stores/cache"  // go-zero 缓存
	"github.com/zeromicro/go-zero/core/stores/kv"     // go-zero KV
	"github.com/zeromicro/go-zero/core/stores/redis"  // go-zero Redis
	"gorm.io/gorm"                                    // ORM 数据库

	"github.com/yaoxc/EasySwapSync/service/orderbookindexer" // 订单簿同步器

	"github.com/yaoxc/EasySwapSync/model"                    // 数据模型
	"github.com/yaoxc/EasySwapSync/service/collectionfilter" // 集合过滤器
	"github.com/yaoxc/EasySwapSync/service/config"           // 配置
)

// Service 主服务结构体，包含各类依赖和组件
type Service struct {
	ctx              context.Context            // 全局上下文
	config           *config.Config             // 配置
	kvStore          *xkv.Store                 // KV 存储
	db               *gorm.DB                   // 数据库连接
	wg               *sync.WaitGroup            // 并发等待组
	collectionFilter *collectionfilter.Filter   // 集合过滤器
	orderbookIndexer *orderbookindexer.Service  // 订单簿同步器
	orderManager     *ordermanager.OrderManager // 订单管理器
}

// New 构造 Service 实例，初始化各类依赖
func New(ctx context.Context, cfg *config.Config) (*Service, error) {
	var kvConf kv.KvConf // KV 配置
	// 遍历 Redis 配置，组装 KV 节点配置
	for _, con := range cfg.Kv.Redis {
		kvConf = append(kvConf, cache.NodeConf{
			RedisConf: redis.RedisConf{
				Host: con.Host, // Redis 主机
				Type: con.Type, // Redis 类型
				Pass: con.Pass, // Redis 密码
			},
			Weight: 2, // 节点权重
		})
	}

	kvStore := xkv.NewStore(kvConf) // 创建 KV 存储

	var err error
	db := model.NewDB(cfg.DB) // 初始化数据库连接

	collectionFilter := collectionfilter.New(ctx, db, cfg.ChainCfg.Name, cfg.ProjectCfg.Name)  // 创建集合过滤器
	orderManager := ordermanager.New(ctx, db, kvStore, cfg.ChainCfg.Name, cfg.ProjectCfg.Name) // 创建订单管理器
	var orderbookSyncer *orderbookindexer.Service                                              // 订单簿同步器
	var chainClient chainclient.ChainClient                                                    // 区块链客户端
	fmt.Println("chainClient url:" + cfg.AnkrCfg.HttpsUrl + cfg.AnkrCfg.ApiKey)                // 打印链客户端 URL

	// 创建区块链客户端
	chainClient, err = chainclient.New(int(cfg.ChainCfg.ID), cfg.AnkrCfg.HttpsUrl+cfg.AnkrCfg.ApiKey)
	if err != nil {
		return nil, errors.Wrap(err, "failed on create evm client") // 创建失败返回错误
	}

	// 根据链 ID 初始化订单簿同步器
	switch cfg.ChainCfg.ID {
	case chain.EthChainID, chain.OptimismChainID, chain.SepoliaChainID:
		orderbookSyncer = orderbookindexer.New(ctx, cfg, db, kvStore, chainClient, cfg.ChainCfg.ID, cfg.ChainCfg.Name, orderManager)
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed on create trade info server") // 创建失败返回错误
	}
	// 构造 Service 实例
	manager := Service{
		ctx:              ctx,               // 上下文
		config:           cfg,               // 配置
		db:               db,                // 数据库
		kvStore:          kvStore,           // KV 存储
		collectionFilter: collectionFilter,  // 集合过滤器
		orderbookIndexer: orderbookSyncer,   // 订单簿同步器
		orderManager:     orderManager,      // 订单管理器
		wg:               &sync.WaitGroup{}, // 并发等待组
	}
	return &manager, nil // 返回实例
}

// Start 启动服务，预加载集合并启动订单簿和订单管理器
func (s *Service) Start() error {
	// 预加载集合过滤器（不要移动位置，依赖初始化顺序）
	if err := s.collectionFilter.PreloadCollections(); err != nil {
		return errors.Wrap(err, "failed on preload collection to filter") // 预加载失败返回错误
	}

	s.orderbookIndexer.Start() // 启动订单簿同步器
	s.orderManager.Start()     // 启动订单管理器
	return nil                 // 启动成功返回 nil
}
