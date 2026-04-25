<script setup lang="ts">
import { ref, onMounted, reactive } from "vue";
import { ElMessage, ElMessageBox } from "element-plus";
import { apiGetMaterials } from "../api/material";
import type { Material } from "../api/material";
import {
  UploadProductsDianxiaomi,
  ConfirmLogin,
} from "@wailsjs/go/services/AutoUploadService";
import { CountCombinedImagesByMaterialCodes } from "@wailsjs/go/services/ImageService";
import { services } from "@wailsjs/go/models";

const materials = ref<Material[]>([]);
const loading = ref(false);
const currentPage = ref(1);
const pageSize = ref(50);
const total = ref(0);
const totalPages = ref(0);
const combinedImageCounts = ref<Record<string, number | undefined>>({});

// 选中素材
const selectedMaterials = ref<Material[]>([]);

// 登录配置对话框
const loginDialogVisible = ref(false);
const loginForm = reactive({
  url: "https://www.dianxiaomi.com/home.htm",
  username: "17705911697",
  password: "Xzy7530.",
  keepBrowserOpen: true, // 任务完成后保持浏览器开启
});
const uploadLoading = ref(false);

// 等待登录完成对话框
const waitingLoginDialogVisible = ref(false);

// 加载素材列表
const loadMaterials = async () => {
  try {
    loading.value = true;
    const result = await apiGetMaterials({
      page: currentPage.value,
      page_size: pageSize.value,
    });
    materials.value = result.data?.list || [];
    total.value = result.data?.total || 0;
    totalPages.value = result.data?.totalPages || 0;
    await loadCombinedImageCounts();
  } catch (error: any) {
    ElMessage.error(`加载素材失败: ${error.message || error}`);
  } finally {
    loading.value = false;
  }
};

const loadCombinedImageCounts = async () => {
  const materialCodes = materials.value
    .map((material) => String(material.code ?? "").trim())
    .filter(Boolean);

  if (materialCodes.length === 0) {
    combinedImageCounts.value = {};
    return;
  }

  try {
    combinedImageCounts.value =
      await CountCombinedImagesByMaterialCodes(materialCodes);
  } catch (error) {
    combinedImageCounts.value = {};
    console.error("读取合成图片数量失败:", error);
  }
};

const getCombinedImageCount = (code: number | string) =>
  combinedImageCounts.value[String(code ?? "")] || 0;

// 分页改变
const handlePageChange = (page: number) => {
  currentPage.value = page;
  loadMaterials();
};

const handleSizeChange = (size: number) => {
  pageSize.value = size;
  currentPage.value = 1;
  loadMaterials();
};

// 格式化日期
const formatDate = (dateStr: string) => {
  const date = new Date(dateStr);
  return date.toLocaleString("zh-CN");
};

// 打开登录配置对话框
const openUploadDialog = () => {
  if (selectedMaterials.value.length === 0) {
    ElMessage.warning("请先选择要上传的素材");
    return;
  }
  loginDialogVisible.value = true;
};

// 取消上传
const cancelUpload = () => {
  loginDialogVisible.value = false;
};

// 确认上传
const confirmUpload = async () => {
  if (!loginForm.username || !loginForm.password) {
    ElMessage.warning("请输入用户名和密码");
    return;
  }

  loginDialogVisible.value = false;
  uploadLoading.value = true;

  try {
    // 构建登录配置
    const loginConfig = new services.LoginConfig({
      url: loginForm.url,
      username: loginForm.username,
      password: loginForm.password,
      keepBrowserOpen: loginForm.keepBrowserOpen,
    });

    // 构建商品信息映射（批量上传）
    const products: { [key: string]: services.ProductInfo } = {};

    for (const material of selectedMaterials.value) {
      // 使用素材编号作为 productID
      const productID = material.code.toString();

      // 构建产品信息
      const productInfo = new services.ProductInfo({
        titleCh: material.title_cn || `商品${material.code}`,
        titleEn: material.title_en || `Product ${material.code}`,
        materialImg: material.url, // 使用素材URL作为素材图
        imagePaths: [material.url], // 使用素材URL作为品牌图片
      });

      products[productID] = productInfo;
    }

    // 显示等待登录对话框
    waitingLoginDialogVisible.value = true;

    // 调用批量上传接口
    const results = await UploadProductsDianxiaomi(loginConfig, products);

    // 关闭等待对话框
    waitingLoginDialogVisible.value = false;

    // 处理结果
    let successCount = 0;
    let failCount = 0;

    for (const result of results) {
      if (result?.success) {
        successCount++;
      } else {
        failCount++;
      }
    }

    if (successCount > 0) {
      ElMessage.success(
        `上传完成：成功 ${successCount} 个，失败 ${failCount} 个`,
      );
    } else {
      ElMessage.error("所有商品上传失败");
    }

    // 显示详细结果
    if (failCount > 0) {
      const failMessages = results
        .filter((r) => !r?.success)
        .map((r) => r?.message || "未知错误")
        .join("\n");
      console.error("上传失败详情:", failMessages);
    }
  } catch (error: any) {
    waitingLoginDialogVisible.value = false;
    ElMessage.error(`上传失败: ${error.message || error}`);
    console.error("上传错误:", error);
  } finally {
    uploadLoading.value = false;
  }
};

// 确认登录完成
const handleLoginComplete = async () => {
  try {
    await ConfirmLogin();
    ElMessage.success("已确认登录完成，继续执行...");
  } catch (error: any) {
    ElMessage.error(`确认登录失败: ${error.message || error}`);
  }
};

onMounted(() => {
  loadMaterials();
});
</script>

<template>
  <div class="auto-upload-page">
    <section class="page-hero">
      <div class="hero-copy">
        <div class="hero-icon">
          <el-icon :size="20">
            <UploadFilled />
          </el-icon>
        </div>
        <div class="hero-text">
          <h2>自动上传</h2>
          <p class="hero-desc">
            选择已生成标题的素材，批量组装商品信息并自动同步到店小秘。
          </p>
        </div>
      </div>

      <div class="hero-actions">
        <div class="hero-count-pill">
          素材 <strong>{{ total }}</strong>
        </div>
        <div class="hero-count-pill active">
          已选 <strong>{{ selectedMaterials.length }}</strong>
        </div>
      </div>
    </section>

    <section class="action-card">
      <div class="page-header-left">
        <span v-if="selectedMaterials.length > 0" class="selected-count">
          已选择 {{ selectedMaterials.length }} 个素材，准备上传
        </span>
        <span v-else class="selected-hint">
          请先在下方表格中勾选需要上传的素材
        </span>
      </div>
      <el-button
        class="upload-btn"
        type="primary"
        :disabled="selectedMaterials.length === 0 || uploadLoading"
        :loading="uploadLoading"
        @click="openUploadDialog"
      >
        <el-icon class="mr-1"><Upload /></el-icon>
        自动上传到店小秘
      </el-button>
    </section>

    <!-- 表格 -->
    <section class="table-card">
      <div class="table-card-head">
        <div>
          <h3>待上传素材</h3>
          <p>确认中文标题、英文标题和素材 URL 后，再发起自动上传流程。</p>
        </div>
      </div>

      <el-table
        :data="materials"
        v-loading="loading"
        stripe
        @selection-change="(rows: Material[]) => (selectedMaterials = rows)"
      >
        <el-table-column type="selection" width="50" align="center" />
        <el-table-column label="预览" width="100" align="center">
          <template #default="{ row }">
            <el-image
              class="material-thumb"
              :src="row.url"
              :alt="'素材' + row.code"
              fit="cover"
              :preview-src-list="[row.url]"
              preview-teleported
            >
              <template #error>
                <div class="image-placeholder">
                  <el-icon><Picture /></el-icon>
                </div>
              </template>
            </el-image>
          </template>
        </el-table-column>

        <el-table-column
          prop="code"
          label="素材编号"
          width="120"
          align="center"
        >
          <template #default="{ row }">
            <span class="code-chip">#{{ row.code }}</span>
          </template>
        </el-table-column>

        <el-table-column label="合成图数量" width="120" align="center">
          <template #default="{ row }">
            <span
              class="combine-count-chip"
              :class="{
                active: getCombinedImageCount(row.code) > 0,
              }"
            >
              {{ getCombinedImageCount(row.code) }}
            </span>
          </template>
        </el-table-column>

        <el-table-column label="中文标题" min-width="150" align="center">
          <template #default="{ row }">
            <el-tooltip
              v-if="row.title_cn"
              :content="row.title_cn"
              placement="top"
            >
              <span class="title-text">{{ row.title_cn }}</span>
            </el-tooltip>
            <span v-else class="no-title">-</span>
          </template>
        </el-table-column>

        <el-table-column label="英文标题" min-width="150" align="center">
          <template #default="{ row }">
            <el-tooltip
              v-if="row.title_en"
              :content="row.title_en"
              placement="top"
            >
              <span class="title-text">{{ row.title_en }}</span>
            </el-tooltip>
            <span v-else class="no-title">-</span>
          </template>
        </el-table-column>

        <el-table-column label="URL" min-width="300" align="center">
          <template #default="{ row }">
            <el-tooltip :content="row.url" placement="top">
              <span class="url-text">{{ row.url }}</span>
            </el-tooltip>
          </template>
        </el-table-column>

        <el-table-column label="创建时间" width="180" align="center">
          <template #default="{ row }">
            {{ row.createdAt ? formatDate(row.createdAt) : "-" }}
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <div class="pagination-wrapper">
        <el-pagination
          v-model:current-page="currentPage"
          v-model:page-size="pageSize"
          :page-sizes="[50, 100]"
          :total="total"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="handleSizeChange"
          @current-change="handlePageChange"
        />
      </div>
    </section>

    <!-- 等待登录完成对话框 -->
    <el-dialog
      v-model="waitingLoginDialogVisible"
      title="等待登录完成"
      width="400px"
      :close-on-click-modal="false"
      :close-on-press-escape="false"
      :show-close="false"
    >
      <div class="waiting-login-content">
        <el-icon class="waiting-icon"><Loading /></el-icon>
        <p class="waiting-text">请在浏览器中完成以下操作：</p>
        <ol class="waiting-steps">
          <li>如果看到验证码，请输入验证码</li>
          <li>点击登录按钮</li>
          <li>等待登录成功后，点击下方按钮</li>
        </ol>
      </div>
      <template #footer>
        <span class="dialog-footer">
          <el-button type="primary" size="large" @click="handleLoginComplete">
            登录完成，继续执行
          </el-button>
        </span>
      </template>
    </el-dialog>

    <!-- 登录配置对话框 -->
    <el-dialog
      v-model="loginDialogVisible"
      title="店小秘登录配置"
      width="400px"
      :close-on-click-modal="false"
    >
      <el-form :model="loginForm" label-width="80px">
        <el-form-item label="登录地址">
          <el-input
            v-model="loginForm.url"
            placeholder="https://www.dianxiaomi.com/home.htm"
          />
        </el-form-item>
        <el-form-item label="用户名">
          <el-input
            v-model="loginForm.username"
            placeholder="17705911697"
            clearable
          />
        </el-form-item>
        <el-form-item label="密码">
          <el-input
            v-model="loginForm.password"
            type="password"
            placeholder="请输入密码"
            show-password
            clearable
          />
        </el-form-item>
        <el-form-item label="保持开启">
          <el-switch
            v-model="loginForm.keepBrowserOpen"
            active-text="任务完成后保持浏览器开启"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="cancelUpload">取消</el-button>
          <el-button type="primary" @click="confirmUpload">开始上传</el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.auto-upload-page {
  display: block;
  width: 100%;
  box-sizing: border-box;
  padding: 12px;
}

.page-hero {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 20px;
  padding: 14px 16px;
  margin-bottom: 14px;
  border-radius: 18px;
  border: 1px solid #e2e8f0;
  background: rgba(255, 255, 255, 0.8);
}

.hero-copy {
  display: flex;
  align-items: center;
  gap: 14px;
  min-width: 0;
}

.hero-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 44px;
  height: 44px;
  border-radius: 14px;
  color: #4f46e5;
  background: #eef2ff;
  flex-shrink: 0;
}

.hero-text {
  min-width: 0;
}

.hero-text h2 {
  margin: 0;
  color: #0f172a;
  font-size: 20px;
  line-height: 1.2;
  letter-spacing: -0.04em;
}

.hero-desc {
  margin: 6px 0 0;
  color: #64748b;
  font-size: 13px;
  line-height: 1.5;
}

.hero-actions {
  display: flex;
  align-items: center;
  gap: 16px;
  flex-shrink: 0;
}

.hero-count-pill {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  height: 38px;
  padding: 0 14px;
  border-radius: 999px;
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  color: #64748b;
  font-size: 13px;
  white-space: nowrap;
}

.hero-count-pill strong {
  color: #0f172a;
  font-size: 16px;
  font-weight: 800;
}

.hero-count-pill.active {
  background: #eef2ff;
  border-color: #c7d2fe;
  color: #4338ca;
}

.hero-count-pill.active strong {
  color: #4f46e5;
}

.action-card {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 16px;
  padding: 16px 18px;
  margin-bottom: 20px;
  border-radius: 24px;
  border: 1px solid rgba(148, 163, 184, 0.22);
  background: rgba(255, 255, 255, 0.96);
  box-shadow: 0 18px 40px rgba(15, 23, 42, 0.06);
}

.page-header-left {
  display: flex;
  align-items: center;
  gap: 10px;
}

.mr-1 {
  margin-right: 4px;
}

.selected-count {
  font-size: 14px;
  font-weight: 700;
  color: #4338ca;
  background: #eef2ff;
  padding: 10px 16px;
  border-radius: 999px;
}

.selected-hint {
  color: #94a3b8;
  font-size: 13px;
}

.upload-btn {
  height: 46px;
  border: none;
  border-radius: 14px;
  font-weight: 700;
  background: linear-gradient(90deg, #2563eb 0%, #4f46e5 55%, #7c3aed 100%);
  box-shadow: 0 16px 32px rgba(79, 70, 229, 0.2);
}

.table-card {
  padding: 20px;
  border-radius: 28px;
  border: 1px solid rgba(148, 163, 184, 0.22);
  background: rgba(255, 255, 255, 0.96);
  box-shadow: 0 24px 50px rgba(15, 23, 42, 0.08);
}

.table-card-head {
  padding: 4px 6px 18px;
}

.table-card-head h3 {
  margin: 0;
  color: #0f172a;
  font-size: 18px;
  font-weight: 800;
  letter-spacing: -0.04em;
}

.table-card-head p {
  margin: 8px 0 0;
  color: #64748b;
  font-size: 13px;
}

.auto-upload-page :deep(.el-table) {
  border-radius: 22px;
  overflow: hidden;
  --el-table-border-color: #e2e8f0;
  --el-table-header-bg-color: #f8fafc;
  --el-table-row-hover-bg-color: #eef2ff;
  color: #334155;
}

.auto-upload-page :deep(.el-table th.el-table__cell) {
  background: #f8fafc;
  color: #475569;
  font-weight: 700;
}

.auto-upload-page :deep(.el-table .el-table__cell) {
  padding: 14px 0;
}

.image-placeholder {
  width: 60px;
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: center;
  background-color: #f8fafc;
  color: #94a3b8;
  font-size: 20px;
  border-radius: 16px;
  border: 1px solid #e2e8f0;
}

.material-thumb {
  width: 60px;
  height: 60px;
  border-radius: 16px;
  border: 1px solid #e2e8f0;
}

.code-chip {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 8px 12px;
  border-radius: 999px;
  background: #eef2ff;
  color: #4338ca;
  font-size: 12px;
  font-weight: 700;
}

.combine-count-chip {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 48px;
  height: 32px;
  padding: 0 14px;
  border-radius: 999px;
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  color: #64748b;
  font-size: 13px;
  font-weight: 700;
}

.combine-count-chip.active {
  background: #ecfdf5;
  border-color: #bbf7d0;
  color: #16a34a;
}

.url-text {
  display: inline-block;
  max-width: 280px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 12px;
  color: #64748b;
}

.title-text {
  display: inline-block;
  max-width: 130px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 12px;
  color: #475569;
  font-weight: 600;
}

.no-title {
  font-size: 12px;
  color: #cbd5e1;
}

.pagination-wrapper {
  margin-top: 16px;
  display: flex;
  justify-content: flex-end;
}

.waiting-login-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 20px;
  text-align: center;
}

.waiting-icon {
  font-size: 48px;
  color: #4f46e5;
  animation: rotating 2s linear infinite;
  margin-bottom: 16px;
}

@keyframes rotating {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}

.waiting-text {
  font-size: 16px;
  font-weight: 700;
  color: #0f172a;
  margin-bottom: 12px;
}

.waiting-steps {
  text-align: left;
  color: #64748b;
  line-height: 2;
  padding-left: 20px;
}

.waiting-steps li {
  margin-bottom: 8px;
}

.dialog-footer {
  display: flex;
  justify-content: center;
  width: 100%;
}

.auto-upload-page :deep(.el-dialog) {
  border-radius: 24px;
  overflow: hidden;
}

.auto-upload-page :deep(.el-dialog__header) {
  margin: 0;
  padding: 24px 28px 14px;
}

.auto-upload-page :deep(.el-dialog__title) {
  color: #0f172a;
  font-size: 18px;
  font-weight: 800;
}

.auto-upload-page :deep(.el-dialog__body) {
  padding: 8px 28px 20px;
}

.auto-upload-page :deep(.el-dialog__footer) {
  padding: 0 28px 24px;
}

@media (max-width: 760px) {
  .auto-upload-page {
    padding: 16px;
  }

  .page-hero,
  .action-card {
    flex-direction: column;
    align-items: flex-start;
  }

  .page-hero {
    padding: 22px 20px;
  }

  .upload-btn {
    width: 100%;
  }

  .table-card {
    padding: 14px;
  }
}
</style>
