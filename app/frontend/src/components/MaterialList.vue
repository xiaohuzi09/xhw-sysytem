<script setup lang="ts">
import { ref, onMounted } from "vue";
import { ElMessage, ElMessageBox } from "element-plus";
import {
  apiGetMaterials,
  apiDeleteMaterial,
  apiCreateMaterial,
  apiUpdateMaterial,
  apiGenerateProductTitle,
} from "../api/material";
import { apiGetPresignUpload } from "../api/presign";
import { apiGetTemplates } from "../api/template";
import type { Template } from "../api/template";
import { UploadToPresignedURL, CombineImagesWithTemplates } from "@wailsjs/go/services/ImageService";
import { services } from "@wailsjs/go/models";
import type { Material } from "../api/material";

const materials = ref<Material[]>([]);
const loading = ref(false);
const currentPage = ref(1);
const pageSize = ref(50);
const total = ref(0);
const totalPages = ref(0);

// 选中素材
const selectedMaterials = ref<Material[]>([]);

// 弹窗状态
const showViewModal = ref(false);
const showEditModal = ref(false);
const selectedMaterial = ref<Material | null>(null);

// 编辑素材相关状态
const editSelectedFile = ref<File | null>(null);
const editImagePreview = ref("");

// 文件选择 input ref
const fileInputRef = ref<HTMLInputElement | null>(null);
const editFileInputRef = ref<HTMLInputElement | null>(null);

// 上传进度状态
const showUploadProgress = ref(false);
const uploadTotal = ref(0);
const uploadCurrent = ref(0);
const uploadCurrentFile = ref("");

// 合成相关状态
const showCombineModal = ref(false);
const combineLoading = ref(false);
const combineProgress = ref(0);
const combineProgressText = ref("");
const templates = ref<Template[]>([]);
const selectedTemplateIds = ref<string[]>([]);

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
  } catch (error: any) {
    ElMessage.error(`加载素材失败: ${error.message || error}`);
  } finally {
    loading.value = false;
  }
};

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

// 将 File 转换为 base64
const fileToBase64 = (file: File): Promise<string> => {
  return new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.onload = () => resolve(reader.result as string);
    reader.onerror = reject;
    reader.readAsDataURL(file);
  });
};

// 根据文件扩展名推断 Content-Type
const getContentType = (fileName: string): string => {
  const ext = fileName.split(".").pop()?.toLowerCase() || "";
  const map: Record<string, string> = {
    jpg: "image/jpeg",
    jpeg: "image/jpeg",
    png: "image/png",
    gif: "image/gif",
    webp: "image/webp",
    bmp: "image/bmp",
  };
  return map[ext] || "application/octet-stream";
};

// 上传单个文件到云存储
const uploadSingleFile = async (file: File): Promise<string> => {
  // 生成云存储的 key
  const ext = file.name.split(".").pop() || "png";
  const timestamp = Date.now();
  const randomStr = Math.random().toString(36).substring(2, 8);
  const key = `materials/${timestamp}_${randomStr}.${ext}`;

  const contentType = getContentType(file.name);

  // 步骤1: 请求预签名上传链接（带上 Content-Type，确保签名匹配）
  let presignResult: any;
  try {
    presignResult = await apiGetPresignUpload({
      bucket: "xhw",
      key: key,
      expire: 3600,
      content_type: contentType,
    });
  } catch (err: any) {
    throw new Error(`获取预签名链接失败: ${err?.message || err}`);
  }

  if (!presignResult?.data?.url) {
    throw new Error("预签名链接为空");
  }

  // 获取文件 base64 数据
  const base64Data = await fileToBase64(file);

  // 步骤2: 使用 Go 后端方法上传，绕过 WebKit 网络栈
  try {
    await UploadToPresignedURL(presignResult.data.url, base64Data, contentType);
  } catch (err: any) {
    const status = err?.response?.status;
    const msg = err?.response?.data?.message || err?.message || "网络错误";
    throw new Error(`上传到云存储失败(${status || "?"}): ${msg}`);
  }

  // 返回云存储路径
  return presignResult.data.url.split("?")[0];
};

// 触发文件选择
const triggerFileSelect = () => {
  fileInputRef.value?.click();
};

// 处理文件选择
const handleFileSelect = async (event: Event) => {
  const input = event.target as HTMLInputElement;
  const files = input.files;
  if (!files || files.length === 0) return;

  const total = Array.from(files);
  // 初始化进度
  uploadTotal.value = total.length;
  uploadCurrent.value = 0;
  uploadCurrentFile.value = "准备上传...";
  showUploadProgress.value = true;

  let successCount = 0;

  for (let i = 0; i < total.length; i++) {
    const file = total[i];
    // 显示当前处理的文件名
    uploadCurrentFile.value = file.name;

    try {
      // 1. 上传到云存储
      const cloudPath = await uploadSingleFile(file);
      // 2. 创建数据库记录
      await apiCreateMaterial({ url: cloudPath });
      // 3. 两个都成功才算完成一个
      successCount++;
      uploadCurrent.value = successCount;
    } catch (error: any) {
      const errMsg = error?.message || String(error);
      console.error(`上传图片 ${file.name} 失败:`, error);
      ElMessage.error(`${file.name}: ${errMsg}`);
    }
  }

  // 所有文件处理完成
  uploadCurrentFile.value = "上传完成";

  setTimeout(() => {
    showUploadProgress.value = false;
    if (successCount > 0) {
      ElMessage.success(`成功上传 ${successCount} 张图片`);
      loadMaterials();
    }
    if (successCount < total.length) {
      ElMessage.warning(`${total.length - successCount} 张图片上传失败`);
    }
  }, 500);

  // 清空 input 以便重复选择相同文件
  input.value = "";
};

// 查看素材详情
const viewMaterial = (material: Material) => {
  selectedMaterial.value = material;
  showViewModal.value = true;
};

// 打开编辑弹窗
const openEditModal = (material: Material) => {
  selectedMaterial.value = material;
  editSelectedFile.value = null;
  editImagePreview.value = material.url;
  showEditModal.value = true;
};

// 触发编辑文件选择
const triggerEditFileSelect = () => {
  editFileInputRef.value?.click();
};

// 处理编辑文件选择
const handleEditFileSelect = async (event: Event) => {
  const input = event.target as HTMLInputElement;
  const file = input.files?.[0];
  if (!file) return;

  editSelectedFile.value = file;
  // 预览
  const base64 = await fileToBase64(file);
  editImagePreview.value = base64;

  // 清空 input
  input.value = "";
};

// 保存编辑素材
const saveEditMaterial = async () => {
  if (!selectedMaterial.value) return;

  try {
    loading.value = true;

    let url = selectedMaterial.value.url;

    // 如果选择了新图片，则上传
    if (editSelectedFile.value) {
      url = await uploadSingleFile(editSelectedFile.value);
    }

    await apiUpdateMaterial(selectedMaterial.value.id, {
      url: url,
    });

    showEditModal.value = false;
    ElMessage.success("素材更新成功");
    await loadMaterials();
  } catch (error: any) {
    ElMessage.error(`更新素材失败: ${error.message || error}`);
  } finally {
    loading.value = false;
  }
};

// 删除素材
const handleDelete = async (material: Material) => {
  try {
    await ElMessageBox.confirm("确定要删除该素材吗？", "提示", {
      confirmButtonText: "确定",
      cancelButtonText: "取消",
      type: "warning",
    });

    loading.value = true;
    await apiDeleteMaterial(material.id);
    ElMessage.success("素材删除成功");
    await loadMaterials();
  } catch (error: any) {
    if (error !== "cancel") {
      ElMessage.error(`删除素材失败: ${error.message || error}`);
    }
  } finally {
    loading.value = false;
  }
};

// 批量删除选中素材
const batchDelete = async () => {
  if (selectedMaterials.value.length === 0) {
    ElMessage.warning("请先选择要删除的素材");
    return;
  }

  try {
    await ElMessageBox.confirm(
      `确定要删除选中的 ${selectedMaterials.value.length} 个素材吗？`,
      "提示",
      {
        confirmButtonText: "确定",
        cancelButtonText: "取消",
        type: "warning",
      },
    );

    loading.value = true;
    for (const material of selectedMaterials.value) {
      await apiDeleteMaterial(material.id);
    }
    ElMessage.success("批量删除成功");
    selectedMaterials.value = [];
    await loadMaterials();
  } catch (error: any) {
    if (error !== "cancel") {
      ElMessage.error(`删除失败: ${error.message || error}`);
    }
  } finally {
    loading.value = false;
  }
};

// 批量生成标题状态
const showBatchTitleProgress = ref(false);
const batchTitleTotal = ref(0);
const batchTitleCurrent = ref(0);
const batchTitleCurrentFile = ref("");

// 批量生成标题
const batchGenerateTitle = async () => {
  if (selectedMaterials.value.length === 0) {
    ElMessage.warning("请先选择要生成标题的素材");
    return;
  }

  // 初始化进度
  batchTitleTotal.value = selectedMaterials.value.length;
  batchTitleCurrent.value = 0;
  batchTitleCurrentFile.value = "准备生成...";
  showBatchTitleProgress.value = true;

  let successCount = 0;

  for (let i = 0; i < selectedMaterials.value.length; i++) {
    const material = selectedMaterials.value[i];
    batchTitleCurrentFile.value = `素材 ${material.code || i + 1}`;

    try {
      const result = await apiGenerateProductTitle(material.url);
      if (result.data?.title_cn || result.data?.title_en) {
        await apiUpdateMaterial(material.id, {
          title_cn: result.data.title_cn,
          title_en: result.data.title_en,
        });
        successCount++;
      }
    } catch (error: any) {
      console.error(`生成标题失败: ${material.code}`, error);
    }
    batchTitleCurrent.value = i + 1;
  }

  batchTitleCurrentFile.value = "生成完成";

  setTimeout(() => {
    showBatchTitleProgress.value = false;
    if (successCount > 0) {
      ElMessage.success(`成功生成 ${successCount} 个标题`);
      loadMaterials();
    }
    if (successCount < selectedMaterials.value.length) {
      ElMessage.warning(
        `${selectedMaterials.value.length - successCount} 个标题生成失败`,
      );
    }
  }, 500);
};

// 获取模板的 X 偏移值
const getOffsetX = (template: Template): number => {
  if (template.offset_x !== undefined) {
    return template.offset_x;
  }
  return (template.offsetLeft || 0) - (template.offsetRight || 0);
};

// 获取模板的 Y 偏移值
const getOffsetY = (template: Template): number => {
  if (template.offset_y !== undefined) {
    return template.offset_y;
  }
  return (template.offsetTop || 0) - (template.offsetBottom || 0);
};

// 打开合成弹窗（需先选择素材）
const openCombineModal = async () => {
  if (selectedMaterials.value.length === 0) {
    ElMessage.warning("请先选择要合成的素材");
    return;
  }

  showCombineModal.value = true;
  combineProgress.value = 0;
  combineProgressText.value = "";
  selectedTemplateIds.value = [];

  // 加载模板列表
  try {
    combineLoading.value = true;
    const result = await apiGetTemplates();
    templates.value = (result.data || []).map((t) => ({
      ...t,
      id: t.id !== undefined ? String(t.id) : undefined,
    }));

    // 如果只有一个模板，自动选中
    if (templates.value.length === 1 && templates.value[0].id) {
      selectedTemplateIds.value = [templates.value[0].id];
    }
  } catch (error: any) {
    ElMessage.error(`加载模板失败: ${error.message || error}`);
  } finally {
    combineLoading.value = false;
  }
};

// 开始合成
const startCombine = async () => {
  if (selectedTemplateIds.value.length === 0) {
    ElMessage.warning("请选择要使用的模板");
    return;
  }

  try {
    combineLoading.value = true;

    // 获取选中素材的信息（URL 和编号）
    const materials: services.MaterialInfo[] = selectedMaterials.value.map((m) => ({
      url: m.url,
      code: String(m.code || ""),
    }));

    // 获取选中模板的完整信息
    const selectedTemplates: services.TemplateInfo[] = templates.value
      .filter((t) => t.id && selectedTemplateIds.value.includes(t.id))
      .map((t) => ({
        id: t.id!,
        name: t.name,
        width: t.width,
        height: t.height,
        scale: t.scale,
        imagePath: t.imagePath || "",
        url: t.url || "",
        offset_x: getOffsetX(t),
        offset_y: getOffsetY(t),
      }));

    if (selectedTemplates.length === 0) {
      ElMessage.error("未找到选择的模板");
      return;
    }

    // 计算总任务数
    const totalTasks = materials.length * selectedTemplates.length;
    let completedTasks = 0;

    combineProgress.value = 0;
    combineProgressText.value = `0 / ${totalTasks}`;

    // 同步逐个处理，一个完成后再处理下一个
    const failedMaterials: string[] = [];
    for (const material of materials) {
      try {
        // 每次传一个素材，但传所有模板，实际会生成 模板数 张图片
        await CombineImagesWithTemplates(selectedTemplates, [
          material,
        ]);
        // 每次调用完成，增加的已完成任务数 = 模板数量
        completedTasks += selectedTemplates.length;
      } catch (error: any) {
        const errMsg = error?.message || String(error);
        console.error(`合成失败: ${material.code}`, error);
        failedMaterials.push(`${material.code || "无编号"}: ${errMsg}`);
        // 即使失败也计入进度，否则进度条会卡住
        completedTasks += selectedTemplates.length;
      }
      combineProgress.value = Math.round((completedTasks / totalTasks) * 100);
      combineProgressText.value = `${completedTasks} / ${totalTasks}`;
    }

    combineProgress.value = 100;
    combineProgressText.value = "合成完成";

    if (failedMaterials.length === 0) {
      ElMessage.success(`合成完成，共 ${totalTasks} 张图片`);
      // 关闭弹窗
      setTimeout(() => {
        showCombineModal.value = false;
        selectedMaterials.value = [];
      }, 1000);
    } else {
      const successCount = materials.length - failedMaterials.length;
      if (successCount > 0) {
        ElMessage.warning(
          `${successCount} 张素材合成成功，${failedMaterials.length} 张失败`
        );
      } else {
        ElMessage.error(`合成全部失败: ${failedMaterials[0]}`);
      }
    }
  } catch (error: any) {
    ElMessage.error(`合成失败: ${error.message || error}`);
    combineProgress.value = 0;
    combineProgressText.value = "";
  } finally {
    combineLoading.value = false;
  }
};

// 格式化日期
const formatDate = (dateStr: string) => {
  const date = new Date(dateStr);
  return date.toLocaleString("zh-CN");
};

// 生成标题
const generateTitle = async (material: Material) => {
  try {
    loading.value = true;
    const result = await apiGenerateProductTitle(material.url);
    if (result.data?.title_cn || result.data?.title_en) {
      // 更新素材标题
      await apiUpdateMaterial(material.id, {
        title_cn: result.data.title_cn,
        title_en: result.data.title_en,
      });
      ElMessage.success("标题生成并保存成功");
      // 刷新列表
      await loadMaterials();
    }
  } catch (error: any) {
    ElMessage.error(`生成标题失败: ${error.message || error}`);
  } finally {
    loading.value = false;
  }
};

onMounted(() => {
  loadMaterials();
});
</script>

<template>
  <div class="material-list-page">
    <!-- 隐藏的文件选择 input -->
    <input
      ref="fileInputRef"
      type="file"
      multiple
      accept="image/*"
      style="display: none"
      @change="handleFileSelect"
    />
    <input
      ref="editFileInputRef"
      type="file"
      accept="image/*"
      style="display: none"
      @change="handleEditFileSelect"
    />

    <section class="page-hero">
      <div class="hero-copy">
        <div class="hero-icon">
          <el-icon :size="20">
            <PictureFilled />
          </el-icon>
        </div>
        <div class="hero-text">
          <h2>素材管理</h2>
          <p class="hero-desc">
            上传素材图片、批量生成中英文标题，并按模板快速输出合成结果。
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
      <div class="action-left">
        <span v-if="selectedMaterials.length" class="selected-pill">
          已选择 {{ selectedMaterials.length }} 张素材
        </span>
        <span v-else class="selected-hint">可勾选素材进行批量操作</span>
      </div>
      <div class="action-right">
        <el-button
          v-if="selectedMaterials.length > 0"
          class="soft-btn"
          type="danger"
          @click="batchDelete"
        >
          批量删除
        </el-button>
        <el-button
          v-if="selectedMaterials.length > 0"
          class="soft-btn"
          type="info"
          @click="batchGenerateTitle"
        >
          批量生成标题
        </el-button>
        <el-button
          v-if="selectedMaterials.length > 0"
          class="soft-btn"
          type="success"
          @click="openCombineModal"
        >
          <el-icon><PictureFilled /></el-icon>
          合成图片
        </el-button>
        <el-button
          class="primary-btn"
          type="primary"
          @click="triggerFileSelect"
          :loading="loading"
        >
          <el-icon><Plus /></el-icon>
          上传素材
        </el-button>
      </div>
    </section>

    <!-- 上传进度弹窗 -->
    <el-dialog
      v-model="showUploadProgress"
      title="上传进度"
      width="400px"
      :close-on-click-modal="false"
      :close-on-press-escape="false"
      :show-close="false"
    >
      <div class="upload-progress">
        <div class="progress-info">
          <span class="file-name">{{ uploadCurrentFile }}</span>
          <span class="progress-count"
            >{{ uploadCurrent }} / {{ uploadTotal }}</span
          >
        </div>
        <el-progress
          :percentage="
            uploadCurrent >= uploadTotal
              ? 100
              : Math.round((uploadCurrent / uploadTotal) * 99)
          "
          :stroke-width="12"
        />
      </div>
    </el-dialog>

    <!-- 批量生成标题进度弹窗 -->
    <el-dialog
      v-model="showBatchTitleProgress"
      title="批量生成标题进度"
      width="400px"
      :close-on-click-modal="false"
      :close-on-press-escape="false"
      :show-close="false"
    >
      <div class="upload-progress">
        <div class="progress-info">
          <span class="file-name">{{ batchTitleCurrentFile }}</span>
          <span class="progress-count"
            >{{ batchTitleCurrent }} / {{ batchTitleTotal }}</span
          >
        </div>
        <el-progress
          :percentage="
            batchTitleCurrent >= batchTitleTotal
              ? 100
              : Math.round((batchTitleCurrent / batchTitleTotal) * 99)
          "
          :stroke-width="12"
        />
      </div>
    </el-dialog>

    <!-- 表格 -->
    <section class="table-card">
      <div class="table-card-head">
        <div>
          <h3>素材列表</h3>
          <p>支持单条编辑、生成标题、预览原图，也可以多选后进行合成。</p>
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

        <el-table-column label="操作" width="280" fixed="right" align="center">
          <template #default="{ row }">
            <el-button
              class="row-action"
              type="primary"
              size="small"
              @click="viewMaterial(row)"
            >
              查看
            </el-button>
            <el-button
              class="row-action"
              type="warning"
              size="small"
              @click="openEditModal(row)"
            >
              编辑
            </el-button>
            <el-button
              class="row-action"
              type="info"
              size="small"
              @click="generateTitle(row)"
            >
              生成标题
            </el-button>
            <el-button
              class="row-action"
              type="danger"
              size="small"
              @click="handleDelete(row)"
            >
              删除
            </el-button>
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

    <!-- 查看素材弹窗 -->
    <el-dialog v-model="showViewModal" title="素材详情" width="80%">
      <div v-if="selectedMaterial" class="view-modal-content">
        <div class="view-image-wrapper">
          <el-image
            class="view-image"
            :src="selectedMaterial.url"
            fit="contain"
            :preview-src-list="[selectedMaterial.url]"
            preview-teleported
          />
        </div>
        <el-descriptions :column="1" border>
          <el-descriptions-item label="素材编号">{{
            selectedMaterial.code
          }}</el-descriptions-item>
          <el-descriptions-item label="素材URL">{{
            selectedMaterial.url
          }}</el-descriptions-item>
          <el-descriptions-item label="创建时间">{{
            formatDate(selectedMaterial.createdAt)
          }}</el-descriptions-item>
          <el-descriptions-item label="更新时间">{{
            formatDate(selectedMaterial.updatedAt)
          }}</el-descriptions-item>
        </el-descriptions>
      </div>
    </el-dialog>

    <!-- 编辑素材弹窗 -->
    <el-dialog v-model="showEditModal" title="编辑素材" width="80%">
      <el-form label-width="80px">
        <el-form-item label="当前图片">
          <el-image
            v-if="editImagePreview"
            class="edit-image"
            :src="editImagePreview"
            fit="contain"
          />
        </el-form-item>
        <el-form-item label="更换图片">
          <el-button
            type="primary"
            size="small"
            @click="triggerEditFileSelect"
            :loading="loading"
          >
            选择新图片
          </el-button>
          <div v-if="editSelectedFile" class="selected-path">
            {{ editSelectedFile.name }}
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showEditModal = false">取消</el-button>
        <el-button type="primary" @click="saveEditMaterial" :loading="loading">
          保存
        </el-button>
      </template>
    </el-dialog>

    <!-- 合成图片弹窗 -->
    <el-dialog
      v-model="showCombineModal"
      title="合成图片"
      width="500px"
      :close-on-click-modal="false"
    >
      <div class="combine-modal-content">
        <div class="combine-info">
          <p>
            已选择
            <strong>{{ selectedMaterials.length }}</strong> 张素材进行合成
          </p>
        </div>

        <el-form label-width="100px">
          <el-form-item label="选择模板">
            <el-select
              v-model="selectedTemplateIds"
              multiple
              collapse-tags
              collapse-tags-tooltip
              placeholder="请选择模板（可多选）"
              :loading="combineLoading"
              style="width: 100%"
            >
              <el-option
                v-for="t in templates"
                :key="t.id"
                :label="`${t.name}（${t.width}×${t.height} px）`"
                :value="t.id!"
              />
            </el-select>
            <div v-if="templates.length === 0 && !combineLoading" class="hint">
              暂无可用模板，请先创建模板
            </div>
            <div v-else class="hint">
              可多选多个模板，同一批素材会分别套用所选模板
            </div>
          </el-form-item>
        </el-form>

        <!-- 进度条 -->
        <div v-if="combineLoading" class="progress-wrapper">
          <el-progress
            :percentage="combineProgress"
            :stroke-width="12"
            :status="combineProgress === 100 ? 'success' : undefined"
          />
          <div class="progress-text">{{ combineProgressText }}</div>
        </div>

        <!-- 素材预览 -->
        <div v-if="selectedMaterials.length > 0" class="material-preview">
          <p class="preview-label">素材预览：</p>
          <div class="preview-images">
            <el-image
              v-for="(m, index) in selectedMaterials.slice(0, 6)"
              :key="m.id"
              class="preview-thumb"
              :src="m.url"
              :alt="'素材' + (index + 1)"
              fit="cover"
              :preview-src-list="selectedMaterials.map((sm) => sm.url)"
              preview-teleported
            />
            <div v-if="selectedMaterials.length > 6" class="more-indicator">
              +{{ selectedMaterials.length - 6 }}
            </div>
          </div>
        </div>
      </div>

      <template #footer>
        <el-button @click="showCombineModal = false" :disabled="combineLoading">
          取消
        </el-button>
        <el-button
          type="primary"
          @click="startCombine"
          :loading="combineLoading"
          :disabled="selectedTemplateIds.length === 0"
        >
          开始合成
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.material-list-page {
  width: 100%;
  padding: 16px;
}

.page-hero {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 16px;
  padding: 14px 16px;
  margin-bottom: 12px;
  border-radius: 14px;
  border: 1px solid var(--apple-border-soft);
  background: var(--apple-surface);
  box-shadow: var(--apple-shadow-soft);
}

.hero-copy {
  display: flex;
  align-items: center;
  gap: 12px;
  min-width: 0;
}

.hero-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 34px;
  height: 34px;
  border-radius: 8px;
  color: var(--apple-blue);
  background: var(--apple-blue-soft);
  flex-shrink: 0;
}

.hero-text {
  min-width: 0;
}

.hero-text h2 {
  margin: 0;
  color: var(--apple-text);
  font-size: 21px;
  line-height: 1.2;
  font-weight: 650;
  letter-spacing: 0;
}

.hero-desc {
  margin: 4px 0 0;
  color: var(--apple-text-muted);
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
  height: 32px;
  padding: 0 12px;
  border-radius: 999px;
  background: #fbfbfd;
  border: 1px solid var(--apple-border-soft);
  color: var(--apple-text-secondary);
  font-size: 13px;
  white-space: nowrap;
}

.hero-count-pill strong {
  color: var(--apple-text);
  font-size: 14px;
  font-weight: 650;
}

.hero-count-pill.active {
  background: var(--apple-blue-soft);
  border-color: #b9dcff;
  color: var(--apple-blue);
}

.hero-count-pill.active strong {
  color: var(--apple-blue);
}

.action-card {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
  padding: 12px 14px;
  margin-bottom: 12px;
  border-radius: 14px;
  border: 1px solid var(--apple-border-soft);
  background: var(--apple-surface);
  box-shadow: var(--apple-shadow-soft);
}

.action-left,
.action-right {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
}

.selected-pill {
  display: inline-flex;
  align-items: center;
  padding: 7px 12px;
  border-radius: 999px;
  background: var(--apple-blue-soft);
  color: var(--apple-blue);
  font-size: 13px;
  font-weight: 600;
}

.selected-hint {
  color: var(--apple-text-muted);
  font-size: 13px;
}

.soft-btn,
.primary-btn {
  height: 34px;
  border-radius: 7px;
  font-weight: 600;
}

.primary-btn {
  border-color: var(--apple-blue);
  background: var(--apple-blue);
}

.table-card {
  padding: 14px;
  border-radius: 14px;
  border: 1px solid var(--apple-border-soft);
  background: var(--apple-surface);
  box-shadow: var(--apple-shadow-soft);
}

.table-card-head {
  padding: 2px 4px 12px;
}

.table-card-head h3 {
  margin: 0;
  color: var(--apple-text);
  font-size: 16px;
  font-weight: 650;
  letter-spacing: 0;
}

.table-card-head p {
  margin: 4px 0 0;
  color: var(--apple-text-muted);
  font-size: 13px;
}

.material-list-page :deep(.el-table) {
  border-radius: 10px;
  overflow: hidden;
}

.image-placeholder {
  width: 60px;
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: center;
  background-color: #fbfbfd;
  color: var(--apple-text-muted);
  font-size: 20px;
  border-radius: 10px;
  border: 1px solid var(--apple-border-soft);
}

.material-thumb {
  width: 60px;
  height: 60px;
  border-radius: 10px;
  border: 1px solid var(--apple-border-soft);
}

.code-chip {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 5px 9px;
  border-radius: 999px;
  background: var(--apple-blue-soft);
  color: var(--apple-blue);
  font-size: 12px;
  font-weight: 600;
}

.row-action {
  border-radius: 7px;
  font-weight: 500;
}

.selected-path {
  margin-top: 8px;
  padding: 8px 10px;
  background-color: #fbfbfd;
  border-radius: 8px;
  border: 1px solid var(--apple-border-soft);
  font-size: 12px;
  color: var(--apple-text-secondary);
  word-break: break-all;
}

.url-text {
  display: inline-block;
  max-width: 280px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 12px;
  color: var(--apple-text-secondary);
}

.title-text {
  display: inline-block;
  max-width: 130px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 12px;
  color: var(--apple-text-secondary);
  font-weight: 500;
}

.no-title {
  font-size: 12px;
  color: var(--apple-text-muted);
}

.pagination-wrapper {
  margin-top: 16px;
  display: flex;
  justify-content: flex-end;
}

.view-modal-content {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.view-image-wrapper {
  display: flex;
  justify-content: center;
  background: #fbfbfd;
  padding: 16px;
  border-radius: 12px;
  border: 1px solid var(--apple-border-soft);
}

.view-image {
  max-width: 100%;
  max-height: 340px;
  border-radius: 10px;
}

.edit-image {
  max-width: 220px;
  max-height: 180px;
  border-radius: 10px;
  border: 1px solid var(--apple-border-soft);
}

.upload-progress {
  padding: 10px 0;
}

.progress-info {
  display: flex;
  justify-content: space-between;
  margin-bottom: 12px;
  font-size: 14px;
  color: var(--apple-text-secondary);
}

.progress-info .file-name {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  margin-right: 10px;
}

.progress-info .progress-count {
  flex-shrink: 0;
}

/* 合成弹窗样式 */
.combine-modal-content {
  padding: 10px 0;
}

.combine-info {
  margin-bottom: 16px;
  padding: 14px 16px;
  background-color: var(--apple-blue-soft);
  border-radius: 10px;
  border: 1px solid #b9dcff;
}

.combine-info p {
  margin: 0;
  color: var(--apple-blue);
  font-size: 14px;
  font-weight: 600;
}

.hint {
  margin-top: 6px;
  font-size: 12px;
  color: var(--apple-text-muted);
}

.progress-wrapper {
  margin-top: 20px;
}

.progress-wrapper .progress-text {
  margin-top: 8px;
  font-size: 13px;
  color: var(--apple-text-secondary);
  text-align: center;
}

.material-preview {
  margin-top: 20px;
  padding: 14px;
  background-color: #fbfbfd;
  border-radius: 12px;
  border: 1px solid var(--apple-border-soft);
}

.preview-label {
  margin: 0 0 8px 0;
  font-size: 13px;
  color: var(--apple-text-secondary);
  font-weight: 600;
}

.preview-images {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  align-items: center;
}

.preview-thumb {
  width: 60px;
  height: 60px;
  border-radius: 10px;
  border: 1px solid var(--apple-border-soft);
}

.more-indicator {
  width: 60px;
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: center;
  background-color: #f2f2f4;
  border-radius: 10px;
  font-size: 12px;
  color: var(--apple-text-secondary);
  font-weight: 600;
}

@media (max-width: 760px) {
  .material-list-page {
    padding: 12px;
  }

  .page-hero {
    flex-direction: column;
    align-items: flex-start;
    padding: 14px;
  }

  .action-card {
    flex-direction: column;
    align-items: flex-start;
  }

  .action-right {
    width: 100%;
  }

  .primary-btn {
    flex: 1;
  }

  .table-card {
    padding: 12px;
  }
}
</style>
