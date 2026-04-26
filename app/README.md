# 图片模板管理器

这是一个基于 Wails v2 的桌面应用，用于管理图片模板、素材，并支持自动化商品上传。

## 功能特性

- 选择图片文件并保存到本地
- 创建图片模板（包含名称、宽度、高度、缩放比例）
- 模板列表展示与管理（添加、删除）
- 素材列表管理
- 图片合成（Combine Image）
- 自动化上传（Playwright 浏览器自动化）
- 用户登录与权限管理
- 配置文件管理

## 快速开始

### 前置要求

- Go 1.21+
- Node.js 18+
- Wails CLI v2 (`go install github.com/wailsapp/wails/v2/cmd/wails@latest`)
- Task (`go install github.com/go-task/task/v3/cmd/task@latest`)（可选，推荐）

### 开发模式

```bash
# 启动开发服务器（支持热重载）
wails dev

# 或使用 task 命令（会自动清理占用的端口）
task dev
```

### 构建应用

```bash
# 构建当前平台的可执行文件
wails build -clean

# 构建并打开结果目录
wails build -o

# 使用 task 命令
task build
```

### 跨平台打包

```bash
# macOS
wails build -platform darwin/universal

# Windows (AMD64)
GOOS=windows GOARCH=amd64 wails build -clean

# Windows (ARM64)
GOOS=windows GOARCH=arm64 wails build

# Linux
GOOS=linux GOARCH=amd64 wails build
```

### 打包参数说明

| 参数 | 说明 | 示例 |
|------|------|------|
| `-platform` | 指定目标平台 | `darwin/universal`, `windows/amd64`, `linux/amd64` |
| `-o` | 构建完成后打开输出目录 | `wails build -o` |
| `-clean` | 构建前清理构建缓存 | `wails build -clean` |
| `-ldflags` | 传递额外的 linker flags | `wails build -ldflags "-s -w"` |
| `-tags` | 传递 Go build tags | `wails build -tags prod` |
| `-upx` | 使用 UPX 压缩可执行文件 | `wails build -upx` |
| `-upxflags` | 传递 UPX 压缩参数 | `wails build -upxflags "--best"` |
| `-webview2` | Windows 下 WebView2 安装策略 | `embed`, `download`, `browser` |
| `-nsis` | Windows 下生成 NSIS 安装包 | `wails build -nsis` |

### 开发模式参数

| 参数 | 说明 | 示例 |
|------|------|------|
| `-port` | 指定开发服务器端口 | `wails dev -port 9245` |
| `-loglevel` | 日志级别 | `wails dev -loglevel debug` |
| `-noreload` | 禁用前端热重载 | `wails dev -noreload` |
| `-frontenddevserverurl` | 使用外部前端开发服务器 | `wails dev -frontenddevserverurl http://localhost:5173` |

## 项目结构

```
app-image-handle/
├── main.go                 # 应用入口，服务注册
├── wails.json              # Wails 项目配置
├── Taskfile.yml            # Task 任务定义
├── services/               # Go 服务层
│   ├── imageservice.go     # 图片处理与模板管理
│   └── autouploadservice.go # 自动化上传（Playwright）
├── frontend/               # 前端代码
│   ├── src/
│   │   ├── App.vue         # 根组件
│   │   ├── router/         # Vue Router 配置
│   │   ├── api/            # Axios API 客户端
│   │   └── components/
│   │       ├── ImageManager.vue   # 图片管理页面
│   │       ├── TemplateList.vue   # 模板列表
│   │       ├── AddTemplate.vue    # 添加模板
│   │       ├── MaterialList.vue   # 素材列表
│   │       ├── CombineImage.vue   # 图片合成
│   │       ├── AutoUpload.vue     # 自动上传
│   │       ├── ConfigView.vue     # 配置管理
│   │       ├── LoginView.vue      # 登录页面
│   │       └── UserManage.vue     # 用户管理
│   └── package.json
├── images/                 # 保存的图片（运行时创建）
└── templates.json          # 模板数据（运行时创建）
```

## 使用说明

1. **选择图片**：点击"选择图片"按钮，从文件系统中选择图片文件
2. **保存图片**：点击"保存到当前文件夹"，图片会被复制到 `images/` 目录
3. **填写模板信息**：
   - 模板名称：给模板起一个名字
   - 宽度/高度：设置图片的目标尺寸（像素）
   - 缩放比例：设置缩放倍数
4. **添加模板**：点击"添加模板"按钮保存
5. **查看列表**：所有模板会在下方列表中展示
6. **删除模板**：点击模板右上角的"删除"按钮
7. **自动上传**：配置账号信息后，可使用 Playwright 自动上传商品

## 技术栈

- **后端**：Go + Wails v2
- **前端**：Vue 3 + TypeScript + Vite
- **UI 框架**：Element Plus + UnoCSS
- **路由**：Vue Router 4
- **自动化**：Playwright for Go

## 开发指南

### 添加新服务

1. 在 `services/` 目录创建新的 `.go` 文件
2. 使用 `package services`
3. 在 `main.go` 中实例化并注册到 `wails.App`
4. 运行 `wails dev` 自动生成 TypeScript 绑定

### 前端开发

```bash
cd frontend
npm install
npm run dev      # 独立前端开发（不带 Wails 绑定）
npm run build    # 生产构建
npm run preview  # 预览生产构建
```

## 注意事项

- 图片文件会保存在应用运行目录的 `images/` 文件夹
- 模板数据保存在应用运行目录的 `templates.json` 文件
- 支持的图片格式：JPG, JPEG, PNG, GIF, BMP, WEBP
- 自动上传功能需要预先安装 Playwright 浏览器

## 许可证

MIT
