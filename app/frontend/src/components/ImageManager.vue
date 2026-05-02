<script setup lang="ts">
import { computed, ref } from "vue";
import { useRouter } from "vue-router";
import { ElMessage, ElMessageBox } from "element-plus";
import {
  Avatar,
  Collection,
  PictureFilled,
  Setting,
  SwitchButton,
  UploadFilled,
  UserFilled,
} from "@element-plus/icons-vue";
import { clearAuthToken, getCurrentUser, isAdmin } from "../utils/auth";

const router = useRouter();
const logoutLoading = ref(false);

const currentUser = computed(() => {
  const user = getCurrentUser();
  console.log("[ImageManager] getCurrentUser:", user);
  if (!user) return null;
  return {
    displayName: user.nickname?.trim() || user.username || "用户",
    username: user.username || "",
    avatar: user.avatar?.trim() || "",
    roleLabel: user.role === "admin" ? "管理员" : "普通用户",
    isAdmin: user.role === "admin",
  };
});

const avatarText = computed(() => {
  const name = currentUser.value?.displayName || "";
  return name.charAt(0).toUpperCase();
});

const baseNavigationItems = [
  {
    to: "/list",
    label: "模板列表",
    desc: "模板尺寸与偏移配置",
    icon: Collection,
  },
  {
    to: "/materials",
    label: "素材管理",
    desc: "素材上传、标题生成与合成",
    icon: PictureFilled,
  },
  {
    to: "/auto-upload",
    label: "自动上传",
    desc: "批量同步到店小秘",
    icon: UploadFilled,
  },
  {
    to: "/config",
    label: "系统配置",
    desc: "运行环境与接口配置",
    icon: Setting,
  },
];

const navigationItems = computed(() => {
  const items = [...baseNavigationItems];
  if (isAdmin()) {
    items.push({
      to: "/users",
      label: "用户管理",
      desc: "账号与权限维护",
      icon: UserFilled,
    });
  }
  return items;
});

const handleLogout = async () => {
  try {
    await ElMessageBox.confirm("确定要退出登录吗？", "退出登录", {
      confirmButtonText: "确定",
      cancelButtonText: "取消",
      type: "warning",
    });

    logoutLoading.value = true;
    clearAuthToken();
    ElMessage.success("已退出登录");
    await router.replace("/login");
  } catch (error) {
    if (error !== "cancel") {
      ElMessage.error("退出登录失败");
    }
  } finally {
    logoutLoading.value = false;
  }
};
</script>

<template>
  <div class="image-manager">
    <div class="layout-wrapper">
      <!-- 左侧：导航菜单 -->
      <div class="sidebar">
        <div class="brand-block">
          <div class="brand-logo">
            <el-icon :size="22">
              <PictureFilled />
            </el-icon>
          </div>
          <div class="brand-meta">
            <h1>图片合成工作台</h1>
            <p>模板 · 素材 · 上传一体化管理</p>
          </div>
        </div>

        <nav class="tabs">
          <router-link
            v-for="item in navigationItems"
            :key="item.to"
            :to="item.to"
            class="tab-button"
            active-class="active"
          >
            <span class="tab-icon">
              <el-icon :size="18">
                <component :is="item.icon" />
              </el-icon>
            </span>
            <span class="tab-text">
              <strong>{{ item.label }}</strong>
              <small>{{ item.desc }}</small>
            </span>
          </router-link>
        </nav>

        <div v-if="currentUser" class="user-section">
          <div class="user-card">
            <div class="user-avatar">
              <img
                v-if="currentUser.avatar"
                :src="currentUser.avatar"
                :alt="currentUser.displayName"
                class="avatar-img"
              />
              <div v-else class="avatar-fallback">
                <el-icon :size="18">
                  <Avatar />
                </el-icon>
              </div>
            </div>
            <div class="user-meta">
              <div class="user-name-row">
                <span class="user-display-name">{{ currentUser.displayName }}</span>
                <span
                  class="user-role-tag"
                  :class="currentUser.isAdmin ? 'role-admin' : 'role-user'"
                >
                  {{ currentUser.roleLabel }}
                </span>
              </div>
              <span class="user-username">@{{ currentUser.username }}</span>
            </div>
          </div>
        </div>

        <div class="logout-section">
          <div class="logout-tip">安全退出当前账号</div>
          <el-button
            class="logout-button"
            :icon="SwitchButton"
            :loading="logoutLoading"
            text
            @click="handleLogout"
          >
            退出登录
          </el-button>
        </div>
      </div>

      <!-- 右侧：路由内容 -->
      <div class="tab-content">
        <div class="content-scroll">
          <router-view />
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.image-manager {
  width: 100%;
  height: 100vh;
  overflow: hidden;
  background: var(--apple-bg);
}

.layout-wrapper {
  display: flex;
  height: 100vh;
}

.sidebar {
  width: 250px;
  min-width: 250px;
  height: 100vh;
  padding: 18px 12px 14px;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  border-right: 1px solid var(--apple-border-soft);
  background: #fbfbfd;
}

.brand-block {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 2px 8px 18px;
}

.brand-logo {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 38px;
  height: 38px;
  border-radius: 10px;
  color: #fff;
  background: var(--apple-blue);
}

.brand-meta h1 {
  margin: 0;
  color: var(--apple-text);
  font-size: 16px;
  font-weight: 650;
  line-height: 1.2;
  letter-spacing: 0;
}

.brand-meta p {
  margin: 4px 0 0;
  color: var(--apple-text-muted);
  font-size: 12px;
  line-height: 1.3;
}

.tabs {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.tab-button {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 9px 10px;
  border-radius: 8px;
  color: var(--apple-text-secondary);
  cursor: pointer;
  text-decoration: none;
  border: 1px solid transparent;
  transition:
    background 0.16s ease,
    border-color 0.16s ease,
    color 0.16s ease;
}

.tab-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border-radius: 7px;
  background: #f0f0f2;
  color: var(--apple-text-muted);
}

.tab-text {
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 0;
}

.tab-text strong {
  color: inherit;
  font-size: 14px;
  line-height: 1.2;
  font-weight: 600;
}

.tab-text small {
  color: var(--apple-text-muted);
  font-size: 12px;
  line-height: 1.3;
}

.tab-button:hover {
  background: #f2f2f4;
}

.tab-button.active {
  color: var(--apple-text);
  background: var(--apple-surface);
  border-color: var(--apple-border-soft);
  box-shadow: var(--apple-shadow-soft);
}

.tab-button.active .tab-icon {
  color: #fff;
  background: var(--apple-blue);
}

.tab-button.active .tab-text small {
  color: var(--apple-blue);
}

.user-section {
  margin-top: auto;
  padding-top: 14px;
  border-top: 1px solid var(--apple-border-soft);
}

.user-card {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px;
  border-radius: 10px;
  background: var(--apple-surface);
  border: 1px solid var(--apple-border-soft);
}

.user-card:hover {
  border-color: var(--apple-border);
}

.user-avatar {
  flex-shrink: 0;
  width: 34px;
  height: 34px;
  border-radius: 9px;
  overflow: hidden;
  background: var(--apple-blue);
  display: inline-flex;
  align-items: center;
  justify-content: center;
}

.avatar-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.avatar-fallback {
  width: 100%;
  height: 100%;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  font-size: 16px;
  font-weight: 700;
}

.user-meta {
  display: flex;
  flex-direction: column;
  gap: 4px;
  min-width: 0;
}

.user-name-row {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
}

.user-display-name {
  color: var(--apple-text);
  font-size: 14px;
  font-weight: 600;
  line-height: 1.2;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.user-role-tag {
  flex-shrink: 0;
  padding: 2px 8px;
  border-radius: 999px;
  font-size: 11px;
  font-weight: 600;
  line-height: 1.4;
}

.user-role-tag.role-admin {
  color: var(--apple-blue);
  background: var(--apple-blue-soft);
}

.user-role-tag.role-user {
  color: var(--apple-green);
  background: #ecf8f0;
}

.user-username {
  color: var(--apple-text-muted);
  font-size: 12px;
  line-height: 1.3;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.logout-section {
  padding-top: 12px;
}

.logout-tip {
  padding: 0 8px 8px;
  color: var(--apple-text-muted);
  font-size: 12px;
}

.logout-button {
  width: 100%;
  justify-content: flex-start;
  height: 38px;
  padding: 0 10px;
  color: var(--apple-text-secondary);
  font-size: 14px;
  font-weight: 500;
  border-radius: 8px;
}

.logout-button:hover {
  color: var(--apple-red);
  background-color: #fff2f4;
}

.tab-content {
  flex: 1;
  min-width: 0;
  height: 100vh;
  overflow: hidden;
}

.content-scroll {
  width: 100%;
  height: 100%;
  overflow-y: auto;
  background: var(--apple-bg);
}

@media (max-width: 980px) {
  .layout-wrapper {
    flex-direction: column;
    overflow: hidden;
  }

  .sidebar {
    width: 100%;
    min-width: 0;
    height: auto;
    border-right: none;
    border-bottom: 1px solid var(--apple-border-soft);
  }

  .tabs {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
  }

  .tab-button {
    padding: 9px 10px;
  }

  .tab-icon {
    width: 28px;
    height: 28px;
  }

  .tab-text small {
    display: none;
  }

  .logout-section {
    margin-top: 18px;
  }

  .tab-content {
    height: auto;
    min-height: 0;
    flex: 1;
  }
}
</style>
