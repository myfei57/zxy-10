# EdgeLog

EdgeLog 是分布式日志采集与转发平台：采集器按配置尾随边缘节点的日志文件，按行读取并
维护检查点游标，读取内容进入环形缓冲并按批次落盘，批次按路由规则分发到不同接收端，
失败按预算退避重试、超限进入死信队列，管道配置支持热更新，配额控制单源速率，全部操作
留审计。

## 运行

```bash
go build -mod=vendor -o edgelog ./cmd/edgelog
./edgelog -addr :8080 -data ./data
```

打开 http://localhost:8080/ 查看管道监控，/ui/config、/ui/rules、/ui/deadletter
分别查看采集配置、路由规则与死信队列页面。

## 测试

```bash
go test -mod=vendor ./...
go vet -mod=vendor ./...
```

## Docker

```bash
bash build_benzhi_docker.sh edgelog linux/amd64
docker run --rm -p 8080:8080 edgelog bash -c 'go run ./cmd/edgelog -addr :8080 -data /tmp/data'
```
