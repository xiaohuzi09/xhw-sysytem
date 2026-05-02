import request from "./index";
import type { ApiResponse } from "./types";

// 素材数据结构
export interface Material {
  id: number;
  url: string; // 素材URL
  code: number; // 素材编号
  title_cn?: string; // 中文标题
  title_en?: string; // 英文标题
  createdAt: string;
  updatedAt: string;
}

// 素材列表返回结果
export interface MaterialListResult {
  list: Material[];
  total: number;
  page: number;
  pageSize: number;
  totalPages: number;
}

// 获取素材列表参数
export interface GetMaterialsParams {
  page?: number;
  page_size?: number;
}

// 创建素材参数
export interface CreateMaterialParams {
  url: string;
}

// 更新素材参数
export interface UpdateMaterialParams {
  url?: string;
  title_cn?: string;
  title_en?: string;
}

// 获取素材列表
export function apiGetMaterials(
  params: GetMaterialsParams = {},
): Promise<ApiResponse<MaterialListResult>> {
  const { page = 1, page_size = 10 } = params;
  return request.get<ApiResponse<MaterialListResult>>("/v1/materials", {
    params: { page, page_size },
  });
}

// 获取单个素材
export function apiGetMaterial(id: number): Promise<ApiResponse<Material>> {
  return request.get<ApiResponse<Material>>(`/v1/materials/${id}`);
}

// 创建素材
export function apiCreateMaterial(
  data: CreateMaterialParams,
): Promise<ApiResponse<Material>> {
  return request.post<ApiResponse<Material>>("/v1/materials", data);
}

// 更新素材
export function apiUpdateMaterial(
  id: number,
  data: UpdateMaterialParams,
): Promise<ApiResponse<Material>> {
  return request.put<ApiResponse<Material>>(`/v1/materials/${id}`, data);
}

// 删除素材
export function apiDeleteMaterial(id: number): Promise<ApiResponse<void>> {
  return request.delete<ApiResponse<void>>(`/v1/materials/${id}`);
}

// 生成产品标题
export function apiGenerateProductTitle(
  url: string,
): Promise<ApiResponse<{ title_cn: string; title_en: string }>> {
  return request.post<ApiResponse<{ title_cn: string; title_en: string }>>(
    "/v1/ark/product-title",
    {
      image_url: url,
    },
  );
}

export default {
  apiGetMaterials,
  apiGetMaterial,
  apiCreateMaterial,
  apiUpdateMaterial,
  apiDeleteMaterial,
};
