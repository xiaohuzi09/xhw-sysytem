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
    <div class="manager-orb orb-top" />
    <div class="manager-orb orb-bottom" />

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
  position: relative;
  width: 100%;
  height: 100vh;
  overflow: hidden;
  background:
    radial-gradient(circle at top right, rgba(79, 70, 229, 0.16), transparent 32%),
    radial-gradient(circle at 12% 20%, rgba(14, 165, 233, 0.12), transparent 28%),
    #eef2ff;
}

.manager-orb {
  position: absolute;
  border-radius: 999px;
  pointer-events: none;
  filter: blur(8px);
}

.orb-top {
  top: 72px;
  right: 15%;
  width: 220px;
  height: 220px;
  background: radial-gradient(circle, rgba(99, 102, 241, 0.18), transparent 70%);
}

.orb-bottom {
  left: 180px;
  bottom: -120px;
  width: 280px;
  height: 280px;
  background: radial-gradient(circle, rgba(14, 165, 233, 0.14), transparent 70%);
}

.layout-wrapper {
  position: relative;
  z-index: 1;
  display: flex;
  height: 100vh;
  padding: 0;
  gap: 0;
}

.sidebar {
  width: 248px;
  min-width: 248px;
  height: calc(100vh - 24px);
  margin: 12px 0 12px 12px;
  padding: 20px 14px 16px;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  border: 1px solid rgba(148, 163, 184, 0.22);
  border-radius: 28px;
  background: rgba(255, 255, 255, 0.88);
  box-shadow:
    0 24px 50px rgba(15, 23, 42, 0.08),
    0 8px 20px rgba(79, 70, 229, 0.06);
  backdrop-filter: blur(16px);
}

.brand-block {
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 4px 6px 20px;
}

.brand-logo {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 54px;
  height: 54px;
  border-radius: 18px;
  color: #fff;
  background: linear-gradient(135deg, #2563eb 0%, #4f46e5 55%, #7c3aed 100%);
  box-shadow: 0 16px 28px rgba(79, 70, 229, 0.24);
}

.brand-meta h1 {
  margin: 0;
  color: #0f172a;
  font-size: 17px;
  font-weight: 800;
  line-height: 1.2;
  letter-spacing: -0.04em;
}

.brand-meta p {
  margin: 6px 0 0;
  color: #64748b;
  font-size: 12px;
  line-height: 1.3;
}

.tabs {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.tab-button {
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 12px 14px;
  border: none;
  background: #f8fafc;
  border-radius: 20px;
  color: #475569;
  cursor: pointer;
  transition: all 0.3s;
  text-decoration: none;
  border: 1px solid transparent;
}

.tab-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  border-radius: 14px;
  background: #e2e8f0;
  color: #475569;
  transition: all 0.3s;
}

.tab-text {
  display: flex;
  flex-direction: column;
  gap: 4px;
  min-width: 0;
}

.tab-text strong {
  color: inherit;
  font-size: 15px;
  line-height: 1.2;
  font-weight: 700;
}

.tab-text small {
  color: #94a3b8;
  font-size: 12px;
  line-height: 1.3;
}

.tab-button:hover {
  transform: translateY(-1px);
  background: #fff;
  border-color: rgba(99, 102, 241, 0.12);
  box-shadow: 0 12px 24px rgba(15, 23, 42, 0.06);
}

.tab-button.active {
  color: #3730a3;
  background: linear-gradient(135deg, #eef2ff 0%, #eff6ff 100%);
  border-color: rgba(79, 70, 229, 0.18);
  box-shadow: 0 16px 30px rgba(79, 70, 229, 0.12);
}

.tab-button.active .tab-icon {
  color: #fff;
  background: linear-gradient(135deg, #2563eb 0%, #4f46e5 55%, #7c3aed 100%);
}

.tab-button.active .tab-text small {
  color: #6366f1;
}

.user-section {
  margin-top: auto;
  padding-top: 18px;
  border-top: 1px solid #e2e8f0;
}

.user-card {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 14px;
  border-radius: 18px;
  background: #f8fafc;
  border: 1px solid transparent;
  transition: all 0.3s;
}

.user-card:hover {
  background: #fff;
  border-color: rgba(99, 102, 241, 0.12);
  box-shadow: 0 8px 20px rgba(15, 23, 42, 0.05);
}

.user-avatar {
  flex-shrink: 0;
  width: 40px;
  height: 40px;
  border-radius: 14px;
  overflow: hidden;
  background: linear-gradient(135deg, #2563eb 0%, #4f46e5 55%, #7c3aed 100%);
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
  color: #0f172a;
  font-size: 14px;
  font-weight: 700;
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
  color: #4f46e5;
  background: #eef2ff;
}

.user-role-tag.role-user {
  color: #0f766e;
  background: #f0fdfa;
}

.user-username {
  color: #94a3b8;
  font-size: 12px;
  line-height: 1.3;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.logout-section {
  padding-top: 14px;
}

.logout-tip {
  padding: 0 8px 10px;
  color: #94a3b8;
  font-size: 12px;
}

.logout-button {
  width: 100%;
  justify-content: flex-start;
  height: 50px;
  padding: 0 18px;
  color: #64748b;
  font-size: 15px;
  font-weight: 700;
  border-radius: 16px;
  background: #f8fafc;
}

.logout-button:hover {
  color: #dc2626;
  background-color: #fef2f2;
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
  }

  .tabs {
    display: grid;
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }

  .tab-button {
    padding: 12px;
  }

  .tab-icon {
    width: 40px;
    height: 40px;
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
