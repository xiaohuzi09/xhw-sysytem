import request from "./index";
import type { ApiResponse } from "./types";

// 预签名上传请求参数
export interface PresignUploadRequest {
  bucket: string;
  key: string;
  expire?: number; // 过期时间（秒），默认3600
}

// 预签名上传响应数据
export interface PresignUploadData {
  url: string;
  expires_in: number;
}

// 获取预签名上传URL
export function apiGetPresignUpload(
  data: PresignUploadRequest,
): Promise<ApiResponse<PresignUploadData>> {
  return request.post<ApiResponse<PresignUploadData>>(
    "/v1/rustfs/presign/upload",
    data,
  );
}

export default {
  apiGetPresignUpload,
};
