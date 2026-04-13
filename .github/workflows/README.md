# ============================================
# GitHub Actions 完整示例指南
# ============================================

# 一、基础结构
# ----------------
# workflow 文件放在 .github/workflows/ 目录下
# 文件名任意，后缀 .yml 或 .yaml

name: Workflow 名称          # 可选，显示在 Actions 页面
on:                         # 触发条件
  push:
    branches: [main]
permissions:                # 权限设置
  contents: read
jobs:                       # 任务列表
  job-name:
    runs-on: ubuntu-latest  # 运行环境
    steps:                  # 步骤列表
      - name: Step name
        run: echo "Hello"

# ============================================
# 二、触发条件详解
# ============================================

on:
  # 1. 推送触发
  push:
    branches: [main, develop]       # 特定分支
    paths: ['src/**', 'pkg/**']     # 特定路径变更才触发
    tags: ['v*']                    # 推送 tag 时

  # 2. PR 触发
  pull_request:
    branches: [main]

  # 3. 定时触发 (UTC 时间)
  schedule:
    - cron: '0 0 * * *'    # 每天 0:00 UTC
    - cron: '0 9 * * 1-5'  # 工作日 9:00 UTC

  # 4. 手动触发 (在 Actions 页面点击 Run)
  workflow_dispatch:
    inputs:
      version:
        description: 'Version to build'
        required: true
        default: '1.0.0'

  # 5. 其他 workflow 完成后触发
  workflow_call:
  workflow_run:
    workflows: ["CI"]
    types: [completed]

# ============================================
# 三、运行环境 (Runner)
# ============================================

# GitHub 提供的免费 Runner:
runs-on:
  - ubuntu-latest      # Linux x64
  - ubuntu-22.04       # 指定版本
  - macos-latest       # macOS (ARM)
  - macos-13           # macOS (Intel)
  - windows-latest     # Windows

# 自托管 Runner (私有服务器):
runs-on: self-hosted

# 矩阵构建 (多平台并行):
strategy:
  matrix:
    os: [ubuntu-latest, macos-latest, windows-latest]
    go: ['1.21', '1.22', '1.23']
runs-on: ${{ matrix.os }}

# ============================================
# 四、Steps 类型
# ============================================

steps:
  # 类型1: uses - 使用现成的 Action
  - name: Checkout code
    uses: actions/checkout@v4
    with:
      fetch-depth: 0    # 参数

  # 类型2: run - 执行命令
  - name: Build
    run: |
      go build -o app ./cmd/main
      echo "Build complete"

  # 类型3: 环境变量
  - name: Deploy
    run: deploy.sh
    env:
      API_KEY: ${{ secrets.API_KEY }}
      VERSION: ${{ github.ref_name }}

# ============================================
# 五、常用 Actions
# ============================================

# 代码检出
- uses: actions/checkout@v4

# Go 环境
- uses: actions/setup-go@v5
  with:
    go-version: '1.23'
    cache: true          # 启用模块缓存

# Node 环境
- uses: actions/setup-node@v4
  with:
    node-version: '20'
    cache: 'npm'

# Python 环境
- uses: actions/setup-python@v5
  with:
    python-version: '3.11'

# Java 环境
- uses: actions/setup-java@v4
  with:
    java-version: '17'
    distribution: 'temurin'

# Docker 构建
- uses: docker/build-push-action@v5
  with:
    context: .
    push: true
    tags: user/app:latest

# 缓存 (加速构建)
- uses: actions/cache@v4
  with:
    path: ~/.cache/go-build
    key: ${{ runner.os }}-go-${{ hashFiles('go.sum') }}

# 上传产物
- uses: actions/upload-artifact@v4
  with:
    name: build-output
    path: bin/
    retention-days: 30   # 保留天数

# 下载产物 (跨 job 共享)
- uses: actions/download-artifact@v4
  with:
    name: build-output

# 发布 Release
- uses: softprops/action-gh-release@v2
  with:
    files: bin/*
    draft: false
    prerelease: false

# ============================================
# 六、上下文变量
# ============================================

# github 上下文
${{ github.repository }}      # owner/repo
${{ github.ref }}             # refs/tags/v1.0.0
${{ github.ref_name }}        # v1.0.0
${{ github.sha }}             # commit hash
${{ github.event_name }}      # push, pull_request...
${{ github.actor }}           # 触发者用户名
${{ github.token }}           # GITHUB_TOKEN
${{ github.workflow }}        # workflow 名称
${{ github.run_id }}          # 运行 ID
${{ github.run_number }}      # 运行序号

# runner 上下文
${{ runner.os }}              # Linux, macOS, Windows
${{ runner.arch }}            # X64, ARM64

# job 上下文
${{ job.status }}             # success, failure, cancelled

# steps 上下文 (需要 step 有 id)
- id: build
  run: echo "version=1.0" >> $GITHUB_OUTPUT
- run: echo ${{ steps.build.outputs.version }}

# env 上下文
env:
  MY_VAR: value
${{ env.MY_VAR }}

# secrets 上下文 (敏感信息)
${{ secrets.API_KEY }}
${{ secrets.GITHUB_TOKEN }}   # 自动提供

# ============================================
# 七、条件执行
# ============================================

# 基于条件跳过 step
- if: github.ref == 'refs/heads/main'
  run: deploy.sh

# 基于条件跳过 job
jobs:
  deploy:
    if: github.event_name == 'push'

# 成功/失败时执行
- if: success()    # 前面步骤都成功
  run: echo "Success"
- if: failure()    # 有步骤失败
  run: echo "Failed"
- if: always()     # 总是执行
  run: cleanup.sh

# ============================================
# 八、权限设置
# ============================================

permissions:
  contents: read       # 读取代码
  contents: write      # 写入 (用于 release)
  packages: write      # 发布包
  issues: write        # 写 issues
  pull-requests: write # 写 PR

# ============================================
# 九、Job 依赖
# ============================================

jobs:
  build:
    runs-on: ubuntu-latest
    steps: [...]

  test:
    needs: build       # 依赖 build job
    runs-on: ubuntu-latest
    steps: [...]

  deploy:
    needs: [build, test]  # 依赖多个 job
    if: success()
    runs-on: ubuntu-latest
    steps: [...]

# ============================================
# 十、产物共享 (跨 Job)
# ============================================

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - run: go build -o app
      - uses: actions/upload-artifact@v4
        with:
          name: app-binary
          path: app

  deploy:
    needs: build
    runs-on: ubuntu-latest
    steps:
      - uses: actions/download-artifact@v4
        with:
          name: app-binary
      - run: ./app

# ============================================
# 十一、Secrets 管理
# ============================================

# 设置位置: 仓库 Settings → Secrets and variables → Actions

# 使用方式
env:
  API_KEY: ${{ secrets.API_KEY }}
  DATABASE_URL: ${{ secrets.DATABASE_URL }}

# 注意事项
# - Fork 的 PR 不能访问 secrets (安全考虑)
# - secrets 值在日志中会被隐藏
# - 最大 48KB 每个 secret

# ============================================
# 十二、常用技巧
# ============================================

# 1. 提取 tag 版本号
- id: version
  run: echo "VERSION=${GITHUB_REF#refs/tags/}" >> $GITHUB_OUTPUT
- run: echo ${{ steps.version.outputs.VERSION }}

# 2. 多行命令
- run: |
    echo "Line 1"
    echo "Line 2"
    echo "Line 3"

# 3. 条件输出
- run: |
    if [ "${{ matrix.os }}" = "windows" ]; then
      echo "EXT=.exe" >> $GITHUB_OUTPUT
    else
      echo "EXT=" >> $GITHUB_OUTPUT
    fi

# 4. 使用矩阵变量
- run: go build -o app-${{ matrix.os }}-${{ matrix.arch }}
  env:
    GOOS: ${{ matrix.os }}
    GOARCH: ${{ matrix.arch }}

# 5. 检查文件是否存在
- run: test -f app && echo "exists" || echo "not found"

# ============================================
# 十三、完整示例：Go 项目多平台发布
# ============================================

name: Release

on:
  push:
    tags: ['v*']

permissions:
  contents: write

jobs:
  build:
    strategy:
      matrix:
        include:
          - { os: linux, arch: amd64, runner: ubuntu-latest }
          - { os: linux, arch: arm64, runner: ubuntu-latest }
          - { os: darwin, arch: amd64, runner: macos-13 }
          - { os: darwin, arch: arm64, runner: macos-latest }
          - { os: windows, arch: amd64, runner: windows-latest }

    runs-on: ${{ matrix.runner }}

    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: '1.23'

      - name: Build
        env:
          GOOS: ${{ matrix.os }}
          GOARCH: ${{ matrix.arch }}
          CGO_ENABLED: 0
        run: |
          VERSION=${GITHUB_REF#refs/tags/}
          EXT=${{ matrix.os == 'windows' && '.exe' || '' }}
          go build -ldflags "-s -w -X main.Version=$VERSION" \
            -o email-${{ matrix.os }}-${{ matrix.arch }}$EXT \
            ./cmd/email

      - uses: actions/upload-artifact@v4
        with:
          name: email-${{ matrix.os }}-${{ matrix.arch }}
          path: email-${{ matrix.os }}-${{ matrix.arch }}*

  release:
    needs: build
    runs-on: ubuntu-latest

    steps:
      - uses: actions/download-artifact@v4
        with:
          path: releases
          pattern: email-*
          merge-multiple: true

      - uses: softprops/action-gh-release@v2
        with:
          files: releases/*