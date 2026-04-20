import request from "./index";
import type { ApiResponse } from "./types";

// 模板数据结构
export interface Template {
  id?: string;
  name: string;
  width: number;
  height: number;
  scale: number;
  imagePath: string;
  // API 返回的字段
  offset_x?: number;
  offset_y?: number;
  // 兼容旧字段（本地存储使用）
  offsetTop?: number;
  offsetRight?: number;
  offsetBottom?: number;
  offsetLeft?: number;
  url?: string;
  createdAt?: string;
}

// 创建模板请求参数
export interface CreateTemplateRequest {
  name: string;
  width: number;
  height: number;
  scale: number;
  imagePath: string;
  offset_x: number;
  offset_y: number;
  url: string;
}

// 创建模板
export function apiCreateTemplate(
  data: CreateTemplateRequest,
): Promise<ApiResponse<Template>> {
  return request.post<ApiResponse<Template>>("/v1/templates", data);
}

// 获取模板列表
export function apiGetTemplates(): Promise<ApiResponse<Template[]>> {
  return request.get<ApiResponse<Template[]>>("/v1/templates");
}

// 删除模板
export function apiDeleteTemplate(id: string): Promise<ApiResponse<void>> {
  return request.delete<ApiResponse<void>>(`/v1/templates/${id}`);
}

// 更新模板
export function apiUpdateTemplate(
  id: string,
  data: CreateTemplateRequest,
): Promise<ApiResponse<Template>> {
  return request.put<ApiResponse<Template>>(`/v1/templates/${id}`, data);
}

export default {
  apiCreateTemplate,
  apiGetTemplates,
  apiDeleteTemplate,
  apiUpdateTemplate,
};
