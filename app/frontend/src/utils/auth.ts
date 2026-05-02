export const AUTH_TOKEN_KEY = "app-image-handle-token";
export const AUTH_LOGIN_PATH = "/v1/auth/login";

/** 与 token 同生命周期：退出登录时清除 */
const AUTH_ADMIN_SESSION_KEY = "app-image-handle-is-admin";
const AUTH_USER_SESSION_KEY = "app-image-handle-user";

export function getAuthToken(): string {
  return localStorage.getItem(AUTH_TOKEN_KEY) || "";
}

export function setAuthToken(token: string) {
  localStorage.setItem(AUTH_TOKEN_KEY, token);
}

function clearAdminSession() {
  sessionStorage.removeItem(AUTH_ADMIN_SESSION_KEY);
}

function clearUserSession() {
  sessionStorage.removeItem(AUTH_USER_SESSION_KEY);
}

export function clearAuthToken() {
  localStorage.removeItem(AUTH_TOKEN_KEY);
  clearAdminSession();
  clearUserSession();
}

/** 当前登录用户信息 */
export interface CurrentUser {
  id?: number;
  username: string;
  nickname?: string;
  avatar?: string;
  role?: string;
}

/** 登录成功后保存用户信息 */
export function persistUserAfterLogin(user: CurrentUser | undefined) {
  if (user) {
    sessionStorage.setItem(AUTH_USER_SESSION_KEY, JSON.stringify(user));
  }
}

/** 获取当前登录用户信息（优先从 sessionStorage，否则从 JWT 解码） */
export function getCurrentUser(): CurrentUser | null {
  const cached = sessionStorage.getItem(AUTH_USER_SESSION_KEY);
  if (cached) {
    try {
      return JSON.parse(cached) as CurrentUser;
    } catch {
      // ignore parse error
    }
  }
  const payload = decodeJwtPayload(getAuthToken());
  if (payload) {
    return {
      username: (payload.username as string) || (payload.sub as string) || "",
      role: (payload.role as string) || "",
    };
  }
  return null;
}

export function hasAuthToken(): boolean {
  return Boolean(getAuthToken());
}

/** 解码 JWT payload（不校验签名，仅用于前端展示与路由；权限以服务端为准） */
function decodeJwtPayload(
  token: string,
): Record<string, unknown> | null {
  const parts = token.split(".");
  if (parts.length < 2) {
    return null;
  }
  try {
    let base64 = parts[1].replace(/-/g, "+").replace(/_/g, "/");
    const pad = (4 - (base64.length % 4)) % 4;
    base64 += "=".repeat(pad);
    const json = atob(base64);
    return JSON.parse(json) as Record<string, unknown>;
  } catch {
    return null;
  }
}

function payloadIndicatesAdmin(
  payload: Record<string, unknown> | null,
): boolean {
  if (!payload) {
    return false;
  }
  const role = payload.role;
  if (role === "admin" || role === "ADMIN") {
    return true;
  }
  if (payload.is_admin === true) {
    return true;
  }
  const roles = payload.roles;
  if (Array.isArray(roles) && roles.some((r) => r === "admin" || r === "ADMIN")) {
    return true;
  }
  const userRole = (payload.user as Record<string, unknown> | undefined)?.role;
  if (userRole === "admin" || userRole === "ADMIN") {
    return true;
  }
  if ((payload.user as Record<string, unknown> | undefined)?.is_admin === true) {
    return true;
  }
  return false;
}

/** 登录接口 data 中是否明示管理员（兼容多种常见字段） */
export function loginDataIndicatesAdmin(data: {
  role?: string;
  is_admin?: boolean;
  user?: { role?: string; is_admin?: boolean };
}): boolean | null {
  if (data.is_admin === true) {
    return true;
  }
  if (data.is_admin === false) {
    return false;
  }
  if (data.role === "admin" || data.role === "ADMIN") {
    return true;
  }
  if (data.role != null && data.role !== "") {
    return false;
  }
  const u = data.user;
  if (u?.is_admin === true) {
    return true;
  }
  if (u?.is_admin === false) {
    return false;
  }
  if (u?.role === "admin" || u?.role === "ADMIN") {
    return true;
  }
  if (u?.role != null && u?.role !== "") {
    return false;
  }
  return null;
}

/**
 * 登录成功后调用：优先使用接口返回的角色字段，否则从 JWT 推断并写入 sessionStorage。
 */
export function persistAdminAfterLogin(data: {
  token?: string;
  role?: string;
  is_admin?: boolean;
  user?: { role?: string; is_admin?: boolean };
}) {
  const explicit = loginDataIndicatesAdmin(data);
  if (explicit === true) {
    sessionStorage.setItem(AUTH_ADMIN_SESSION_KEY, "true");
    return;
  }
  if (explicit === false) {
    sessionStorage.setItem(AUTH_ADMIN_SESSION_KEY, "false");
    return;
  }
  const token = data.token?.trim() || getAuthToken();
  const payload = decodeJwtPayload(token);
  sessionStorage.setItem(
    AUTH_ADMIN_SESSION_KEY,
    payloadIndicatesAdmin(payload) ? "true" : "false",
  );
}

/**
 * 当前用户是否为管理员（用于路由与菜单；以服务端接口鉴权为准）。
 */
export function isAdmin(): boolean {
  if (!hasAuthToken()) {
    return false;
  }
  const cached = sessionStorage.getItem(AUTH_ADMIN_SESSION_KEY);
  if (cached === "true") {
    return true;
  }
  if (cached === "false") {
    return false;
  }
  const payload = decodeJwtPayload(getAuthToken());
  const admin = payloadIndicatesAdmin(payload);
  sessionStorage.setItem(AUTH_ADMIN_SESSION_KEY, admin ? "true" : "false");
  return admin;
}
