# 图片模板管理器

这是一个基于 Wails v3 的桌面应用，用于管理图片模板。

## 功能特性

- ✅ 选择图片文件
- ✅ 保存图片到当前文件夹的 `images/` 目录
- ✅ 创建图片模板（包含名称、宽度、高度、缩放比例）
- ✅ 模板列表展示
- ✅ 删除模板
- ✅ 模板数据持久化（保存到 `templates.json`）

## 快速开始

### 开发模式

```bash
# 启动开发服务器（支持热重载）
wails3 dev

# 或使用 task 命令
task dev
```

### 构建应用

```bash
# 构建当前平台的可执行文件
wails3 build

# 构建exe
GOOS=windows GOARCH=amd64 wails3 build

# 或使用 task 命令
task build
```

## 项目结构

```
app-image-handle/
├── main.go                 # 应用入口
├── services/               # 服务层
│   ├── greetservice.go    # 示例服务
│   └── imageservice.go    # 图片处理服务
├── frontend/              # 前端代码
│   ├── src/
│   │   ├── App.vue       # 根组件
│   │   └── components/
│   │       └── ImageManager.vue  # 图片管理页面
│   └── bindings/         # 自动生成的 TypeScript 绑定
├── images/               # 保存的图片（运行时创建）
└── templates.json        # 模板数据（运行时创建）
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

## 技术栈

- **后端**：Go 1.25 + Wails v3
- **前端**：Vue 3 + TypeScript + Vite
- **UI**：原生 CSS（响应式设计）

## 开发指南

### 添加新服务

1. 在 `services/` 目录创建新的 `.go` 文件
2. 使用 `package services`
3. 在 `main.go` 中注册服务
4. 运行 `wails3 dev` 自动生成 TypeScript 绑定

### 前端开发

```bash
cd frontend
npm install
npm run dev
```

## 注意事项

- 图片文件会保存在应用运行目录的 `images/` 文件夹
- 模板数据保存在应用运行目录的 `templates.json` 文件
- 支持的图片格式：JPG, JPEG, PNG, GIF, BMP, WEBP

## 许可证

MIT
