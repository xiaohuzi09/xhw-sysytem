# Services

此目录用于存放所有的 Go 服务文件。

## 如何创建新服务

1. 在此目录下创建新的 `.go` 文件
2. 使用 `package services`
3. 定义服务结构体和方法（方法必须是导出的，即首字母大写）
4. 在 `main.go` 中注册服务

### 示例

```go
package services

type ImageService struct{}

func (s *ImageService) ProcessImage(path string) string {
    // 处理图片逻辑
    return "处理完成"
}
```

然后在 `main.go` 中注册：

```go
Services: []application.Service{
    application.NewService(&services.GreetService{}),
    application.NewService(&services.ImageService{}),
},
```

## 注意事项

- 所有服务方法必须是导出的（首字母大写）
- 服务方法会自动生成 TypeScript 绑定到 `frontend/bindings/` 目录
- 运行 `wails3 dev` 后会自动生成绑定文件
