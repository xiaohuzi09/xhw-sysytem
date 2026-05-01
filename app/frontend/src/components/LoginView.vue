<script setup lang="ts">
import { Collection, Lock, PictureFilled, User } from "@element-plus/icons-vue";
import { reactive, ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import { ElMessage } from "element-plus";
import { apiLogin } from "../api/auth";
import { persistAdminAfterLogin, persistUserAfterLogin, setAuthToken } from "../utils/auth";

const route = useRoute();
const router = useRouter();
const loading = ref(false);
const form = reactive({
  username: "",
  password: "",
});

function resolveRedirectPath(): string {
  const redirect = route.query.redirect;
  if (
    typeof redirect === "string" &&
    redirect.startsWith("/") &&
    redirect !== "/login"
  ) {
    return redirect;
  }
  return "/list";
}

const handleLogin = async () => {
  const username = form.username.trim();
  const password = form.password;

  if (!username || !password) {
    ElMessage.warning("请输入账号和密码");
    return;
  }

  try {
    loading.value = true;
    const result = await apiLogin({ username, password });
    const token = result.data?.token?.trim();

    if (!token) {
      throw new Error(result.message || "登录失败：接口未返回 token");
    }

    setAuthToken(token);
    persistAdminAfterLogin({ ...(result.data ?? {}), token });
    console.log("[login] result:", result);
    console.log("[login] result.data:", result.data);
    console.log("[login] result.data.user:", result.data?.user);
    persistUserAfterLogin(result.data?.user);
    ElMessage.success("登录成功");
    await router.replace(resolveRedirectPath());
  } catch (error: any) {
    ElMessage.error(`登录失败: ${error.message || error}`);
  } finally {
    loading.value = false;
  }
};
</script>

<template>
  <div class="login-page">
    <section class="login-shell">
      <div class="login-brand-panel">
        <div class="brand-badge">
          <el-icon :size="20">
            <PictureFilled />
          </el-icon>
          图片模板管理器
        </div>

        <h1>图片合成工作台</h1>
        <p class="brand-desc">
          管理模板、素材和上传流程。
        </p>

        <div class="brand-highlights">
          <div class="highlight-item">
            <span class="highlight-icon">
              <el-icon :size="18">
                <Collection />
              </el-icon>
            </span>
            <div>
              <strong>模板统一管理</strong>
              <p>集中查看和维护全部图片模板</p>
            </div>
          </div>

          <div class="highlight-item">
            <span class="highlight-icon">
              <el-icon :size="18">
                <PictureFilled />
              </el-icon>
            </span>
            <div>
              <strong>素材快速处理</strong>
              <p>高效组合图片并管理素材资源</p>
            </div>
          </div>
        </div>
      </div>

      <div class="login-form-panel">
        <div class="login-header">
          <p class="welcome-label">登录</p>
          <h2>登录账号</h2>
          <p class="login-subtitle">请输入账号密码后继续使用系统</p>
        </div>

        <el-form :model="form" label-position="top" @keyup.enter="handleLogin">
          <el-form-item label="账号">
            <el-input
              v-model="form.username"
              autocomplete="username"
              placeholder="请输入账号"
              size="large"
            >
              <template #prefix>
                <el-icon>
                  <User />
                </el-icon>
              </template>
            </el-input>
          </el-form-item>

          <el-form-item label="密码">
            <el-input
              v-model="form.password"
              autocomplete="current-password"
              placeholder="请输入密码"
              show-password
              size="large"
              type="password"
            >
              <template #prefix>
                <el-icon>
                  <Lock />
                </el-icon>
              </template>
            </el-input>
          </el-form-item>

          <el-button
            class="login-button"
            :loading="loading"
            size="large"
            type="primary"
            @click="handleLogin"
          >
            {{ loading ? "登录中..." : "立即登录" }}
          </el-button>

          <p class="login-tip">按 Enter 可快速提交登录</p>
        </el-form>
      </div>
    </section>
  </div>
</template>

<style scoped>
.login-page {
  width: 100%;
  min-height: 100vh;
  display: flex;
  justify-content: center;
  align-items: center;
  overflow: hidden;
  background: var(--apple-bg);
  padding: 24px;
}

.login-shell {
  width: 100%;
  max-width: 860px;
  display: grid;
  grid-template-columns: minmax(280px, 0.9fr) minmax(360px, 1.1fr);
  border: 1px solid var(--apple-border-soft);
  border-radius: 18px;
  background: var(--apple-surface);
  box-shadow: var(--apple-shadow);
  overflow: hidden;
}

.login-brand-panel {
  position: relative;
  padding: 44px 36px;
  background: #fbfbfd;
  border-right: 1px solid var(--apple-border-soft);
  color: var(--apple-text);
}

.brand-badge {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  padding: 8px 12px;
  border-radius: 999px;
  background: var(--apple-blue-soft);
  color: var(--apple-blue);
  font-size: 13px;
  font-weight: 600;
}

.login-brand-panel h1 {
  margin: 28px 0 0;
  font-size: 30px;
  line-height: 1.2;
  font-weight: 700;
  letter-spacing: 0;
}

.brand-desc {
  margin: 12px 0 0;
  color: var(--apple-text-muted);
  font-size: 14px;
  line-height: 1.6;
}

.brand-highlights {
  display: flex;
  flex-direction: column;
  gap: 10px;
  margin-top: 36px;
}

.highlight-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  border-radius: 12px;
  background: var(--apple-surface);
  border: 1px solid var(--apple-border-soft);
}

.highlight-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 34px;
  height: 34px;
  border-radius: 8px;
  background: #f2f2f4;
  color: var(--apple-text-secondary);
}

.highlight-item strong {
  display: block;
  color: var(--apple-text);
  font-size: 13px;
  font-weight: 600;
}

.highlight-item p {
  margin: 4px 0 0;
  font-size: 12px;
  color: var(--apple-text-muted);
}

.login-form-panel {
  padding: 44px 44px 40px;
  background: var(--apple-surface);
}

.login-header {
  margin-bottom: 28px;
}

.welcome-label {
  margin: 0;
  color: var(--apple-blue);
  font-size: 13px;
  font-weight: 600;
}

.login-header h2 {
  margin: 8px 0 0;
  color: var(--apple-text);
  font-size: 26px;
  line-height: 1.2;
  font-weight: 650;
  letter-spacing: 0;
}

.login-subtitle {
  margin: 10px 0 0;
  color: var(--apple-text-muted);
  font-size: 14px;
  line-height: 1.5;
}

.login-button {
  width: 100%;
  height: 44px;
  margin-top: 6px;
  border-radius: 8px;
  font-size: 15px;
  font-weight: 600;
  letter-spacing: 0;
}

.login-tip {
  margin: 14px 0 0;
  text-align: center;
  color: var(--apple-text-muted);
  font-size: 12px;
}

:deep(.el-form-item) {
  margin-bottom: 20px;
}

:deep(.el-form-item__label) {
  margin-bottom: 8px;
  color: var(--apple-text-secondary);
  font-size: 13px;
  font-weight: 600;
  line-height: 1.2;
}

:deep(.el-input__wrapper) {
  min-height: 42px;
  padding: 0 12px;
}

:deep(.el-input__wrapper:hover) {
  background: var(--apple-surface);
}

:deep(.el-input__inner) {
  color: var(--apple-text);
  font-size: 14px;
}

:deep(.el-input__inner::placeholder) {
  color: var(--apple-text-muted);
}

:deep(.el-input__prefix) {
  color: var(--apple-text-muted);
}

@media (max-width: 860px) {
  .login-shell {
    grid-template-columns: 1fr;
    max-width: 520px;
  }

  .login-brand-panel {
    padding: 32px;
    border-right: none;
    border-bottom: 1px solid var(--apple-border-soft);
  }

  .login-brand-panel h1 {
    font-size: 26px;
  }

  .brand-highlights {
    margin-top: 24px;
  }

  .login-form-panel {
    padding: 32px;
  }
}

@media (max-width: 540px) {
  .login-page {
    padding: 16px;
  }

  .login-brand-panel,
  .login-form-panel {
    padding-left: 24px;
    padding-right: 24px;
  }

  .login-brand-panel h1,
  .login-header h2 {
    font-size: 24px;
  }
}
</style>
