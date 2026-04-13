.PHONY: build run test clean install build-all build-linux build-darwin build-windows

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

# 多平台构建目标
PLATFORMS := linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64

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
	@echo "  build-all    - Build for all platforms (linux/darwin/windows)"
	@echo "  build-linux  - Build for Linux (amd64/arm64)"
	@echo "  build-darwin - Build for macOS (amd64/arm64)"
	@echo "  build-windows- Build for Windows (amd64)"
	@echo "  run          - Build and run the application"
	@echo "  test         - Run tests"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  clean        - Clean build artifacts"
	@echo "  deps         - Download dependencies"
	@echo "  tidy         - Tidy dependencies"
	@echo "  install      - Install to system"
	@echo "  dev          - Development build"

# 多平台构建 - 全平台
build-all:
	@echo "Building for all platforms..."
	@mkdir -p $(BUILD_DIR)/releases
	@for platform in $(PLATFORMS); do \
		GOOS=$$(echo $$platform | cut -d'/' -f1); \
		GOARCH=$$(echo $$platform | cut -d'/' -f2); \
		output=$(BUILD_DIR)/releases/$(PROJECT_NAME)-$$GOOS-$$GOARCH; \
		if [ "$$GOOS" = "windows" ]; then output=$$output.exe; fi; \
		echo "Building $$platform..."; \
		GOOS=$$GOOS GOARCH=$$GOARCH $(GOBUILD) $(LDFLAGS) -o $$output $(MAIN_PATH); \
	done
	@echo "All builds complete in $(BUILD_DIR)/releases/"

# Linux 构建
build-linux:
	@echo "Building for Linux..."
	@mkdir -p $(BUILD_DIR)/releases
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/releases/$(PROJECT_NAME)-linux-amd64 $(MAIN_PATH)
	GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/releases/$(PROJECT_NAME)-linux-arm64 $(MAIN_PATH)
	@echo "Linux builds complete"

# macOS 构建
build-darwin:
	@echo "Building for macOS..."
	@mkdir -p $(BUILD_DIR)/releases
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/releases/$(PROJECT_NAME)-darwin-amd64 $(MAIN_PATH)
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/releases/$(PROJECT_NAME)-darwin-arm64 $(MAIN_PATH)
	@echo "macOS builds complete"

# Windows 构建
build-windows:
	@echo "Building for Windows..."
	@mkdir -p $(BUILD_DIR)/releases
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/releases/$(PROJECT_NAME)-windows-amd64.exe $(MAIN_PATH)
	@echo "Windows build complete"