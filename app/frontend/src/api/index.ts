import axios, {
  AxiosInstance,
  AxiosRequestConfig,
  AxiosResponse,
  InternalAxiosRequestConfig,
} from "axios";
import router from "../router";
import {
  AUTH_LOGIN_PATH,
  clearAuthToken,
  getAuthToken,
} from "../utils/auth";

// 环境变量
const isDev = import.meta.env.VITE_APP_ENV === "development";
const baseURL = import.meta.env.VITE_API_BASE_URL || "/api";

// 创建 axios 实例
const instance: AxiosInstance = axios.create({
  baseURL,
  timeout: 30000,
  headers: {
    "Content-Type": "application/json",
  },
});

function shouldAttachToken(url = ""): boolean {
  return !url.endsWith(AUTH_LOGIN_PATH);
}

// 请求拦截器
instance.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    if (shouldAttachToken(config.url)) {
      const token = getAuthToken();
      if (token) {
        config.headers.Authorization = `Bearer ${token}`;
      }
    }

    if (isDev) {
      console.log(
        "[API Request]",
        config.method?.toUpperCase(),
        config.url,
        config,
      );
    }

    return config;
  },
  (error) => {
    console.error("[API Request Error]", error);
    return Promise.reject(error);
  },
);

// 响应拦截器
instance.interceptors.response.use(
  (response: AxiosResponse) => {
    if (isDev) {
      console.log("[API Response]", response.config.url, response.data);
    }

    // 可以在这里统一处理响应数据
    const { data } = response;
    return data;
  },
  (error) => {
    console.error("[API Response Error]", error);

    // 统一错误处理
    const status = error.response?.status;
    const message =
      error.response?.data?.message || error.message || "请求失败";

    if (status === 401) {
      clearAuthToken();
      const currentRoute = router.currentRoute.value;
      if (currentRoute.path !== "/login") {
        router
          .replace({
            path: "/login",
            query: { redirect: currentRoute.fullPath },
          })
          .catch(() => {});
      }
    }

    return Promise.reject(new Error(message));
  },
);

// 封装请求方法
const request = {
  get<T = any>(url: string, config?: AxiosRequestConfig): Promise<T> {
    return instance.get(url, config);
  },

  post<T = any>(
    url: string,
    data?: any,
    config?: AxiosRequestConfig,
  ): Promise<T> {
    return instance.post(url, data, config);
  },

  put<T = any>(
    url: string,
    data?: any,
    config?: AxiosRequestConfig,
  ): Promise<T> {
    return instance.put(url, data, config);
  },

  patch<T = any>(
    url: string,
    data?: any,
    config?: AxiosRequestConfig,
  ): Promise<T> {
    return instance.patch(url, data, config);
  },

  delete<T = any>(url: string, config?: AxiosRequestConfig): Promise<T> {
    return instance.delete(url, config);
  },
};

export default request;
export { instance as axiosInstance };
