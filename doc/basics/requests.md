# 请求

## 访问请求

要通过依赖注入获取当前的 HTTP 请求实例，您应该在控制器方法或构造函数中类型提示 `contracts.Request` 接口。传入的请求实例将自动被服务容器注入：

```go
package controllers

import (
    "github.com/goal-web/contracts"
    "github.com/goal-web/http"
)

type UserController struct {
    request contracts.Request
}

func NewUserController(request contracts.Request) *UserController {
    return &UserController{request: request}
}

func (controller *UserController) Store() contracts.Response {
    name := controller.request.Input("name")
    
    // 存储用户...
    
    return http.Success(map[string]interface{}{
        "created": true,
    })
}
```

### 依赖注入和路由参数

如果您的控制器方法也需要从路由中获取参数，您应该在其他依赖项之后列出路由参数：

```go
func (controller *UserController) Show(id string) contracts.Response {
    // 获取用户...
    
    return http.Success(user)
}
```

## 请求路径和方法

`contracts.Request` 接口提供了各种方法来检查传入的 HTTP 请求。以下是一些最常用的方法：

### 获取请求路径

`Path` 方法返回请求的路径信息。例如，如果传入的请求针对 `http://example.com/foo/bar`，则 `Path` 方法将返回 `foo/bar`：

```go
path := request.Path()
```

`Url` 方法返回完整的 URL，包括查询字符串：

```go
url := request.Url() // 例如：/foo/bar?q=search
```

### 获取请求方法

`Method` 方法将返回 HTTP 请求的方法。这可以用来验证请求的 HTTP 动词：

```go
method := request.Method()

if method == "POST" {
    // 处理 POST 请求...
}
```

## 输入

### 获取所有输入数据

您可以使用 `All` 方法以 `map[string]interface{}` 的形式获取所有输入数据：

```go
data := request.All()
```

### 获取输入值

您可以使用 `Input` 方法从请求中获取用户输入，而不必担心使用了哪个 HTTP 动词。无论请求是 GET 还是 POST，`Input` 方法都可以使用：

```go
name := request.Input("name")
```

您可以传递一个默认值作为 `Input` 方法的第二个参数。如果请求中不存在请求的输入值，则返回此默认值：

```go
name := request.Input("name", "John Doe")
```

处理包含数组输入的表单时，可以使用点符号访问数组：

```go
names := request.Input("products.0.name")
```

### 确定输入值是否存在

您可以使用 `Has` 方法确定请求中是否存在值。如果该值存在且不为空，则 `Has` 方法返回 `true`：

```go
if request.Has("name") {
    // 处理...
}
```

### 获取部分输入数据

如果您需要获取输入数据的子集，可以使用 `Only` 和 `Except` 方法：

```go
// 只获取 name 和 email 字段
data := request.Only("name", "email")

// 获取除了 password 之外的所有字段
data := request.Except("password")
```

## 文件

### 获取上传的文件

您可以使用 `File` 方法从请求中获取上传的文件。`File` 方法返回一个实现了 `contracts.UploadedFile` 接口的对象：

```go
file := request.File("photo")

if file != nil {
    // 处理文件...
    path := file.Store("uploads")
}
```

### 确定文件是否存在

您可以使用 `HasFile` 方法确定请求中是否存在文件：

```go
if request.HasFile("photo") {
    // 处理文件...
}
```

### 验证成功上传

除了检查文件是否存在外，您还可以通过 `IsValid` 方法验证上传过程中是否没有问题：

```go
if request.HasFile("photo") && request.File("photo").IsValid() {
    // 处理文件...
}
```

### 存储上传的文件

要存储上传的文件，通常使用 `Store` 方法。此方法将文件移动到您指定的位置：

```go
if request.HasFile("photo") {
    path := request.File("photo").Store("uploads")
}
```

您也可以使用 `StoreAs` 方法指定文件名：

```go
path := request.File("photo").StoreAs("uploads", "custom-name.jpg")
```

## 配置可信代理

当您的应用程序运行在负载均衡器后面，该负载均衡器终止 TLS / SSL 证书时，您可能会注意到您的应用程序有时不会生成 HTTPS 链接。这通常是因为应用程序正在从负载均衡器转发的端口 80 上的流量中获取请求，并且不知道它应该生成安全链接。

要解决这个问题，您可以在 `app/http/kernel.go` 文件中设置可信代理：

```go
package http

import (
    "github.com/goal-web/http"
)

type Kernel struct {
    http.Kernel
}

func NewKernel() *Kernel {
    kernel := &Kernel{}
    
    // 设置可信代理
    kernel.TrustedProxies = []string{"10.0.0.1", "10.0.0.2"}
    
    return kernel
}
```

## 请求生命周期

Goal-Web 应用程序的入口点是 `public/index.go` 文件。所有请求都通过此文件进入您的应用程序，并由框架启动。了解请求如何通过您的应用程序是开始理解框架的重要一步。

1. 首先，请求进入 `public/index.go` 文件，该文件加载 Composer 生成的自动加载器，并从 `bootstrap/app.go` 获取 Goal-Web 应用程序的实例。

2. 来自应用程序的 HTTP 内核的实例被获取。内核充当所有请求的中心位置，并负责请求的实际处理。

3. HTTP 内核加载服务提供者。

4. 请求被发送到 HTTP 内核的 `Handle` 方法。

5. 内核引导应用程序并准备处理请求。

6. 内核将请求通过一系列中间件，这些中间件可以执行各种任务，如读取和写入会话、确定应用程序是否处于维护模式等。

7. 如果应用程序处于维护模式，中间件将向客户端返回一个维护模式响应。否则，请求将继续通过中间件堆栈。

8. 一旦请求通过所有中间件，路由器将请求分派到适当的控制器方法或闭包。

9. 控制器或闭包处理请求并返回一个响应。

10. 响应通过中间件堆栈返回，然后发送回客户端。

## 下一步

现在您已经了解了 Goal-Web 的请求处理基础，您可以继续阅读以下文档：

- [响应](responses.md)
- [视图](views.md)
- [验证](validation.md)
- [错误处理](errors.md)