import { createRouter, createWebHashHistory } from "vue-router";
import type { RouteRecordRaw } from "vue-router";
import { ElMessage } from "element-plus";
import { hasAuthToken, isAdmin } from "../utils/auth";

const DEFAULT_AUTHENTICATED_PATH = "/list";

function resolveRedirectPath(redirect?: unknown): string {
  if (
    typeof redirect === "string" &&
    redirect.startsWith("/") &&
    redirect !== "/login"
  ) {
    return redirect;
  }

  return DEFAULT_AUTHENTICATED_PATH;
}

const routes: RouteRecordRaw[] = [
  {
    path: "/login",
    name: "Login",
    component: () => import("../components/LoginView.vue"),
    meta: { title: "登录" },
  },
  {
    path: "/",
    component: () => import("../components/ImageManager.vue"),
    redirect: DEFAULT_AUTHENTICATED_PATH,
    meta: { requiresAuth: true },
    children: [
      {
        path: "list",
        name: "TemplateList",
        component: () => import("../components/TemplateList.vue"),
        meta: { title: "模板列表" },
      },
      {
        path: "materials",
        name: "MaterialList",
        component: () => import("../components/MaterialList.vue"),
        meta: { title: "素材管理" },
      },
      {
        path: "auto-upload",
        name: "AutoUpload",
        component: () => import("../components/AutoUpload.vue"),
        meta: { title: "自动上传" },
      },
      {
        path: "users",
        name: "UserManage",
        component: () => import("../components/UserManage.vue"),
        meta: { title: "用户管理", requiresAdmin: true },
      },
    ],
  },
];

const router = createRouter({
  history: createWebHashHistory(),
  routes,
});

router.beforeEach((to) => {
  const isLoggedIn = hasAuthToken();

  if (to.path === "/login") {
    if (isLoggedIn) {
      return resolveRedirectPath(to.query.redirect);
    }
    return true;
  }

  const requiresAuth = to.matched.some((route) =>
    Boolean(route.meta.requiresAuth),
  );

  if (requiresAuth && !isLoggedIn) {
    return {
      path: "/login",
      query: {
        redirect: to.fullPath,
      },
    };
  }

  const requiresAdmin = to.matched.some((route) =>
    Boolean(route.meta.requiresAdmin),
  );
  if (requiresAdmin && isLoggedIn && !isAdmin()) {
    ElMessage.warning("需要管理员权限");
    return DEFAULT_AUTHENTICATED_PATH;
  }

  return true;
});

export default router;
