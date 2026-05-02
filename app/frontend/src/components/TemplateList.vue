<script setup lang="ts">
import { ref, onMounted, computed, watch } from "vue";
import { ElMessage, ElMessageBox } from "element-plus";
import { SelectImage, GetImageBase64, UploadToPresignedURL } from "@wailsjs/go/services/ImageService";
import {
  apiGetTemplates,
  apiDeleteTemplate,
  apiCreateTemplate,
  apiUpdateTemplate,
} from "../api/template";
import { apiGetPresignUpload } from "../api/presign";
import type { Template } from "../api/template";

const templates = ref<Template[]>([]);
const loading = ref(false);

// 获取模板的 X 偏移值（兼容 API 和本地存储两种格式）
const getOffsetX = (template: Template): number => {
  // API 返回 offset_x
  if (template.offset_x !== undefined) {
    return template.offset_x;
  }
  // 本地存储使用 offsetLeft - offsetRight
  return (template.offsetLeft || 0) - (template.offsetRight || 0);
};

// 获取模板的 Y 偏移值（兼容 API 和本地存储两种格式）
const getOffsetY = (template: Template): number => {
  // API 返回 offset_y
  if (template.offset_y !== undefined) {
    return template.offset_y;
  }
  // 本地存储使用 offsetTop - offsetBottom
  return (template.offsetTop || 0) - (template.offsetBottom || 0);
};

const selectedTemplate = ref<Template | null>(null);
const showViewModal = ref(false);
const showEditModal = ref(false);
const showAddModal = ref(false);
const previewImageBase64 = ref("");

const viewImageElement = ref<HTMLImageElement | null>(null);
const viewImageNaturalWidth = ref(0);
const viewImageNaturalHeight = ref(0);
const viewImageLoaded = ref(false);

const editName = ref("");
const editWidth = ref(0);
const editHeight = ref(0);
const editScale = ref(1);
const editOffsetX = ref(0);
const editOffsetY = ref(0);

const editImageElement = ref<HTMLImageElement | null>(null);
const editImageNaturalWidth = ref(0);
const editImageNaturalHeight = ref(0);
const editImageLoaded = ref(false);
const editImageBase64 = ref("");

// 新增模板相关状态
const addName = ref("");
const addWidth = ref(800);
const addHeight = ref(600);
const addScale = ref(1.0);
const addOffsetX = ref(0);
const addOffsetY = ref(0);
const addSelectedImagePath = ref("");
const addImageElement = ref<HTMLImageElement | null>(null);
const addImageNaturalWidth = ref(0);
const addImageNaturalHeight = ref(0);
const addImageLoaded = ref(false);
const addImageBase64 = ref("");

// 查看时矩形在预览图中的位置和大小
const viewRectangleStyle = computed(() => {
  if (
    !viewImageLoaded.value ||
    !viewImageElement.value ||
    !selectedTemplate.value
  )
    return;

  const img = viewImageElement.value;
  const displayWidth = img.clientWidth;
  const displayHeight = img.clientHeight;

  // 如果 naturalWidth/naturalHeight 为 0（跨域问题），使用 clientWidth/clientHeight
  const naturalW = viewImageNaturalWidth.value || displayWidth;
  const naturalH = viewImageNaturalHeight.value || displayHeight;

  if (!displayWidth || !displayHeight) {
    return;
  }

  const scaleX = displayWidth / naturalW;
  const scaleY = displayHeight / naturalH;

  const rectWidth = selectedTemplate.value.width * scaleX;
  const rectHeight = selectedTemplate.value.height * scaleY;

  const baseLeft = (displayWidth - rectWidth) / 2;
  const baseTop = (displayHeight - rectHeight) / 2;

  const dx = getOffsetX(selectedTemplate.value) * scaleX;
  const dy = getOffsetY(selectedTemplate.value) * scaleY;

  const left = baseLeft + dx;
  const top = baseTop + dy;

  console.log("viewRectangleStyle:", {
    displayWidth,
    displayHeight,
    naturalW,
    naturalH,
    rectWidth,
    rectHeight,
    left,
    top,
    offsetLeft: selectedTemplate.value.offsetLeft,
    offsetRight: selectedTemplate.value.offsetRight,
    offsetTop: selectedTemplate.value.offsetTop,
    offsetBottom: selectedTemplate.value.offsetBottom,
    dx,
    dy,
  });

  return {
    width: `${rectWidth}px`,
    height: `${rectHeight}px`,
    left: `${left}px`,
    top: `${top}px`,
  };
});

// 编辑时矩形在预览图中的位置和大小
const editRectangleStyle = computed(() => {
  if (!editImageLoaded.value || !editImageElement.value) return;

  const img = editImageElement.value;
  const displayWidth = img.clientWidth;
  const displayHeight = img.clientHeight;

  // 如果 naturalWidth/naturalHeight 为 0（跨域问题），使用 clientWidth/clientHeight
  const naturalW = editImageNaturalWidth.value || displayWidth;
  const naturalH = editImageNaturalHeight.value || displayHeight;

  if (!displayWidth || !displayHeight) {
    return;
  }

  const scaleX = displayWidth / naturalW;
  const scaleY = displayHeight / naturalH;

  const rectWidth = editWidth.value * scaleX;
  const rectHeight = editHeight.value * scaleY;

  const baseLeft = (displayWidth - rectWidth) / 2;
  const baseTop = (displayHeight - rectHeight) / 2;

  const dx = (editOffsetX.value || 0) * scaleX;
  const dy = (editOffsetY.value || 0) * scaleY;

  const left = baseLeft + dx;
  const top = baseTop + dy;

  console.log("editRectangleStyle:", {
    displayWidth,
    displayHeight,
    naturalW,
    naturalH,
    rectWidth,
    rectHeight,
    left,
    top,
    editOffsetX: editOffsetX.value,
    editOffsetY: editOffsetY.value,
    dx,
    dy,
  });

  return {
    width: `${rectWidth}px`,
    height: `${rectHeight}px`,
    left: `${left}px`,
    top: `${top}px`,
  };
});

// 加载模板列表
const loadTemplates = async () => {
  try {
    loading.value = true;
    const result = await apiGetTemplates();
    templates.value = result.data || [];
  } catch (error: any) {
    ElMessage.error(`加载模板失败: ${error.message || error}`);
  } finally {
    loading.value = false;
  }
};

// 删除模板
const handleDelete = async (row: Template) => {
  if (!row.id) return;

  try {
    await ElMessageBox.confirm("确定要删除该模板吗？", "提示", {
      confirmButtonText: "确定",
      cancelButtonText: "取消",
      type: "warning",
    });

    loading.value = true;
    await apiDeleteTemplate(row.id);
    ElMessage.success("模板删除成功");
    await loadTemplates();
  } catch (error: any) {
    if (error !== "cancel") {
      ElMessage.error(`删除模板失败: ${error.message || error}`);
    }
  } finally {
    loading.value = false;
  }
};

// 查看模板详情
const viewTemplate = async (template: Template) => {
  selectedTemplate.value = template;
  showViewModal.value = true;
  previewImageBase64.value = "";
  viewImageLoaded.value = false;
  viewImageNaturalWidth.value = 0;
  viewImageNaturalHeight.value = 0;

  try {
    // 优先使用云存储 URL，其次使用本地路径
    const imageUrl = template.url || template.imagePath;

    if (
      imageUrl &&
      (imageUrl.startsWith("http://") || imageUrl.startsWith("https://"))
    ) {
      // 云存储 URL，直接使用
      previewImageBase64.value = imageUrl;
    } else if (template.imagePath) {
      // 本地路径，调用后端读取
      const base64 = await GetImageBase64(template.imagePath);
      previewImageBase64.value = base64;
    } else {
      ElMessage.warning("该模板没有关联的图片");
    }
  } catch (error: any) {
    ElMessage.error(`加载预览失败: ${error.message || error}`);
  }
};

// 查看图片加载完成
const onViewImageLoad = (event: Event) => {
  const img = event.target as HTMLImageElement;
  console.log("View image loaded:", {
    naturalWidth: img.naturalWidth,
    naturalHeight: img.naturalHeight,
    clientWidth: img.clientWidth,
    clientHeight: img.clientHeight,
  });
  viewImageNaturalWidth.value = img.naturalWidth;
  viewImageNaturalHeight.value = img.naturalHeight;
  viewImageLoaded.value = true;
};

// 查看图片加载失败
const onViewImageError = () => {
  viewImageLoaded.value = false;
  ElMessage.error("图片加载失败");
};

// 打开编辑弹窗
const openEditTemplate = async (template: Template) => {
  selectedTemplate.value = template;
  editName.value = template.name;
  editWidth.value = template.width;
  editHeight.value = template.height;
  editScale.value = template.scale;
  editOffsetX.value = getOffsetX(template);
  editOffsetY.value = getOffsetY(template);
  editImageBase64.value = "";
  editImageLoaded.value = false;
  showEditModal.value = true;

  try {
    // 优先使用云存储 URL，其次使用本地路径
    const imageUrl = template.url || template.imagePath;

    if (
      imageUrl &&
      (imageUrl.startsWith("http://") || imageUrl.startsWith("https://"))
    ) {
      // 云存储 URL，直接使用
      editImageBase64.value = imageUrl;
    } else if (template.imagePath) {
      // 本地路径，调用后端读取
      const base64 = await GetImageBase64(template.imagePath);
      editImageBase64.value = base64;
    } else {
      ElMessage.warning("该模板没有关联的图片");
    }
  } catch (error: any) {
    ElMessage.error(`加载预览失败: ${error.message || error}`);
  }
};

// 编辑图片加载完成
const onEditImageLoad = (event: Event) => {
  const img = event.target as HTMLImageElement;
  console.log("Edit image loaded:", {
    naturalWidth: img.naturalWidth,
    naturalHeight: img.naturalHeight,
    clientWidth: img.clientWidth,
    clientHeight: img.clientHeight,
  });
  editImageNaturalWidth.value = img.naturalWidth;
  editImageNaturalHeight.value = img.naturalHeight;
  editImageLoaded.value = true;

  if (editWidth.value > editImageNaturalWidth.value) {
    editWidth.value = editImageNaturalWidth.value;
  }
  if (editHeight.value > editImageNaturalHeight.value) {
    editHeight.value = editImageNaturalHeight.value;
  }
};

// 编辑图片加载失败
const onEditImageError = () => {
  editImageLoaded.value = false;
  ElMessage.error("图片加载失败");
};

// 校验编辑宽度
const validateEditWidth = () => {
  if (editImageLoaded.value && editWidth.value > editImageNaturalWidth.value) {
    editWidth.value = editImageNaturalWidth.value;
    ElMessage.warning(`宽度不能超过图片宽度 ${editImageNaturalWidth.value}px`);
  }
};

// 校验编辑高度
const validateEditHeight = () => {
  if (
    editImageLoaded.value &&
    editHeight.value > editImageNaturalHeight.value
  ) {
    editHeight.value = editImageNaturalHeight.value;
    ElMessage.warning(`高度不能超过图片高度 ${editImageNaturalHeight.value}px`);
  }
};

// 保存编辑后的模板
const saveEditedTemplate = async () => {
  if (!selectedTemplate.value || !selectedTemplate.value.id) return;

  if (!editName.value.trim()) {
    ElMessage.warning("请输入模板名称");
    return;
  }

  try {
    loading.value = true;

    await apiUpdateTemplate(selectedTemplate.value.id, {
      name: editName.value,
      width: editWidth.value,
      height: editHeight.value,
      scale: editScale.value,
      imagePath: selectedTemplate.value.imagePath,
      offset_x: editOffsetX.value,
      offset_y: editOffsetY.value,
      url: selectedTemplate.value.url || selectedTemplate.value.imagePath,
    });

    showEditModal.value = false;
    ElMessage.success("模板更新成功");
    await loadTemplates();
  } catch (error: any) {
    ElMessage.error(`更新模板失败: ${error.message || error}`);
  } finally {
    loading.value = false;
  }
};

// ========== 新增模板相关方法 ==========

// 监听新增图片路径变化，加载图片
watch(addSelectedImagePath, async (newPath) => {
  if (!newPath) {
    addImageLoaded.value = false;
    addImageNaturalWidth.value = 0;
    addImageNaturalHeight.value = 0;
    addImageBase64.value = "";
    return;
  }

  try {
    const base64Data = await GetImageBase64(newPath);
    addImageBase64.value = base64Data;
  } catch (error: any) {
    ElMessage.error(`加载图片失败: ${error.message || error}`);
    addImageLoaded.value = false;
  }
});

// 新增图片加载完成
const onAddImageLoad = (event: Event) => {
  const img = event.target as HTMLImageElement;
  addImageNaturalWidth.value = img.naturalWidth;
  addImageNaturalHeight.value = img.naturalHeight;
  addImageLoaded.value = true;
  // 自动设置宽高为图片尺寸
  addWidth.value = img.naturalWidth;
  addHeight.value = img.naturalHeight;
};

// 新增图片加载失败
const onAddImageError = () => {
  addImageLoaded.value = false;
  ElMessage.error("图片加载失败");
};

// 新增模板矩形样式
const addRectangleStyle = computed(() => {
  if (!addImageLoaded.value || !addImageElement.value) return;

  const img = addImageElement.value;
  const displayWidth = img.clientWidth;
  const displayHeight = img.clientHeight;

  const scaleX = displayWidth / addImageNaturalWidth.value;
  const scaleY = displayHeight / addImageNaturalHeight.value;

  const rectWidth = addWidth.value * scaleX;
  const rectHeight = addHeight.value * scaleY;

  const baseLeft = (displayWidth - rectWidth) / 2;
  const baseTop = (displayHeight - rectHeight) / 2;

  const dx = (addOffsetX.value || 0) * scaleX;
  const dy = (addOffsetY.value || 0) * scaleY;

  const left = baseLeft + dx;
  const top = baseTop + dy;

  return {
    width: `${rectWidth}px`,
    height: `${rectHeight}px`,
    left: `${left}px`,
    top: `${top}px`,
  };
});

// 验证新增宽度
const validateAddWidth = () => {
  if (addImageLoaded.value && addWidth.value > addImageNaturalWidth.value) {
    addWidth.value = addImageNaturalWidth.value;
    ElMessage.warning(`宽度不能超过图片宽度 ${addImageNaturalWidth.value}px`);
  }
};

// 验证新增高度
const validateAddHeight = () => {
  if (addImageLoaded.value && addHeight.value > addImageNaturalHeight.value) {
    addHeight.value = addImageNaturalHeight.value;
    ElMessage.warning(`高度不能超过图片高度 ${addImageNaturalHeight.value}px`);
  }
};

// 选择图片（新增模板用）
const selectAddImage = async () => {
  try {
    loading.value = true;
    const path = await SelectImage();
    if (path) {
      addSelectedImagePath.value = path;
    }
  } catch (error: any) {
    ElMessage.error(`选择图片失败: ${error.message || error}`);
  } finally {
    loading.value = false;
  }
};

// 打开新增模板弹窗
const openAddModal = () => {
  addName.value = "";
  addWidth.value = 800;
  addHeight.value = 600;
  addScale.value = 1.0;
  addOffsetX.value = 0;
  addOffsetY.value = 0;
  addSelectedImagePath.value = "";
  addImageLoaded.value = false;
  addImageBase64.value = "";
  addImageNaturalWidth.value = 0;
  addImageNaturalHeight.value = 0;
  showAddModal.value = true;
};

// 保存新增模板
const saveAddTemplate = async () => {
  if (!addName.value.trim()) {
    ElMessage.warning("请输入模板名称");
    return;
  }

  if (!addSelectedImagePath.value) {
    ElMessage.warning("请先选择图片");
    return;
  }

  try {
    loading.value = true;

    // 生成云存储的 key
    const ext = addSelectedImagePath.value.split(".").pop() || "png";
    const timestamp = Date.now();
    const key = `templates/${addName.value}_${timestamp}.${ext}`;

    // 获取图片 base64 数据
    const base64Data = await GetImageBase64(
      addSelectedImagePath.value,
    );

    // 从 data URI 提取 MIME type
    const mimeMatch = base64Data.match(/^data:([^;]+);base64,/);
    const contentType = mimeMatch ? mimeMatch[1] : "application/octet-stream";

    // 请求预签名上传链接（带上 Content-Type，确保签名匹配）
    const presignResult = await apiGetPresignUpload({
      bucket: "xhw",
      key: key,
      expire: 3600,
      content_type: contentType,
    });

    // 使用 Go 后端方法上传，绕过 WebKit 网络栈
    await UploadToPresignedURL(presignResult.data.url, base64Data, contentType);

    // 获取云存储路径
    const cloudPath = presignResult.data.url.split("?")[0];

    // 调用 API 创建模板
    await apiCreateTemplate({
      name: addName.value,
      width: addWidth.value,
      height: addHeight.value,
      scale: addScale.value,
      imagePath: cloudPath,
      offset_x: addOffsetX.value,
      offset_y: addOffsetY.value,
      url: cloudPath,
    });

    showAddModal.value = false;
    ElMessage.success("模板添加成功");
    await loadTemplates();
  } catch (error: any) {
    const status = error?.response?.status;
    const detail = error?.response?.data?.message || error?.message || error;
    ElMessage.error(`添加模板失败${status ? "(" + status + ")" : ""}: ${detail}`);
  } finally {
    loading.value = false;
  }
};

// 格式化日期
const formatDate = (dateStr: string) => {
  const date = new Date(dateStr);
  return date.toLocaleString("zh-CN");
};

// 刷新列表（供父组件调用）
const refresh = () => {
  loadTemplates();
};

// 暴露方法给父组件
defineExpose({
  refresh,
});

onMounted(() => {
  loadTemplates();
});
</script>

<template>
  <div class="template-list-page">
    <section class="page-hero">
      <div class="hero-copy">
        <div class="hero-icon">
          <el-icon :size="20">
            <Collection />
          </el-icon>
        </div>
        <div class="hero-text">
          <h2>模板列表</h2>
          <p class="hero-desc">
            维护模板尺寸、缩放比例和偏移配置，确保批量合成输出一致。
          </p>
        </div>
      </div>
      <div class="hero-actions">
        <div class="hero-count-pill">
          共 <strong>{{ templates.length }}</strong> 个模板
        </div>
        <el-button class="hero-button" type="primary" @click="openAddModal">
          <el-icon><Plus /></el-icon>
          新增模板
        </el-button>
      </div>
    </section>

    <!-- 表格 -->
    <section class="table-card">
      <div class="table-card-head">
        <div>
          <h3>全部模板</h3>
          <p>点击查看可预览裁切区域，编辑可直接调整尺寸与偏移量。</p>
        </div>
      </div>

      <el-table :data="templates" v-loading="loading" stripe>
        <el-table-column label="图片" width="120" align="center">
          <template #default="{ row }">
            <el-image
              class="template-thumb"
              :src="row.url || row.imagePath"
              :alt="row.name"
              fit="cover"
              :preview-src-list="[row.url || row.imagePath]"
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

        <el-table-column prop="id" label="ID" width="96" align="center" />

        <el-table-column
          prop="name"
          label="模板名称"
          min-width="150"
          align="center"
        >
          <template #default="{ row }">
            <span class="template-name">{{ row.name }}</span>
          </template>
        </el-table-column>

        <el-table-column label="画布尺寸" width="150" align="center">
          <template #default="{ row }">
            <span class="size-chip">{{ row.width }} × {{ row.height }} px</span>
          </template>
        </el-table-column>

        <el-table-column prop="scale" label="缩放" width="100" align="center">
          <template #default="{ row }"> {{ row.scale }}x </template>
        </el-table-column>

        <el-table-column label="偏移量" width="160" align="center">
          <template #default="{ row }">
            X: {{ getOffsetX(row) }} / Y: {{ getOffsetY(row) }}
          </template>
        </el-table-column>

        <el-table-column label="创建时间" width="180" align="center">
          <template #default="{ row }">
            {{ row.createdAt ? formatDate(row.createdAt) : "-" }}
          </template>
        </el-table-column>

        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button
              class="row-action"
              type="primary"
              size="small"
              @click="viewTemplate(row)"
            >
              查看
            </el-button>
            <el-button
              class="row-action"
              type="warning"
              size="small"
              @click="openEditTemplate(row)"
            >
              编辑
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
    </section>

    <!-- 查看模板弹窗 -->
    <el-dialog v-model="showViewModal" title="查看模板" width="80%">
      <div class="modal-content" v-if="selectedTemplate">
        <div class="modal-preview-wrapper">
          <div class="modal-preview">
            <div v-if="previewImageBase64" class="preview-wrapper">
              <img
                ref="viewImageElement"
                :src="previewImageBase64"
                alt="模板预览"
                @load="onViewImageLoad"
                @error="onViewImageError"
              />
              <div
                v-if="viewImageLoaded"
                class="rectangle-overlay"
                :style="viewRectangleStyle"
              ></div>
            </div>
            <div v-else class="preview-placeholder">正在加载图片预览...</div>
          </div>
          <div v-if="viewImageLoaded" class="image-info">
            <p>
              图片尺寸: {{ viewImageNaturalWidth }} ×
              {{ viewImageNaturalHeight }} px
            </p>
            <p>
              选择区域: {{ selectedTemplate.width }} ×
              {{ selectedTemplate.height }} px
            </p>
          </div>
        </div>
        <div class="modal-details">
          <el-descriptions :column="1" border>
            <el-descriptions-item label="模板名称">{{
              selectedTemplate.name
            }}</el-descriptions-item>
            <el-descriptions-item label="尺寸"
              >{{ selectedTemplate.width }} ×
              {{ selectedTemplate.height }} px</el-descriptions-item
            >
            <el-descriptions-item label="偏移量">
              X
              {{ getOffsetX(selectedTemplate) }} / Y
              {{ getOffsetY(selectedTemplate) }}
              px
            </el-descriptions-item>
            <el-descriptions-item label="缩放"
              >{{ selectedTemplate.scale }}x</el-descriptions-item
            >
            <el-descriptions-item label="创建时间">
              {{
                selectedTemplate.createdAt
                  ? formatDate(selectedTemplate.createdAt)
                  : "-"
              }}
            </el-descriptions-item>
          </el-descriptions>
        </div>
      </div>
    </el-dialog>

    <!-- 编辑模板弹窗 -->
    <el-dialog v-model="showEditModal" title="编辑模板" width="80%">
      <div class="modal-content" v-if="selectedTemplate">
        <div class="modal-preview-wrapper">
          <div class="modal-preview">
            <div v-if="editImageBase64" class="preview-wrapper">
              <img
                ref="editImageElement"
                :src="editImageBase64"
                alt="预览图片"
                @load="onEditImageLoad"
                @error="onEditImageError"
              />
              <div
                v-if="editImageLoaded"
                class="rectangle-overlay"
                :style="editRectangleStyle"
              ></div>
            </div>
            <div v-else class="preview-placeholder">正在加载图片预览...</div>
          </div>
          <div v-if="editImageLoaded" class="image-info">
            <p>
              图片尺寸: {{ editImageNaturalWidth }} ×
              {{ editImageNaturalHeight }} px
            </p>
            <p>选择区域: {{ editWidth }} × {{ editHeight }} px</p>
          </div>
        </div>

        <div class="modal-details">
          <el-form label-width="80px">
            <el-form-item label="模板名称">
              <el-input v-model="editName" placeholder="请输入模板名称" />
            </el-form-item>
            <el-form-item label="宽度">
              <el-input-number
                v-model="editWidth"
                :min="1"
                @change="validateEditWidth"
              />
            </el-form-item>
            <el-form-item label="高度">
              <el-input-number
                v-model="editHeight"
                :min="1"
                @change="validateEditHeight"
              />
            </el-form-item>
            <el-form-item label="缩放">
              <el-input-number v-model="editScale" :min="0.1" :step="0.1" />
            </el-form-item>
            <el-form-item label="偏移量">
              <div class="offset-inputs">
                <span>X:</span>
                <el-input-number v-model="editOffsetX" size="small" />
                <span>Y:</span>
                <el-input-number v-model="editOffsetY" size="small" />
              </div>
            </el-form-item>
          </el-form>
        </div>
      </div>
      <template #footer>
        <el-button @click="showEditModal = false">取消</el-button>
        <el-button
          type="primary"
          @click="saveEditedTemplate"
          :loading="loading"
        >
          保存修改
        </el-button>
      </template>
    </el-dialog>

    <!-- 新增模板弹窗 -->
    <el-dialog v-model="showAddModal" title="新增模板" width="80%">
      <div class="modal-content">
        <div class="modal-preview-wrapper">
          <div class="modal-preview">
            <div v-if="!addSelectedImagePath" class="preview-placeholder">
              <p>请先选择图片</p>
              <el-button
                type="primary"
                @click="selectAddImage"
                :loading="loading"
              >
                选择图片
              </el-button>
            </div>
            <div v-else-if="addImageBase64" class="preview-wrapper">
              <img
                ref="addImageElement"
                :src="addImageBase64"
                alt="预览图片"
                @load="onAddImageLoad"
                @error="onAddImageError"
              />
              <div
                v-if="addImageLoaded"
                class="rectangle-overlay"
                :style="addRectangleStyle"
              ></div>
            </div>
          </div>
          <div v-if="addImageLoaded" class="image-info">
            <p>
              图片尺寸: {{ addImageNaturalWidth }} ×
              {{ addImageNaturalHeight }} px
            </p>
            <p>选择区域: {{ addWidth }} × {{ addHeight }} px</p>
          </div>
        </div>

        <div class="modal-details">
          <el-form label-width="80px">
            <el-form-item label="选择图片">
              <el-button
                type="primary"
                size="small"
                @click="selectAddImage"
                :loading="loading"
              >
                选择图片
              </el-button>
              <div v-if="addSelectedImagePath" class="selected-path">
                {{ addSelectedImagePath }}
              </div>
            </el-form-item>
            <el-form-item label="模板名称">
              <el-input v-model="addName" placeholder="请输入模板名称" />
            </el-form-item>
            <el-form-item label="宽度">
              <el-input-number
                v-model="addWidth"
                :min="1"
                :max="addImageNaturalWidth || undefined"
                @change="validateAddWidth"
              />
              <span v-if="addImageLoaded" class="input-hint"
                >最大: {{ addImageNaturalWidth }}px</span
              >
            </el-form-item>
            <el-form-item label="高度">
              <el-input-number
                v-model="addHeight"
                :min="1"
                :max="addImageNaturalHeight || undefined"
                @change="validateAddHeight"
              />
              <span v-if="addImageLoaded" class="input-hint"
                >最大: {{ addImageNaturalHeight }}px</span
              >
            </el-form-item>
            <el-form-item label="缩放">
              <el-input-number v-model="addScale" :min="0.1" :step="0.1" />
            </el-form-item>
            <el-form-item label="偏移量">
              <div class="offset-inputs">
                <span>X:</span>
                <el-input-number
                  v-model="addOffsetX"
                  size="small"
                  :disabled="!addImageLoaded"
                />
                <span>Y:</span>
                <el-input-number
                  v-model="addOffsetY"
                  size="small"
                  :disabled="!addImageLoaded"
                />
              </div>
              <span class="input-hint">X 向右为正，Y 向下为正</span>
            </el-form-item>
          </el-form>
        </div>
      </div>
      <template #footer>
        <el-button @click="showAddModal = false">取消</el-button>
        <el-button
          type="primary"
          @click="saveAddTemplate"
          :loading="loading"
          :disabled="!addImageLoaded"
        >
          添加模板
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.template-list-page {
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
  border: 1px solid var(--apple-border-soft);
  border-radius: 14px;
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

.hero-button {
  height: 34px;
  padding: 0 14px;
  border-radius: 7px;
  font-weight: 600;
}

.table-card {
  padding: 14px;
  border: 1px solid var(--apple-border-soft);
  border-radius: 14px;
  background: var(--apple-surface);
  box-shadow: var(--apple-shadow-soft);
}

.table-card-head {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 12px;
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

.template-list-page :deep(.el-table) {
  width: 100%;
  border-radius: 10px;
  overflow: hidden;
}

.template-thumb {
  width: 80px;
  height: 80px;
  border-radius: 10px;
  border: 1px solid var(--apple-border-soft);
}

.template-name {
  color: var(--apple-text);
  font-size: 14px;
  font-weight: 600;
}

.size-chip {
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

.image-placeholder {
  width: 80px;
  height: 80px;
  display: flex;
  align-items: center;
  justify-content: center;
  background-color: #fbfbfd;
  color: var(--apple-text-muted);
  font-size: 24px;
  border-radius: 10px;
  border: 1px solid var(--apple-border-soft);
}

.modal-content {
  display: grid;
  grid-template-columns: 2fr 1fr;
  gap: 18px;
}

.modal-preview {
  border: 1px solid var(--apple-border-soft);
  border-radius: 12px;
  overflow: hidden;
  background-color: #fbfbfd;
}

.modal-preview-wrapper {
  display: flex;
  flex-direction: column;
}

.modal-preview-wrapper .image-info {
  margin-top: 8px;
}

.preview-wrapper {
  position: relative;
  max-width: 100%;
  max-height: 400px;
  overflow: auto;
}

.preview-wrapper img {
  display: block;
  max-width: 100%;
  max-height: 400px;
  height: auto;
}

.preview-placeholder {
  min-height: 260px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 12px;
  color: var(--apple-text-muted);
  font-size: 14px;
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

.input-hint {
  display: block;
  margin-top: 4px;
  font-size: 12px;
  color: var(--apple-text-muted);
}

.rectangle-overlay {
  position: absolute;
  background-color: rgba(0, 113, 227, 0.12);
  border: 2px solid rgba(0, 113, 227, 0.9);
  pointer-events: none;
}

.image-info {
  margin-top: 12px;
  padding: 10px 12px;
  background-color: #fbfbfd;
  border-radius: 10px;
  border: 1px solid var(--apple-border-soft);
  font-size: 13px;
  color: var(--apple-text-secondary);
}

.image-info p {
  margin: 4px 0;
}

.offset-inputs {
  display: flex;
  align-items: center;
  gap: 10px;
}

.offset-inputs span {
  color: var(--apple-text-secondary);
  font-size: 13px;
  font-weight: 600;
}

@media (max-width: 760px) {
  .template-list-page {
    padding: 12px;
  }

  .page-hero {
    flex-direction: column;
    align-items: flex-start;
    padding: 14px;
  }

  .hero-actions {
    width: 100%;
    flex-wrap: wrap;
  }

  .hero-button {
    flex: 1;
  }

  .table-card {
    padding: 12px;
  }

  .modal-content {
    grid-template-columns: 1fr;
  }
}
</style>
