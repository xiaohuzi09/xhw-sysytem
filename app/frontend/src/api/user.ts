import request from "./index";
import type { ApiResponse } from "./types";

/** 对齐 xhw-service Swagger：models.User */
export interface User {
  id: number;
  username: string;
  email?: string;
  nickname?: string;
  avatar?: string;
  /** admin | user */
  role?: string;
  /** 1: 正常, 0: 禁用 */
  status?: number;
  created_at?: string;
  updated_at?: string;
}

/** 对齐 controllers.UserCreateRequest */
export interface CreateUserParams {
  username: string;
  password: string;
  email: string;
  nickname?: string;
  avatar?: string;
  status?: number;
}

/** 对齐 controllers.UserUpdateRequest */
export interface UpdateUserParams {
  username?: string;
  password?: string;
  email?: string;
  nickname?: string;
  avatar?: string;
  status?: number;
}

/** GET /api/v1/users — data 为用户数组（无服务端分页） */
export function apiGetUsers(): Promise<ApiResponse<User[]>> {
  return request.get<ApiResponse<User[]>>("/v1/users");
}

export function apiGetUser(id: number): Promise<ApiResponse<User>> {
  return request.get<ApiResponse<User>>(`/v1/users/${id}`);
}

export function apiCreateUser(
  data: CreateUserParams,
): Promise<ApiResponse<User>> {
  return request.post<ApiResponse<User>>("/v1/users", data);
}

export function apiUpdateUser(
  id: number,
  data: UpdateUserParams,
): Promise<ApiResponse<User>> {
  return request.put<ApiResponse<User>>(`/v1/users/${id}`, data);
}

export function apiDeleteUser(id: number): Promise<ApiResponse<void>> {
  return request.delete<ApiResponse<void>>(`/v1/users/${id}`);
}
