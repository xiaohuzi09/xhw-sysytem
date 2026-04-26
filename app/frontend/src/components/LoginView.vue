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
    <div class="bg-orb orb-left" />
    <div class="bg-orb orb-right" />

    <section class="login-shell">
      <div class="login-brand-panel">
        <div class="brand-badge">
          <el-icon :size="20">
            <PictureFilled />
          </el-icon>
          图片模板管理器
        </div>

        <h1>更高效地管理图片模板与素材上传</h1>
        <p class="brand-desc">
          统一维护模板、素材与自动上传流程，让内容生产更轻量、更稳定。
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
          <p class="welcome-label">Welcome back</p>
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
  position: relative;
  width: 100%;
  min-height: 100vh;
  display: flex;
  justify-content: center;
  align-items: center;
  overflow: hidden;
  background:
    radial-gradient(circle at 10% 20%, rgba(79, 70, 229, 0.2) 0, transparent 30%),
    radial-gradient(circle at 85% 20%, rgba(14, 165, 233, 0.2) 0, transparent 28%),
    linear-gradient(135deg, #f8faff 0%, #eef2ff 45%, #f8fbff 100%);
  padding: 24px;
  box-sizing: border-box;
}

.bg-orb {
  position: absolute;
  border-radius: 999px;
  filter: blur(12px);
  opacity: 0.8;
  pointer-events: none;
}

.orb-left {
  left: -120px;
  bottom: 10%;
  width: 280px;
  height: 280px;
  background: radial-gradient(circle, rgba(59, 130, 246, 0.2), transparent 70%);
}

.orb-right {
  right: -100px;
  top: 12%;
  width: 320px;
  height: 320px;
  background: radial-gradient(circle, rgba(99, 102, 241, 0.18), transparent 70%);
}

.login-shell {
  position: relative;
  z-index: 1;
  width: 100%;
  max-width: 980px;
  display: grid;
  grid-template-columns: minmax(320px, 1.05fr) minmax(360px, 0.95fr);
  border: 1px solid rgba(148, 163, 184, 0.25);
  border-radius: 28px;
  background: rgba(255, 255, 255, 0.82);
  box-shadow:
    0 28px 60px rgba(30, 41, 59, 0.12),
    0 10px 26px rgba(59, 130, 246, 0.08);
  backdrop-filter: blur(20px);
  overflow: hidden;
}

.login-brand-panel {
  position: relative;
  padding: 56px 48px;
  background:
    linear-gradient(160deg, rgba(37, 99, 235, 0.96) 0%, rgba(79, 70, 229, 0.94) 50%, rgba(91, 33, 182, 0.96) 100%);
  color: #fff;
  overflow: hidden;
}

.login-brand-panel::after {
  content: "";
  position: absolute;
  right: -72px;
  bottom: -72px;
  width: 220px;
  height: 220px;
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.12);
}

.brand-badge {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  padding: 12px 18px;
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.14);
  border: 1px solid rgba(255, 255, 255, 0.18);
  font-size: 14px;
  font-weight: 600;
  letter-spacing: 0.02em;
}

.login-brand-panel h1 {
  margin: 32px 0 0;
  max-width: 360px;
  font-size: 42px;
  line-height: 1.2;
  letter-spacing: -0.04em;
}

.brand-desc {
  margin: 20px 0 0;
  max-width: 380px;
  color: rgba(239, 246, 255, 0.9);
  font-size: 15px;
  line-height: 1.8;
}

.brand-highlights {
  display: flex;
  flex-direction: column;
  gap: 18px;
  margin-top: 56px;
}

.highlight-item {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 16px 18px;
  border-radius: 20px;
  background: rgba(255, 255, 255, 0.12);
  border: 1px solid rgba(255, 255, 255, 0.16);
}

.highlight-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 44px;
  height: 44px;
  border-radius: 14px;
  background: rgba(255, 255, 255, 0.16);
}

.highlight-item strong {
  display: block;
  font-size: 15px;
  font-weight: 700;
}

.highlight-item p {
  margin: 4px 0 0;
  font-size: 13px;
  color: rgba(239, 246, 255, 0.84);
}

.login-form-panel {
  padding: 56px 48px 48px;
  background: rgba(255, 255, 255, 0.72);
}

.login-header {
  margin-bottom: 32px;
}

.welcome-label {
  margin: 0;
  color: #4f46e5;
  font-size: 13px;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.12em;
}

.login-header h2 {
  margin: 12px 0 0;
  color: #0f172a;
  font-size: 34px;
  line-height: 1.2;
  letter-spacing: -0.04em;
}

.login-subtitle {
  margin: 12px 0 0;
  color: #64748b;
  font-size: 14px;
  line-height: 1.7;
}

.login-button {
  width: 100%;
  height: 52px;
  margin-top: 6px;
  border: none;
  border-radius: 16px;
  font-size: 16px;
  font-weight: 700;
  letter-spacing: 0.08em;
  background: linear-gradient(90deg, #2563eb 0%, #4f46e5 55%, #7c3aed 100%);
  box-shadow: 0 16px 32px rgba(79, 70, 229, 0.22);
}

.login-tip {
  margin: 16px 0 0;
  text-align: center;
  color: #94a3b8;
  font-size: 13px;
}

:deep(.el-form-item) {
  margin-bottom: 24px;
}

:deep(.el-form-item__label) {
  margin-bottom: 10px;
  color: #334155;
  font-size: 14px;
  font-weight: 600;
  line-height: 1.2;
}

:deep(.el-input__wrapper) {
  min-height: 52px;
  padding: 0 18px;
  border-radius: 16px;
  background: #f8fafc;
  box-shadow: 0 0 0 1px #e2e8f0 inset;
  transition:
    box-shadow 0.2s ease,
    background 0.2s ease,
    transform 0.2s ease;
}

:deep(.el-input__wrapper:hover) {
  background: #fff;
  box-shadow: 0 0 0 1px #cbd5e1 inset;
}

:deep(.el-input__wrapper.is-focus) {
  background: #fff;
  box-shadow:
    0 0 0 1px rgba(79, 70, 229, 0.72) inset,
    0 0 0 4px rgba(79, 70, 229, 0.1);
  transform: translateY(-1px);
}

:deep(.el-input__inner) {
  color: #0f172a;
  font-size: 14px;
  font-weight: 500;
}

:deep(.el-input__inner::placeholder) {
  color: #94a3b8;
}

:deep(.el-input__prefix) {
  color: #64748b;
}

@media (max-width: 860px) {
  .login-shell {
    grid-template-columns: 1fr;
    max-width: 520px;
  }

  .login-brand-panel {
    padding: 40px 32px 36px;
  }

  .login-brand-panel h1 {
    max-width: none;
    font-size: 34px;
  }

  .brand-desc {
    max-width: none;
  }

  .brand-highlights {
    margin-top: 32px;
  }

  .login-form-panel {
    padding: 38px 32px 36px;
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
    font-size: 30px;
  }
}
</style>
