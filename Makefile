.PHONY: swagger build run clean

# 安装swag工具
install-swagger:
	go install github.com/swaggo/swag/cmd/swag@latest

# 生成Swagger文档
swagger:
	swag init -g main.go -o docs

# 构建应用
build:
	go build -o social .

# 运行应用
run:
	go run main.go

# 清理生成的文件
clean:
	rm -f social
	rm -rf docs

# 安装依赖
deps:
	go mod tidy
	go mod download

# 运行测试
test:
	go test ./...

# 格式化代码
fmt:
	go fmt ./...

# 代码检查
lint:
	golangci-lint run

# 配置管理
config-list:
	go run scripts/config-helper.go list

config-add:
	@if [ -z "$(SERVER)" ]; then \
		echo "Usage: make config-add SERVER=<server_name>"; \
		exit 1; \
	fi
	go run scripts/config-helper.go add $(SERVER)

config-remove:
	@if [ -z "$(SERVER)" ]; then \
		echo "Usage: make config-remove SERVER=<server_name>"; \
		exit 1; \
	fi
	go run scripts/config-helper.go remove $(SERVER)

config-show:
	@if [ -z "$(SERVER)" ]; then \
		echo "Usage: make config-show SERVER=<server_name>"; \
		exit 1; \
	fi
	go run scripts/config-helper.go show $(SERVER)

config-validate:
	go run scripts/config-helper.go validate

# 完整构建流程
all: deps swagger build

# 部署测试环境
deploy:
	GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o social main.go
	upx social

	ssh root@13.58.131.189 "rm /var/www/social/social"
	scp social root@13.58.131.189:/var/www/social
	scp ./static/* root@13.58.131.189:/var/www/social/static
	scp ./config.dev.yaml root@13.58.131.189:/var/www/social/config.dev.yaml
	ssh root@13.58.131.189 "supervisorctl restart social"
	rm social


