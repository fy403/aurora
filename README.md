This README is just a fast quick start document. You can find more detailed documentation at Aurora.

## Why is Aurora?
任务队列是异步处理的，因为它们往往不是由一个处理请求组成，或因它们是长时间运行的任务和非即时任务，单靠使用同步处理反而降低整体响应的性能。例如：登录请求发出后，系统需要为其推送消息通知，订阅更新信息，同步处理的失败会阻碍登录的成功；另一个例子是在请求期间，像数据库查询，数据加工等等花费，延长了请求-响应周期。相反，查询可以在后台异步执行，降低整体风险。当然，还有一种解决方案，当一个进来的请求需要这些处理，查询将获取预先计算的结果，而不是重新执行更长的查询。但这种缓存机制过于死板，不适合应对动态任务请求，因此就有了开发异步任务队列的需求！
任务队列的其他类型的任务包括但不限于如下：
1. 数据同步:包括业务服务之间的业务数据和日志信息同步等等；
2. 异步处理:包括发送IM消息、异步日志、异步短信/邮件或注册/开启任务；
3. 信息收集:主要用于统计、监控、搜索、爬虫、抢票等等；
4. 服务解耦:主要用于重构和新设计时，对频繁变动的接口服务进行解耦；
5. 分布式事务消息:尤其是对数据—致性有要求的异步处理场景；
6. 削峰:秒杀、分流、限流。
7. 异步任务解耦：将程序中涉及的异步任务独立出来处理，消除组件依赖。主要用途：数据的预处理（元数据统计，分析），后处理（存储，收集，分发）
8. 后台高扩展性：任务队列处理器组件可根据用户的需要精确扩展，亦可以按需动态伸缩。程序无需过多关注任务算力消耗。
9. 无服务器化：主应用仅需保留最基本的交互接口，而将大多数数据处理任务外包给任务队列。理论上，任务队列模型可以支撑起一个完整的程序架构。



### 未来计划

- [x] 链路追踪
- [x] 日志记录
- [x] 健康检查
- [x] 系统解耦
- [x] 独立配置
- [x] 前端管理
- [x] 函数计算
- [x] 动态句柄 
- [x] 分布式锁
- [x] 分布式缓存 
- [x] 监控支持
  

[体验地址](暂时未开源)
### 架构模式
![总体架构](images/framework.png)
任务队列由以下模块组成：
1. Gateway：负责请求的准入，校验，授权；[前端项目](https://github.com/fy403/aurora_frontend)
2. 代理Broker：负责消息的存储，传递；
3. 后台Backend：负责持久化存储整个系统的元数据和任务；
4. Center: 实现worker和gateway所公共的功能，例如连接broker，mongodb
5. Worker：负责存储和调用函数句柄处理任务。
### 关键词
* `Worker`：负责实现任务消费，处理任务，并更新状态到数据库；
* `Handler`：具体的任务处理模型，在model目录中定义；
### 核心部分
**图任务**
![任务模型](./images/workflow.png)

Aurora模型基础架构基于[RichardKnop/machinery](https://github.com/RichardKnopmachinery); 在此基础上，抽象出图任务模型，该模型的普适性更强，更能代表workFlow的设计模式。举例来说：本图模型是一系列任务的**前驱-后继**关系组合的任务集合，通过使用**拓扑排序**得到一个有序序列，并依照有序序列依次执行任务，例如：如上所示，Task-0与Task-1并行执行，随后Task-2执行，最后Task-3，Task-4并行执行。

### 配置文件案例，配置依赖中间件为RabbitMQ，Redis，MongoDB
```toml
# 公共配置区域
Region = "China"
Help = "Aurora"
[HTTP]
#            center is :4332
#健康检查端口 worker is :8080
Port = ""
[Auth]
NewAuthenticationKey = "xxxxxxxxxxxxxxxxxxx"
NewEncryptionKey     = "xxxxxxxxxxxxxxxxxxx"
OldAuthenticationKey = ""
OldEncryptionKey     = ""
Users = { admin = "123456", admin2 = "123456" }
[Auth.DefaultSessionOption]
	Path =     "/"
	MaxAge =   604800 #  a week
	HttpOnly = true
[Files]
Metrics = "./config/metrics.toml"
Log = "./config/log.toml"

[[Faas]]
Driver = "aliyunfc"
Endpoint = "xxxxxxxxxxxxxxxxxxx.cn-hangzhou.fc.aliyuncs.com"
ServiceName = "test" 
AccessKeyId = "xxxxxxxxxxxxxxxxxxx"
AccessKeySecret = "xxxxxxxxxxxxxxxxxxx"
[[Faas]]
Driver = "openfaas"
Endpoint = "http://localhost:8080"
Name = "admin"
Password = "xxxxxxxxxxxxxxxxxxx"

[Opentracing]
ServiceName = "" # defalut ServiceName is "aurora_" + componentName
CollectorEndpoint = "http://localhost:14268/api/traces"
logSpans        =   true
# 公共配置区域END
# ---------------------------------------------------------------------------
# Gateway配置区域
[Gateway]
Broker = "amqp://guest:guest@localhost:5672/"
BrokerApi = "http://guest:guest@localhost:15672/api/queues/%2F/" # + queueName
DefaultQueue = "aurora_tasks" # Default Message Publish Queue, Required
ResultBackend = "mongodb://admin:admin@localhost:27017/?authSource=admin"
Lock = "redis"
Cache = "redis"
ResultsExpireIn = 3600
	[Gateway.Web]
		WebAddr   = ":4332"            # web 地址
		WebIndex  = "./web" 		   # web html目录
		StaticFS  = true               # 开启静态资源浏览，路径 /static。
		FilePath  = "./web/static"     # 文件存放目录
	[Gateway.AMQP]
		Exchange = "aurora_exchange"
		ExchangeType = "direct"
		BindingKey = "aurora_tasks"
		PrefetchCount = 3
	[Gateway.MongoDB]
		Database = "aurora"
	[Gateway.Redis]
		DB = 0
		Password = "123456"
		Addr = ["localhost:6379"]
		Retries = -1 # 数则无限制
# Gateway配置区域END
# ---------------------------------------------------------------------------
# Worker配置区域
[Worker]
Broker = "amqp://guest:guest@localhost:5672/"
BrokerApi = "http://guest:guest@localhost:15672/api/queues/%2F/" # + queueName
ResultBackend = "mongodb://admin:admin@localhost:27017/?authSource=admin"
Lock = "redis"
Cache = "redis"
ResultsExpireIn = 3600
ConsumerTag = "" 
Concurrency = 0  # worker concurrency number:0 is not limit
Queue = "aurora_tasks-" # Message Subscribe Queue, Required
Labels = { k1 = "v1", k2 = "v2"} # Used by labelSelector
	[Worker.AMQP]
		Exchange = "aurora_exchange"
		ExchangeType = "direct"
		BindingKey = "aurora_tasks"
		PrefetchCount = 3
	[Worker.MongoDB]
		Database = "aurora"
	[Worker.Redis]
		DB = 0
		Password = "123456"
		Addr = ["localhost:6379"]
		Retries = -1 # 重试加锁无限制
# Worker配置区域END
```
## Building Aurora
```shell
# 方式一：自行编译
go build . -o aurora
# 方式二：全局安装
go install .
```
## Running Aurora
```shell
# 运行Web实例
# 确保config.toml已经替换为自己环境的配置，尤其是broker和backend
aurora web --config=./config/config.toml
# 运行worker实例
aurora worker --config=./config/config.toml
```
## Playing with Aurora
由于aurora依赖于mongoDB，以及RabbitMQ；因此，请确保提前运行依赖并在配置文件中显式给出。
启动完web以及worker后，即可在提示的前端中配置图任务流程。
### 前端图编辑
通过该前端，可以快速，高效定义一个任务模型。编辑好图后，点击格式导出并复制，该格式用于构建请求。
![图编辑](./images/frontend.png)

### API请求指引
1. 构建登录请求
```shell
response=$(curl --location --request POST 'http://localhost:4332/auth/login' \
--header 'Content-Type: application/json' \
--data-raw '{
    "username": "admin",
    "password": "123456"
}')

# 提取token值
token=$(echo "$response" | jq -r '.data.token')

# 将token值存储到环境变量中
export token=$token
```
>返回结果结构
```json
{
    "code": 200,
    "message": "",
    "data": {
        "user": {"roles":["admin"],"user":{"createTime":"time.Date(2023, time.April, 6, 3, 52, 5, 919008400, time.UTC)","enabled":true,"nickName":"管理员","username":"admin"}},
        "token": "aurora_session=MTY4MDc1MzEyNXwyNFNfUHBzUS1YYmJEVjJuMVI4MVN0WVQzcFJpQWFQOXdPa1h6ZDV2QVM3Ympha0o0X0NjbGhyLWxvN3g3UnBJUW5ETHBINThLaFlpSndPZlhoSjYtVURrQXZ3VmJLTEFCSkF1eGdsYmh5X2pVYVAwaDlfemVpcGw1OE1FM04xTXB0ZElhYmlJcUctbDJreTJQYk9YQmZ5WXVtUGFWMjYtQVZ3MDlEaGwwMlFZfFwZyam1JIPW4qVHBY0uhre8yW9qFMZKzM9noxLcmZTk; Path=/; Expires=Thu, 13 Apr 2023 03:52:05 GMT; Max-Age=604800; HttpOnly"
    }
}
```
2. 使用导出格式发送任务请求(将导出格式附加到body里面)
```shell
curl --location --request POST 'http://localhost:4332/api/task/send' \
--header "Authorization: ${token}" \
--header 'Content-Type: application/json' \
--data-raw '{"task_type":"graph","timestamp":1680752792,"signatures":[{"id":"y4e6vr63qb","name":"add","args":[{"type":"int64","value":1},{"type":"int64","value":2}],"label_selector":{"k1":"v1","k2":"v2"}},{"id":"ql0iytfa36","name":"add","args":[{"type":"int64","value":3},{"type":"int64","value":4}],"label_selector":{"k1":"v1","k2":"v2"}}],"relations":[{"0":1}],"timeout_duration":0,"sleep_duration":2000,"send_concurrency":0}'
```
>返回结果结构
```json
{
    "code": 200,
    "message": "",
    "data": {
        "uuid": "82814928-2de6-49c6-97fd-8cfaec0d6df9",
        "user": "admin",
        "batch_id": "db7358d7-da35-4f2c-aa96-f78436a0a682",
        "timestamp": 1680753494,
        "task_type": "graph",
        "task_uuid": "graph_ddc1f042-d4eb-4a81-812a-b0e477619cc0",
        "task_responses": [
            {
                "results": [
                    3
                ],
                "state": "SUCCESS",
                "signatures": [
                    {
                        "id": "y4e6vr63qb",
                        "uuid": "task_f29914e3-dd27-4a14-9caa-9843b566e9b9",
                        "name": "add",
                        "routing_key": "aurora_tasks-2660244723",
                        "label_selector": {
                            "k1": "v1",
                            "k2": "v2"
                        },
                        "eta": null,
                        "group_uuid": "",
                        "group_task_count": 0,
                        "graph_uuid": "graph_ddc1f042-d4eb-4a81-812a-b0e477619cc0",
                        "graph_task_count": 2,
                        "args": [
                            {
                                "type": "int64",
                                "value": 1
                            },
                            {
                                "type": "int64",
                                "value": 2
                            }
                        ],
                        "headers": null,
                        "priority": 0,
                        "immutable": false,
                        "retry_count": 0,
                        "retry_timeout": 0,
                        "on_success": null,
                        "on_error": null,
                        "chord_callback": null,
                        "broker_message_group_id": "",
                        "sqs_receipt_handle": "",
                        "stop_task_deletion_on_error": false,
                        "ignore_when_task_not_registered": false
                    }
                ],
                "callBack": null
            },
            {"results":[7],"state":"SUCCESS","signatures":[{"id":"ql0iytfa36","uuid":"task_cceedffe-7e6b-4a0f-bcd0-362e885ed359","name":"add","routing_key":"aurora_tasks-2660244723","label_selector":{"k1":"v1","k2":"v2"},"eta":null,"group_uuid":"","group_task_count":0,"graph_uuid":"graph_ddc1f042-d4eb-4a81-812a-b0e477619cc0","graph_task_count":2,"args":[{"type":"int64","value":3},{"type":"int64","value":4}],"headers":null,"priority":0,"immutable":false,"retry_count":0,"retry_timeout":0,"on_success":null,"on_error":null,"chord_callback":null,"broker_message_group_id":"","sqs_receipt_handle":"","stop_task_deletion_on_error":false,"ignore_when_task_not_registered":false}],"callBack":null}
        ]
    }
}
```
3. 发送查询请求
```shell
curl --location --request POST 'http://localhost:4332/api/task/simple/touch' \
--header "Authorization: ${token}" \
--header 'Content-Type: application/json' \
--data-raw '{
    "task_type": "graph",
    "task_uuid": "graph_ddc1f042-d4eb-4a81-812a-b0e477619cc0"
}'
```
>返回结果结构
```json
{
    "code": 200,
    "message": "",
    "data": [
        {
            "TaskUUID": "task_f29914e3-dd27-4a14-9caa-9843b566e9b9",
            "TaskName": "add",
            "State": "SUCCESS",
            "Results": [
                {
                    "Type": "int64",
                    "Value": 3
                }
            ],
            "Error": "",
            "CreatedAt": "2023-04-06T03:58:14.104Z",
            "TTL": 0
        },
        {
            "TaskUUID": "task_cceedffe-7e6b-4a0f-bcd0-362e885ed359",
            "TaskName": "add",
            "State": "SUCCESS",
            "Results": [
                {
                    "Type": "int64",
                    "Value": 7
                }
            ],
            "Error": "",
            "CreatedAt": "2023-04-06T03:58:14.115Z",
            "TTL": 0
        }
    ]
}
```
## Source code layout
```txt
├───client # 专属客户端实例
│   └───go # golang的客户端脚本
|———bin    # 存放动态句柄，注：此目录会持续监控 
├───cmd    # 使用github.com/spf13/cobra定义的cmd命令实例
├───config # 启动配置文件模板
├───draw   # 架构指南
├───images # MD资源
├───web    # 前端编译入口文件
└───internal        # 内部所有代码
    ├───api         # Gateway的api接口定义
    ├───app         # 暂未使用
    ├───auth        # 授权模块
    ├───backends    # 后端数据存储
    ├───brokers     # 消息队列
    ├───center      # 一个请求实例，请求broker和backend
    ├───common      # 通用基础类
    ├───config      # 配置实例定义
    |———cache       # 分布式缓存
    ├───locks       # 分布式锁
    |———log         # 日志组件
    |———metrics     # 监控组件
    ├───model       # 所有句柄模型
    ├───faas        # faas函数计算驱动模块
    ├───opentracing # 链路追踪
    ├───request     # 请求实例定义
    ├───retry       # 重试模块
    |———repo        # 文件存储仓库接口
    ├───tasks       # 任务处理反射模型
    └───worker      # 任务处理模型
````

### 动态句柄
为了支持各种语言的处理，你可以将其他语言编译成二进制可执行文件并将其放在./bin目录下：
worker会监控并扫描当前目录并向上注册，以供使用（在前端图编辑可以查看）；当然，需要遵循一定的规范：必须以标志输入读取参数，以标准输出结构。
以C++为例：
```cpp
#include<iostream>
#include<string>
using namespace std;
int main(){
    string input="";
    cin>>input;
    # do something you like
    string output="exec ok";
    cout<<output;
    return 0;
}
```

### 链式追踪,精准定位
由于架构的设计方式，使得分布式定位问题变得困难；因此引入了opentracing追踪模块，记录每一次任务处理过程和细节。
![链式追踪](./images/opentracing.png)

### 性能监控
使用了Prometheus作为监控组件，具体用法是在config/metrics.toml中添加指标（有格式要求）；在项目中的用法具体为：
> metrics.toml
```toml
# Index 每个指标唯一标识索引
# Type:1,2,3,4; 分别对应prometheus统计类型: counter,gauge,histogram,summary
[Metrics]
  Namespace = "fy403"
  Subsystem = "aurora"
  [[Metrics.Datas]]
    Index = 0
    Type = 1
    Name = "test_event"
    Help = "test测试次数"
    Labels = "k1-k2"
```
> main.go
```go
index := 0
err := metrics.IncCounterWithLabelValues(index, "v1", "v2")
```