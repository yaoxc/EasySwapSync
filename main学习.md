- [请给我分析出这个项目是如何运行起来的](#请给我分析出这个项目是如何运行起来的)
		- [1. 启动流程](#1-启动流程)
		- [2. 配置与资源加载](#2-配置与资源加载)
		- [3. 对象创建与依赖注入](#3-对象创建与依赖注入)
		- [4. 数据库与 Redis 连接](#4-数据库与-redis-连接)
		- [5. 启动常驻线程服务](#5-启动常驻线程服务)
		- [6. 扫链获取数据并保存到数据库](#6-扫链获取数据并保存到数据库)
			- [主要方法/代码参考](#主要方法代码参考)



# 请给我分析出这个项目是如何运行起来的

根据你的项目结构和代码分析，EasySwapSync 项目的运行流程大致如下：

---

### 1. 启动流程

- 入口在 main.go，调用 `cmd.Execute()`。
- 实际启动命令为：`go run main.go daemon`，见 README.md。

---

### 2. 配置与资源加载

- 配置文件为 `config/config.toml`，通过 `UnmarshalConfig` 方法加载（见 service/config/config.go）。
- 配置内容包括数据库、Redis、链节点等信息。

---

### 3. 对象创建与依赖注入

- 通过 `service.New(ctx, cfg)` 构造主服务对象（见 service/service.go）。
- 依赖注入流程：
  - Redis 配置转为 `kv.KvConf`，创建 `kvStore`（见 xkv.go）。
  - 数据库连接对象 `db` 通过 `model.NewDB(cfg.DB)` 创建。
  - 订单管理器 `orderManager`、链客户端 `chainClient`、订单簿同步器 `orderbookSyncer` 等依次初始化。

---

### 4. 数据库与 Redis 连接

- 数据库连接信息从配置文件读取，创建 MySQL 连接（见 docker-compose-arm64.yml）。
- Redis 连接同样从配置文件读取，创建 Redis 客户端。
- 相关代码见 `NewStore` 和 `NewDB` 方法。

---

### 5. 启动常驻线程服务

- 主服务对象调用 `Start()` 方法，启动各类同步、监听、处理线程（见 service/service.go）。
- 例如订单簿同步器、链事件监听器等会作为 goroutine 常驻运行。

---

### 6. 扫链获取数据并保存到数据库

- 通过链客户端（如 `chainclient.New`）连接区块链节点，监听合约事件。
- 事件数据由同步器（如 `orderbookindexer.New`）解析后，写入数据库和缓存（Redis）。
- 相关流程见 `orderbookSyncer`、`orderManager` 的初始化和方法调用。

---

#### 主要方法/代码参考

- 入口：main.go
- 配置加载：`UnmarshalConfig`（service/config/config.go）
- 服务构造：`New`（service/service.go）
- Redis/DB连接：`NewStore`、`NewDB`
- 常驻服务启动：`Start`（service/service.go）
- 扫链与数据保存：`orderbookindexer.New`、`chainclient.New`

---

如需具体某一部分的详细代码分析，可进一步指定模块。

