.PHONY: build run test clean install

# 项目名称
PROJECT_NAME := email
VERSION := 1.0.0
BUILD_DIR := bin

# Go 相关
GOCMD := go
GOBUILD := $(GOCMD) build
GOCLEAN := $(GOCMD) clean
GOTEST := $(GOCMD) test
GOGET := $(GOCMD) get
GOMOD := $(GOCMD) mod

# 主程序入口
MAIN_PATH := ./cmd/email

# 构建标志
LDFLAGS := -ldflags "-s -w -X main.Version=$(VERSION)"

# 默认目标
all: clean build

# 构建
build:
	@echo "Building..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(PROJECT_NAME) $(MAIN_PATH)
	@echo "Build complete: $(BUILD_DIR)/$(PROJECT_NAME)"

# 运行
run: build
	@echo "Running..."
	@./$(BUILD_DIR)/$(PROJECT_NAME)

# 测试
test:
	@echo "Testing..."
	$(GOTEST) -v ./...

# 测试覆盖率
test-coverage:
	@echo "Testing with coverage..."
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

# 清理
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html

# 安装依赖
deps:
	@echo "Installing dependencies..."
	$(GOMOD) download

# 整理依赖
tidy:
	@echo "Tidying dependencies..."
	$(GOMOD) tidy

# 安装到系统
install: build
	@echo "Installing..."
	@cp $(BUILD_DIR)/$(PROJECT_NAME) /usr/local/bin/
	@echo "Installed to /usr/local/bin/$(PROJECT_NAME)"

# 开发模式（热构建）
dev:
	@echo "Development mode..."
	@$(GOBUILD) -o $(BUILD_DIR)/$(PROJECT_NAME) $(MAIN_PATH)

# 帮助
help:
	@echo "Available targets:"
	@echo "  build        - Build the application"
	@echo "  run          - Build and run the application"
	@echo "  test         - Run tests"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  clean        - Clean build artifacts"
	@echo "  deps         - Download dependencies"
	@echo "  tidy         - Tidy dependencies"
	@echo "  install      - Install to system"
	@echo "  dev          - Development build"