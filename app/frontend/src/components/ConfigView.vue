<script setup lang="ts">
import { computed } from "vue";
import {
  Connection,
  OfficeBuilding,
  InfoFilled,
  Monitor,
} from "@element-plus/icons-vue";

interface ConfigItem {
  label: string;
  value: string;
  desc: string;
  icon: typeof Connection;
  type: "primary" | "success" | "warning" | "info";
}

const configItems = computed<ConfigItem[]>(() => [
  {
    label: "运行环境",
    value:
      import.meta.env.VITE_APP_ENV === "production" ? "生产环境" : "开发环境",
    desc: "VITE_APP_ENV",
    icon: OfficeBuilding,
    type: import.meta.env.VITE_APP_ENV === "production" ? "success" : "info",
  },
  {
    label: "API 基础地址",
    value: import.meta.env.VITE_API_BASE_URL || "/api",
    desc: "VITE_API_BASE_URL",
    icon: Connection,
    type: "primary",
  },
  {
    label: "调试模式",
    value: import.meta.env.VITE_DEBUG === "true" ? "已开启" : "已关闭",
    desc: "VITE_DEBUG",
    icon: Monitor,
    type: import.meta.env.VITE_DEBUG === "true" ? "warning" : "info",
  },
  {
    label: "应用标题",
    value: "图片模板管理器",
    desc: "App Title",
    icon: InfoFilled,
    type: "info",
  },
]);
</script>

<template>
  <div class="config-view">
    <div class="page-header">
      <h2>系统配置</h2>
      <p class="subtitle">当前运行环境与接口配置信息</p>
    </div>

    <div class="config-list">
      <div v-for="item in configItems" :key="item.label" class="config-card">
        <div class="config-icon" :class="`type-${item.type}`">
          <el-icon :size="20">
            <component :is="item.icon" />
          </el-icon>
        </div>
        <div class="config-body">
          <div class="config-meta">
            <span class="config-label">{{ item.label }}</span>
            <el-tag :type="item.type" size="small" effect="plain">
              {{ item.desc }}
            </el-tag>
          </div>
          <div class="config-value">{{ item.value }}</div>
        </div>
      </div>
    </div>

    <div class="config-hint">
      <el-icon><InfoFilled /></el-icon>
      <span>配置在构建时注入，如需修改请调整对应的 .env 文件后重新打包</span>
    </div>
  </div>
</template>

<style scoped>
.config-view {
  padding: 24px 28px;
}

.page-header {
  margin-bottom: 24px;
}

.page-header h2 {
  margin: 0;
  font-size: 22px;
  font-weight: 800;
  color: #0f172a;
  letter-spacing: -0.02em;
}

.subtitle {
  margin: 6px 0 0;
  color: #94a3b8;
  font-size: 14px;
}

.config-list {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.config-card {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 18px 20px;
  background: #fff;
  border: 1px solid rgba(148, 163, 184, 0.18);
  border-radius: 18px;
  transition: all 0.3s;
}

.config-card:hover {
  border-color: rgba(99, 102, 241, 0.2);
  box-shadow: 0 8px 24px rgba(15, 23, 42, 0.06);
}

.config-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 46px;
  height: 46px;
  border-radius: 14px;
  color: #fff;
  flex-shrink: 0;
}

.config-icon.type-primary {
  background: linear-gradient(135deg, #2563eb, #4f46e5);
}
.config-icon.type-success {
  background: linear-gradient(135deg, #059669, #10b981);
}
.config-icon.type-warning {
  background: linear-gradient(135deg, #d97706, #f59e0b);
}
.config-icon.type-info {
  background: linear-gradient(135deg, #475569, #64748b);
}

.config-body {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.config-meta {
  display: flex;
  align-items: center;
  gap: 10px;
}

.config-label {
  font-size: 14px;
  font-weight: 700;
  color: #334155;
}

.config-value {
  font-size: 15px;
  font-weight: 600;
  color: #0f172a;
  word-break: break-all;
}

.config-hint {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 24px;
  padding: 14px 18px;
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 14px;
  color: #64748b;
  font-size: 13px;
}

.config-hint .el-icon {
  color: #94a3b8;
  flex-shrink: 0;
}
</style>
