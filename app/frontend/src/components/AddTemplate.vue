<script setup lang="ts">
import { ref, computed, watch } from "vue";
import { ImageService } from "../../bindings/changeme/services";
import { apiCreateTemplate } from "../api/template";
import { apiGetPresignUpload } from "../api/presign";
import axios from "axios";

const emit = defineEmits<{
  templateAdded: [];
}>();

const templateName = ref("");
const width = ref(800);
const height = ref(600);
const scale = ref(1.0);
// 框偏移量：X（水平，向右为正）、Y（垂直，向下为正）
const offsetX = ref(0);
const offsetY = ref(0);
const selectedImagePath = ref("");
const loading = ref(false);
const message = ref("");

// 图片相关状态
const imageElement = ref<HTMLImageElement | null>(null);
const imageNaturalWidth = ref(0);
const imageNaturalHeight = ref(0);
const imageLoaded = ref(false);
const imageBase64 = ref("");

// 监听图片路径变化，加载图片
watch(selectedImagePath, async (newPath) => {
  if (!newPath) {
    imageLoaded.value = false;
    imageNaturalWidth.value = 0;
    imageNaturalHeight.value = 0;
    imageBase64.value = "";
    return;
  }

  try {
    // 清除之前的错误消息
    if (
      message.value === "图片加载失败，请重新选择" ||
      message.value.includes("加载图片失败")
    ) {
      message.value = "";
    }

    // 调用后端方法获取 base64 图片
    const base64Data = await ImageService.GetImageBase64(newPath);
    imageBase64.value = base64Data;
  } catch (error: any) {
    message.value = `加载图片失败: ${error.message || error}`;
    imageLoaded.value = false;
  }
});

// 图片加载完成
const onImageLoad = (event: Event) => {
  const img = event.target as HTMLImageElement;
  imageNaturalWidth.value = img.naturalWidth;
  imageNaturalHeight.value = img.naturalHeight;
  imageLoaded.value = true;

  // 自动设置宽高为图片尺寸
  width.value = img.naturalWidth;
  height.value = img.naturalHeight;

  // 清除错误消息
  if (message.value === "图片加载失败，请重新选择") {
    message.value = "";
  }
};

// 图片加载失败
const onImageError = (event: Event) => {
  console.error("图片加载失败", event);
  message.value = "图片加载失败，请重新选择";
  imageLoaded.value = false;
};

// 计算矩形在预览图中的位置和大小
const rectangleStyle = computed(() => {
  if (!imageLoaded.value || !imageElement.value) return;

  const img = imageElement.value;
  const displayWidth = img.clientWidth;
  const displayHeight = img.clientHeight;

  // 计算缩放比例
  const scaleX = displayWidth / imageNaturalWidth.value;
  const scaleY = displayHeight / imageNaturalHeight.value;

  // 计算矩形的显示尺寸
  const rectWidth = width.value * scaleX;
  const rectHeight = height.value * scaleY;

  // 居中基础位置
  const baseLeft = (displayWidth - rectWidth) / 2;
  const baseTop = (displayHeight - rectHeight) / 2;

  // 偏移量：X 向右为正，Y 向下为正
  const dx = offsetX.value * scaleX;
  const dy = offsetY.value * scaleY;

  const left = baseLeft + dx;
  const top = baseTop + dy;

  return {
    width: `${rectWidth}px`,
    height: `${rectHeight}px`,
    left: `${left}px`,
    top: `${top}px`,
  };
});

// 验证宽度
const validateWidth = () => {
  if (imageLoaded.value && width.value > imageNaturalWidth.value) {
    width.value = imageNaturalWidth.value;
    message.value = `宽度不能超过图片宽度 ${imageNaturalWidth.value}px`;
  }
};

// 验证高度
const validateHeight = () => {
  if (imageLoaded.value && height.value > imageNaturalHeight.value) {
    height.value = imageNaturalHeight.value;
    message.value = `高度不能超过图片高度 ${imageNaturalHeight.value}px`;
  }
};

// 选择图片
const selectImage = async () => {
  try {
    loading.value = true;
    message.value = "";
    const path = await ImageService.SelectImage();
    if (path) {
      selectedImagePath.value = path;
      message.value = `已选择: ${path}`;
    }
  } catch (error: any) {
    message.value = `选择图片失败: ${error.message || error}`;
  } finally {
    loading.value = false;
  }
};

// 添加模板（包含上传到云存储）
const addTemplate = async () => {
  if (!templateName.value.trim()) {
    message.value = "请输入模板名称";
    return;
  }

  if (!selectedImagePath.value) {
    message.value = "请先选择图片";
    return;
  }

  try {
    loading.value = true;
    message.value = "正在获取上传链接...";

    // 生成云存储的 key
    const ext = selectedImagePath.value.split(".").pop() || "png";
    const timestamp = Date.now();
    const key = `templates/${templateName.value}_${timestamp}.${ext}`;

    // 请求预签名上传链接
    const presignResult = await apiGetPresignUpload({
      bucket: "xhw",
      key: key,
      expire: 3600,
    });

    message.value = "正在上传图片...";

    // 获取图片 base64 数据
    const base64Data = await ImageService.GetImageBase64(
      selectedImagePath.value,
    );
    // 将 base64 转换为 Blob
    const base64Content = base64Data.split(",")[1]; // 去除 data:image/xxx;base64, 前缀
    const byteCharacters = atob(base64Content);
    const byteNumbers = new Array(byteCharacters.length);
    for (let i = 0; i < byteCharacters.length; i++) {
      byteNumbers[i] = byteCharacters.charCodeAt(i);
    }
    const byteArray = new Uint8Array(byteNumbers);
    const blob = new Blob([byteArray]);

    // 使用预签名 URL 上传
    await axios.put(presignResult.data.url, blob, {
      headers: {
        "Content-Type": "application/octet-stream",
      },
    });

    // 获取云存储路径
    const cloudPath = presignResult.data.url.split("?")[0]; // 去除查询参数，保留基础 URL

    message.value = "正在添加模板...";

    // 调用 API 创建模板
    await apiCreateTemplate({
      name: templateName.value,
      width: width.value,
      height: height.value,
      scale: scale.value,
      imagePath: cloudPath,
      offset_x: offsetX.value,
      offset_y: offsetY.value,
      url: cloudPath,
    });

    message.value = "模板添加成功";

    // 重置表单
    templateName.value = "";
    width.value = 800;
    height.value = 600;
    scale.value = 1.0;
    offsetX.value = 0;
    offsetY.value = 0;
    selectedImagePath.value = "";
    imageLoaded.value = false;
    imageBase64.value = "";

    // 通知父组件刷新列表
    emit("templateAdded");
  } catch (error: any) {
    console.error("添加模板失败:", error);
    message.value = `添加模板失败: ${error.message || error}`;
  } finally {
    loading.value = false;
  }
};
</script>

<template>
  <div class="add-template">
    <!-- 消息提示 -->
    <div
      v-if="message"
      class="message"
      :class="{ error: message.includes('失败') }"
    >
      {{ message }}
    </div>

    <div class="content-wrapper">
      <!-- 左侧：图片预览区域 -->
      <div class="preview-section">
        <h2>图片预览</h2>
        <div class="preview-container">
          <div v-if="!selectedImagePath" class="preview-placeholder">
            <p>请先选择图片</p>
          </div>
          <div v-else class="preview-image-wrapper">
            <img
              ref="imageElement"
              :src="imageBase64"
              @load="onImageLoad"
              @error="onImageError"
              class="preview-image"
              alt="预览图片"
            />
            <!-- 矩形覆盖层 -->
            <div
              v-if="imageLoaded"
              class="rectangle-overlay"
              :style="rectangleStyle"
            ></div>
          </div>
          <div v-if="imageLoaded" class="image-info">
            <p>
              图片尺寸: {{ imageNaturalWidth }} × {{ imageNaturalHeight }} px
            </p>
            <p>选择区域: {{ width }} × {{ height }} px</p>
          </div>
        </div>
      </div>

      <!-- 右侧：表单区域 -->
      <div class="form-section">
        <h2>添加新模板</h2>

        <div class="form-group">
          <label>选择图片:</label>
          <div class="image-select">
            <button
              @click="selectImage"
              :disabled="loading"
              class="btn btn-primary"
            >
              选择图片
            </button>
          </div>
          <div v-if="selectedImagePath" class="selected-path">
            {{ selectedImagePath }}
          </div>
        </div>

        <div class="form-group">
          <label for="templateName">模板名称:</label>
          <input
            id="templateName"
            v-model="templateName"
            type="text"
            placeholder="请输入模板名称"
            :disabled="loading"
          />
        </div>

        <div class="form-group">
          <label for="width">宽度 (px):</label>
          <input
            id="width"
            v-model.number="width"
            type="number"
            min="1"
            :max="imageNaturalWidth || undefined"
            :disabled="loading || !imageLoaded"
            @blur="validateWidth"
            @input="validateWidth"
          />
          <span v-if="imageLoaded" class="input-hint"
            >最大: {{ imageNaturalWidth }}px</span
          >
        </div>

        <div class="form-group">
          <label for="height">高度 (px):</label>
          <input
            id="height"
            v-model.number="height"
            type="number"
            min="1"
            :max="imageNaturalHeight || undefined"
            :disabled="loading || !imageLoaded"
            @blur="validateHeight"
            @input="validateHeight"
          />
          <span v-if="imageLoaded" class="input-hint"
            >最大: {{ imageNaturalHeight }}px</span
          >
        </div>

        <div class="form-group">
          <label for="scale">缩放比例:</label>
          <input
            id="scale"
            v-model.number="scale"
            type="number"
            step="0.1"
            min="0.1"
            :disabled="loading"
          />
        </div>

        <div class="form-group">
          <label>框偏移量 (px):</label>
          <div class="offset-grid">
            <div class="offset-item">
              <span class="offset-label">X:</span>
              <input
                v-model.number="offsetX"
                type="number"
                :disabled="loading || !imageLoaded"
              />
            </div>
            <div class="offset-item">
              <span class="offset-label">Y:</span>
              <input
                v-model.number="offsetY"
                type="number"
                :disabled="loading || !imageLoaded"
              />
            </div>
          </div>
          <span class="input-hint"> X 向右为正，Y 向下为正（单位: px）。 </span>
        </div>

        <button
          @click="addTemplate"
          :disabled="loading"
          class="btn btn-success btn-large"
        >
          {{ loading ? "处理中..." : "添加模板" }}
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.add-template {
  display: block;
  width: 100%;
  box-sizing: border-box;
  padding: 20px;
}

h2 {
  color: #34495e;
  margin-bottom: 20px;
  font-size: 1.5em;
}

.message {
  padding: 12px 20px;
  margin-bottom: 20px;
  border-radius: 6px;
  background-color: #d4edda;
  color: #155724;
  border: 1px solid #c3e6cb;
}

.message.error {
  background-color: #f8d7da;
  color: #721c24;
  border-color: #f5c6cb;
}

.content-wrapper {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 30px;
}

/* 预览区域样式 */
.preview-section {
  background: #fff;
  padding: 30px;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.preview-container {
  width: 100%;
}

.preview-placeholder {
  width: 100%;
  height: 400px;
  display: flex;
  align-items: center;
  justify-content: center;
  background-color: #f8f9fa;
  border: 2px dashed #ddd;
  border-radius: 8px;
  color: #999;
  font-size: 16px;
}

.preview-image-wrapper {
  position: relative;
  width: 100%;
  display: inline-block;
  border: 1px solid #e0e0e0;
  border-radius: 8px;
  overflow: hidden;
}

.preview-image {
  width: 100%;
  height: auto;
  display: block;
}

.rectangle-overlay {
  position: absolute;
  background-color: rgba(135, 206, 250, 0.4);
  border: 2px solid rgba(30, 144, 255, 0.8);
  pointer-events: none;
  box-shadow: 0 0 10px rgba(30, 144, 255, 0.3);
}

.image-info {
  margin-top: 15px;
  padding: 12px;
  background-color: #f8f9fa;
  border-radius: 6px;
  font-size: 14px;
  color: #555;
}

.image-info p {
  margin: 5px 0;
}

/* 表单区域样式 */
.form-section {
  background: #fff;
  padding: 30px;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.form-group {
  margin-bottom: 20px;
}

.form-group label {
  display: block;
  margin-bottom: 8px;
  font-weight: 600;
  color: #555;
}

.form-group input {
  width: 100%;
  padding: 10px 15px;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 14px;
  transition: border-color 0.3s;
}

.form-group input:focus {
  outline: none;
  border-color: #4caf50;
}

.form-group input:disabled {
  background-color: #f5f5f5;
  cursor: not-allowed;
}

.input-hint {
  display: block;
  margin-top: 5px;
  font-size: 12px;
  color: #999;
}

.image-select {
  display: flex;
  gap: 10px;
  margin-bottom: 10px;
}

.selected-path {
  padding: 8px 12px;
  background-color: #f8f9fa;
  border-radius: 4px;
  font-size: 13px;
  color: #666;
  word-break: break-all;
}

.btn {
  padding: 10px 20px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
  font-weight: 500;
  transition: all 0.3s;
}

.btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.btn-primary {
  background-color: #007bff;
  color: white;
}

.btn-primary:hover:not(:disabled) {
  background-color: #0056b3;
}

.btn-success {
  background-color: #28a745;
  color: white;
}

.btn-success:hover:not(:disabled) {
  background-color: #218838;
}

.btn-large {
  width: 100%;
  padding: 12px;
  font-size: 16px;
}

.offset-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 10px 16px;
}

.offset-item {
  display: flex;
  align-items: center;
  gap: 6px;
}

.offset-label {
  min-width: 24px;
  color: #555;
}

.offset-item input {
  flex: 1;
  padding: 6px 10px;
}
</style>
