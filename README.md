# Goal-Web 框架

<p align="center">
  <img src="https://via.placeholder.com/200x200?text=Goal-Web" alt="Goal-Web Logo" width="200" height="200">
</p>

<p align="center">
  一个优雅、高性能的 Go Web 开发框架
</p>

<p align="center">
  <a href="https://github.com/goal-web/goal-web/actions"><img src="https://github.com/goal-web/goal-web/workflows/tests/badge.svg" alt="Build Status"></a>
  <a href="https://goreportcard.com/report/github.com/goal-web/goal-web"><img src="https://goreportcard.com/badge/github.com/goal-web/goal-web" alt="Go Report Card"></a>
  <a href="https://pkg.go.dev/github.com/goal-web/goal-web"><img src="https://pkg.go.dev/badge/github.com/goal-web/goal-web.svg" alt="PkgGoDev"></a>
  <a href="https://github.com/goal-web/goal-web/blob/master/LICENSE"><img src="https://img.shields.io/github/license/goal-web/goal-web" alt="License"></a>
</p>

## 简介

Goal-Web 是一个用 Go 语言编写的高性能、易于使用的 Web 应用开发框架，设计理念借鉴了 PHP 的 Laravel 框架，为 Go 开发者提供了优雅、简洁的 API。Goal-Web 致力于让 Web 开发变得更加愉悦，同时保持 Go 语言的高性能特性。

## 特性

- **模块化设计**：框架被分割成多个独立的模块，可以根据需要单独使用
- **依赖注入容器**：强大的服务容器，便于管理应用依赖
- **优雅的路由系统**：简洁而灵活的 HTTP 路由定义
- **中间件支持**：可组合的请求处理管道
- **数据库抽象层**：简化数据库操作的查询构建器和 ORM
- **认证与授权**：内置用户认证和权限控制系统
- **缓存系统**：支持多种缓存驱动
- **事件系统**：应用内的事件分发与处理
- **队列服务**：处理后台任务和延迟作业
- **命令行工具**：便捷的命令行接口，简化开发流程

## 安装

### 前提条件

- Go 1.23 或更高版本

### 通过 Goal-CLI 创建项目

```bash
# 安装 Goal-CLI 工具
go install github.com/goal-web/goal-cli@latest

# 创建新项目
goal new my-project

# 进入项目目录
cd my-project

# 启动开发服务器
go run main.go run
```

### 手动安装

```bash
# 创建项目目录
git clone https://github.com/goal-web/goal.git my-project

# 进入项目目录
cd my-project

# 启动开发服务器
go run main.go run
```

## 快速开始

### 创建简单的 HTTP 服务器

```go
package main

import (
    "github.com/goal-web/http"
    "github.com/goal-web/http/router"
)

func main() {
    // 创建路由
    r := router.New()
    
    // 定义路由
    r.Get("/", func() string {
        return "Hello, Goal-Web!"
    })
    
    r.Get("/users/{id}", func(id int) map[string]interface{} {
        return map[string]interface{}{
            "id":   id,
            "name": "Goal User",
        }
    })
    
    // 启动 HTTP 服务器
    http.Run(r, ":8080")
}
```

### 使用控制器

```go
package controllers

import (
    "github.com/goal-web/contracts"
    "github.com/goal-web/http"
)

type UserController struct {
    // 依赖注入
}

func (controller *UserController) Show(id int) contracts.Response {
    // 获取用户数据...
    user := map[string]interface{}{
        "id":   id,
        "name": "Goal User",
    }
    
    return http.Json(user)
}

// 在路由中使用控制器
// r.Get("/users/{id}", controllers.UserController{}.Show)
```

## 文档

详细文档请访问 [Goal-Web 文档](https://goal-web.github.io/docs)。

## 组件更新

### 更新指定组件
```bash
./update.sh -i goal-cli -m 'feat: 更新日志'
```

### 排除指定组件
```bash
./update.sh -e goal-cli -m 'feat: 更新日志'
```

### 指定 tag
```bash
./update.sh -t v1.0.1 -m 'feat: 更新日志'
```

## 贡献

我们欢迎并感谢任何形式的贡献。如果您发现了 bug 或有新功能的想法，请[提交 issue](https://github.com/goal-web/goal-web/issues/new) 或创建 Pull Request。

## 许可证

Goal-Web 框架是在 [MIT 许可证](LICENSE) 下发布的开源软件。