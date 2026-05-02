import request from "./index";
import type { ApiResponse } from "./types";
import type { User } from "./user";
import { AUTH_LOGIN_PATH } from "../utils/auth";

export interface LoginRequest {
  username: string;
  password: string;
}

/** 对齐 Swagger：controllers.LoginResponse */
export interface LoginResult {
  token: string;
  user?: User;
}

export function apiLogin(
  data: LoginRequest,
): Promise<ApiResponse<LoginResult>> {
  return request.post<ApiResponse<LoginResult>>(AUTH_LOGIN_PATH, data);
}

export default {
  apiLogin,
};
