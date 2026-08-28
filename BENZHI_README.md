# sP566_9833

基于 Go 实现的 HTTP Web 项目，一款门诊取药叫号服务，提供资料登记、状态流转与结果查询。

## Standard commands

```bash
go build ./...
go test -count=1 ./...
```

## Run

```bash
go run ./cmd/pharmacy
```

## Frontend

```bash
(cd web && npm run build)
```

## Docker validation

```bash
chmod +x build_benzhi_docker.sh
./build_benzhi_docker.sh my-go-task linux/arm64
./build_benzhi_docker.sh my-go-task linux/amd64
docker run -it my-go-task:latest
```

## Known initial failures

See `BUG_REPRO.md` for the exact command and output captured during packaging.
