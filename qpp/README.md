# 产品自动上传工具 (auto-upload-product)

## 项目简介

这是一个基于Go语言和Playwright实现的电商产品自动上传工具，主要用于自动化填写和上传产品信息到店小秘平台（Dianxiaomi）。

## 功能特性

- ✅ 自动登录店小秘平台
- ✅ 导航到产品编辑页面
- ✅ 自动填写产品信息（中文标题、英文标题）
- ✅ 上传产品素材图
- ✅ 清空和设置品牌图片（按颜色分类）
- ✅ 支持配置文件管理
- ✅ 浏览器自动化控制

## 技术栈

- **编程语言**：Go 1.24.2
- **浏览器自动化**：Playwright
- **配置文件**：YAML
- **目标平台**：店小秘（https://www.dianxiaomi.com）

## 项目结构

```
auto-upload-product/
├── main.go              # 主程序入口
├── config/
│   ├── config.go        # 配置加载模块
│   └── config.yaml      # 配置文件
├── handle/
│   ├── common.go        # 公共处理函数（登录）
│   └── product.go       # 产品信息处理（填写、上传图片）
├── go.mod               # Go模块定义
├── go.sum               # 依赖版本锁定
└── .gitignore           # Git忽略文件
```

## 快速开始

### 环境要求

- Go 1.24.2 或更高版本
- 网络连接（用于下载Playwright浏览器）

### 安装依赖

```bash
go mod download
```

### 配置说明

编辑 `config/config.yaml` 文件：

```yaml
# 浏览器配置
browser:
  headless: false              # 是否无头模式运行
  timeout: 30                  # 页面加载超时时间（秒）
  viewport:
    width: 1920                # 浏览器窗口宽度
    height: 1080               # 浏览器窗口高度

# 登录配置
login:
  url: "https://www.dianxiaomi.com/home.htm"  # 登录页面URL
  username: "17705911697"       # 用户名
  password: "Xzy7530."          # 密码
  wait_time: 10                 # 登录等待时间（秒）

# 存储配置
storage:
  state_path: "./storage/state.json"     # 登录状态保存路径
  screenshot_path: "./screenshots"       # 截图保存路径

# 自动化配置
automation:
  retry_count: 3               # 任务重试次数
  retry_delay: 2               # 重试间隔（秒）
  navigation_timeout: 10        # 导航超时时间（秒）
```

### 运行程序

```bash
go run main.go
```

## 使用流程

1. **启动程序**：程序会自动加载配置并启动浏览器
2. **自动填写**：自动填写用户名和密码
3. **手动登录：需要用户在浏览器中手动点击登录按钮**
4. **按回车继续**：登录成功后按回车键继续执行
5. **自动填写产品**：自动填写产品标题、英文标题和上传图片
6. **完成**：按回车键退出程序

## 核心模块

### 1. 配置模块 (config/config.go)

负责加载和管理配置文件，包含以下配置项：
- 浏览器配置（无头模式、超时、窗口大小）
- 登录配置（URL、用户名、密码）
- 存储配置（状态文件路径、截图路径）
- 自动化配置（重试次数、重试延迟）

### 2. 登录模块 (handle/common.go)

处理登录流程：
- 导航到登录页面
- 等待页面加载
- 自动填写用户名和密码

### 3. 产品模块 (handle/product.go)

处理产品信息填写和图片上传：
- `ToProductPage()` - 导航到产品编辑页面
- `FillProductInfo()` - 填写产品基本信息
- `uploadMaterialImage()` - 上传产品素材图
- `clearBianzhongImage()` - 清空品牌图片
- `setBianzhongImage()` - 设置品牌图片

## 扩展开发

### 添加新的产品字段

在 `handle/product.go` 中的 `Product` 结构体添加新字段：

```go
type Product struct {
    TitleCh     string   `json:"titleCh"`
    TitleEn     string   `json:"titleEn"`
    MaterialImg string   `json:"materialImg"`
    // 添加新字段
    Description string   `json:"description"`
}
```

### 添加新的上传功能

在 `handle/product.go` 中添加新的上传函数，参考现有的 `uploadMaterialImage()` 实现。

## 注意事项

1. **账号安全**：请勿将包含真实密码的配置文件提交到版本控制系统
2. **图片路径**：确保图片文件路径正确，程序会在 `image/` 目录下查找图片
3. **网络要求**：程序需要稳定的网络连接以访问店小秘平台
4. **手动操作**：部分操作（如登录确认）需要手动干预，程序会在必要时暂停等待

## 故障排除

### 常见问题

1. **浏览器启动失败**
   - 检查是否已安装Playwright浏览器：`playwright.Install()`
   - 确保网络连接正常

2. **元素定位失败**
   - 网站页面结构可能发生变化，需要更新选择器
   - 增加等待时间以确保页面完全加载

3. **图片上传失败**
   - 检查图片文件是否存在
   - 验证图片格式是否支持（JPG、PNG等）

### 调试建议

- 将 `config.yaml` 中的 `headless` 设置为 `false` 以查看浏览器操作
- 增加超时时间来处理网络延迟
- 使用截图功能保存页面状态

## 版本历史

- **v0.1.0** - 初始版本，支持基本的登录和产品信息填写功能

## 许可证

本项目仅供学习和研究使用。
