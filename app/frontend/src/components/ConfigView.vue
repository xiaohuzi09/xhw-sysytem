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
  padding: 16px;
}

.page-header {
  padding: 14px 16px;
  margin-bottom: 12px;
  border: 1px solid var(--apple-border-soft);
  border-radius: 14px;
  background: var(--apple-surface);
  box-shadow: var(--apple-shadow-soft);
}

.page-header h2 {
  margin: 0;
  font-size: 21px;
  font-weight: 650;
  color: var(--apple-text);
  letter-spacing: 0;
}

.subtitle {
  margin: 4px 0 0;
  color: var(--apple-text-muted);
  font-size: 13px;
}

.config-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.config-card {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 14px 16px;
  background: var(--apple-surface);
  border: 1px solid var(--apple-border-soft);
  border-radius: 14px;
  box-shadow: var(--apple-shadow-soft);
  transition: border-color 0.16s ease;
}

.config-card:hover {
  border-color: var(--apple-border);
}

.config-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 34px;
  height: 34px;
  border-radius: 8px;
  flex-shrink: 0;
}

.config-icon.type-primary {
  color: var(--apple-blue);
  background: var(--apple-blue-soft);
}
.config-icon.type-success {
  color: var(--apple-green);
  background: #ecf8f0;
}
.config-icon.type-warning {
  color: var(--apple-orange);
  background: #fff6e5;
}
.config-icon.type-info {
  color: var(--apple-text-secondary);
  background: #f2f2f4;
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
  font-weight: 600;
  color: var(--apple-text-secondary);
}

.config-value {
  font-size: 15px;
  font-weight: 600;
  color: var(--apple-text);
  word-break: break-all;
}

.config-hint {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 12px;
  padding: 12px 14px;
  background: var(--apple-surface);
  border: 1px solid var(--apple-border-soft);
  border-radius: 12px;
  color: var(--apple-text-secondary);
  font-size: 13px;
}

.config-hint .el-icon {
  color: var(--apple-text-muted);
  flex-shrink: 0;
}
</style>
