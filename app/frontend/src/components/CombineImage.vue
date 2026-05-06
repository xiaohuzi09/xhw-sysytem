<script setup lang="ts">
import { ref, computed, onMounted, watch } from "vue";
import { GetImageBase64, CombineImagesWithTemplates, SelectImages } from "@wailsjs/go/services/ImageService";
import { services } from "@wailsjs/go/models";
import { apiGetTemplates } from "../api/template";
import type { Template } from "../api/template";

// 获取模板的 X 偏移值（兼容 API 和本地存储两种格式）
const getOffsetX = (template: Template): number => {
  if (template.offset_x !== undefined) {
    return template.offset_x;
  }
  return (template.offsetLeft || 0) - (template.offsetRight || 0);
};

// 获取模板的 Y 偏移值（兼容 API 和本地存储两种格式）
const getOffsetY = (template: Template): number => {
  if (template.offset_y !== undefined) {
    return template.offset_y;
  }
  return (template.offsetTop || 0) - (template.offsetBottom || 0);
};

const templates = ref<Template[]>([]);
// 支持多选模板：用于实际合成
const selectedTemplateIds = ref<string[]>([]);
// 当前用于预览的模板 ID（从所选模板中切换）
const previewTemplateId = ref("");
// 当前用于预览的模板
const selectedTemplate = computed(
  () => templates.value.find((t) => t.id === previewTemplateId.value) || null,
);

const loading = ref(false);
const combining = ref(false);
const message = ref("");

// 模板底图
const templateBase64 = ref("");
const templateNaturalWidth = ref(0);
const templateNaturalHeight = ref(0);

// 要放入框中的图片（支持多选）
const selectedImagePaths = ref<string[]>([]);
const selectedImagePath = ref(""); // 当前用于预览的第一张
const imageBase64 = ref("");
const imageNaturalWidth = ref(0);
const imageNaturalHeight = ref(0);

const compositeCanvas = ref<HTMLCanvasElement | null>(null);

// 裁剪图片四周的空白区域（白色或透明），只保留有效像素
const trimImageCanvas = (img: HTMLImageElement): HTMLCanvasElement => {
  const canvas = document.createElement("canvas");
  const ctx = canvas.getContext("2d");
  // 先绘制原图到临时 canvas 获取像素数据
  const tempCanvas = document.createElement("canvas");
  const tempCtx = tempCanvas.getContext("2d");

  if (!ctx || !tempCtx) {
    // 无法创建 context，返回原图
    canvas.width = img.width;
    canvas.height = img.height;
    const fallbackCtx = canvas.getContext("2d");
    if (fallbackCtx) {
      fallbackCtx.drawImage(img, 0, 0);
    }
    return canvas;
  }

  tempCanvas.width = img.width;
  tempCanvas.height = img.height;
  tempCtx.drawImage(img, 0, 0);

  const imageData = tempCtx.getImageData(0, 0, img.width, img.height);
  const data = imageData.data;

  // 判断像素是否为空白（白色或透明）
  const isBlank = (x: number, y: number): boolean => {
    const idx = (y * img.width + x) * 4;
    const r = data[idx];
    const g = data[idx + 1];
    const b = data[idx + 2];
    const a = data[idx + 3];
    // 如果是完全透明，认为是空白
    if (a === 0) return true;
    // 如果是白色（或接近白色，阈值 250），认为是空白
    if (a >= 250 && r >= 250 && g >= 250 && b >= 250) return true;
    return false;
  };

  const width = img.width;
  const height = img.height;

  // 从上边找非空白行
  let top = 0;
  for (; top < height; top++) {
    let blankRow = true;
    for (let x = 0; x < width; x++) {
      if (!isBlank(x, top)) {
        blankRow = false;
        break;
      }
    }
    if (!blankRow) break;
  }

  // 从下边找非空白行
  let bottom = height - 1;
  for (; bottom >= top; bottom--) {
    let blankRow = true;
    for (let x = 0; x < width; x++) {
      if (!isBlank(x, bottom)) {
        blankRow = false;
        break;
      }
    }
    if (!blankRow) break;
  }

  // 从左边找非空白列
  let left = 0;
  for (; left < width; left++) {
    let blankCol = true;
    for (let y = top; y <= bottom; y++) {
      if (!isBlank(left, y)) {
        blankCol = false;
        break;
      }
    }
    if (!blankCol) break;
  }

  // 从右边找非空白列
  let right = width - 1;
  for (; right >= left; right--) {
    let blankCol = true;
    for (let y = top; y <= bottom; y++) {
      if (!isBlank(right, y)) {
        blankCol = false;
        break;
      }
    }
    if (!blankCol) break;
  }

  // 如果没有有效区域，返回原图
  if (top > bottom || left > right) {
    canvas.width = img.width;
    canvas.height = img.height;
    ctx.drawImage(img, 0, 0);
    return canvas;
  }

  // 裁剪出有效区域
  const croppedWidth = right - left + 1;
  const croppedHeight = bottom - top + 1;
  canvas.width = croppedWidth;
  canvas.height = croppedHeight;
  ctx.drawImage(
    tempCanvas,
    left,
    top,
    croppedWidth,
    croppedHeight,
    0,
    0,
    croppedWidth,
    croppedHeight,
  );

  return canvas;
};

// 加载模板列表
const loadTemplates = async () => {
  try {
    loading.value = true;
    const result = await apiGetTemplates();
    // 确保 id 是字符串类型
    templates.value = (result.data || []).map((t) => ({
      ...t,
      id: t.id !== undefined ? String(t.id) : undefined,
    }));
  } catch (error: any) {
    message.value = `加载模板失败: ${error.message || error}`;
  } finally {
    loading.value = false;
  }
};

// 当预览模板变化时重新加载模板图片
watch(
  () => previewTemplateId.value,
  async (id) => {
    if (!id) {
      templateBase64.value = "";
      templateNaturalWidth.value = 0;
      templateNaturalHeight.value = 0;
      return;
    }

    const template = templates.value.find((t) => t.id === id);
    if (!template) return;

    try {
      const imageUrl = template.url || template.imagePath;

      if (!imageUrl) {
        message.value = "该模板没有关联的图片";
        return;
      }

      console.log("加载预览模板图片:", imageUrl);

      // 优先直接加载 URL（更快），失败时再通过后端加载
      let imageSrc: string;
      const isUrl =
        imageUrl.startsWith("http://") || imageUrl.startsWith("https://");

      if (isUrl) {
        // 尝试直接加载 URL
        try {
          imageSrc = await new Promise<string>((resolve, reject) => {
            const img = new Image();
            img.crossOrigin = "anonymous";
            img.onload = () => resolve(imageUrl);
            img.onerror = () => reject(new Error("跨域加载失败"));
            img.src = imageUrl;
          });
        } catch {
          // 跨域失败，通过后端加载
          console.log("跨域加载失败，通过后端加载");
          imageSrc = await GetImageBase64(imageUrl);
        }
      } else {
        // 本地路径，通过后端加载
        imageSrc = await GetImageBase64(imageUrl);
      }

      console.log("预览图片加载成功");

      templateBase64.value = imageSrc;

      const img = new Image();
      if (isUrl) {
        img.crossOrigin = "anonymous";
      }
      img.onload = () => {
        templateNaturalWidth.value = img.width;
        templateNaturalHeight.value = img.height;
        drawComposite();
      };
      img.onerror = () => {
        message.value = "加载模板图片失败，请检查模板图片路径";
      };
      img.src = imageSrc;
    } catch (error: any) {
      console.error("加载预览模板图片错误:", error);
      message.value = `加载模板图片失败: ${error.message || error}`;
    }
  },
);

// 当选择模板列表变化时，更新预览模板 ID
watch(
  () => [...selectedTemplateIds.value],
  (ids) => {
    if (ids.length === 0) {
      previewTemplateId.value = "";
      return;
    }

    // 如果当前预览的模板不在选中列表中，则切到第一个
    if (!ids.includes(previewTemplateId.value)) {
      previewTemplateId.value = ids[0];
    }
  },
);

// 选择要合成的图片（多选）
const selectImages = async () => {
  try {
    loading.value = true;
    message.value = "";
    const paths: string[] = await SelectImages();
    if (!paths || paths.length === 0) return;

    selectedImagePaths.value = paths;
    selectedImagePath.value = paths[0];

    const base64 = await GetImageBase64(selectedImagePath.value);
    imageBase64.value = base64;

    // 读取原图尺寸
    const img = new Image();
    img.onload = () => {
      imageNaturalWidth.value = img.width;
      imageNaturalHeight.value = img.height;
      drawComposite();
    };
    img.onerror = () => {
      message.value = "加载图片失败，请重新选择";
    };
    img.src = base64;
  } catch (error: any) {
    message.value = `选择图片失败: ${error.message || error}`;
  } finally {
    loading.value = false;
  }
};

// 绘制合成结果到画布：
// 1. 模板图片作为底图
// 2. 选择的图片按模板框尺寸等比缩放并裁剪，放进模板中间的框里（包含偏移量）
const drawComposite = () => {
  const template = selectedTemplate.value;
  if (
    !template ||
    !imageBase64.value ||
    !templateBase64.value ||
    !compositeCanvas.value
  )
    return;

  const canvas = compositeCanvas.value;
  const ctx = canvas.getContext("2d");
  if (!ctx) return;

  const bgImg = new Image();
  bgImg.onload = () => {
    const tw = bgImg.width;
    const th = bgImg.height;
    if (!tw || !th) return;

    // 画布尺寸为模板底图尺寸
    canvas.width = tw;
    canvas.height = th;

    // 先绘制模板底图
    ctx.clearRect(0, 0, canvas.width, canvas.height);
    ctx.drawImage(bgImg, 0, 0, tw, th);

    const fgImg = new Image();
    fgImg.onload = () => {
      // 裁剪素材图片四周的空白区域
      const trimmedCanvas = trimImageCanvas(fgImg);
      const iw = trimmedCanvas.width;
      const ih = trimmedCanvas.height;
      if (!iw || !ih) return;

      // 模板框区域（居中 + 偏移）
      const rectW = template.width;
      const rectH = template.height;
      let rectX = (tw - rectW) / 2;
      let rectY = (th - rectH) / 2;

      // 偏移：与后端一致，使用 offset_x 和 offset_y
      rectX += getOffsetX(template);
      rectY += getOffsetY(template);

      // 等比缩放（contain）：整张图片缩放到"完全放入"框内（不裁剪，可能留边）
      const scale = Math.min(rectW / iw, rectH / ih);
      const dw = iw * scale;
      const dh = ih * scale;
      const dx = rectX + (rectW - dw) / 2;
      const dy = rectY + (rectH - dh) / 2;

      // 旋转：以素材图片中心为原点旋转
      const rotDeg = template.rotation || 0;
      if (rotDeg !== 0) {
        const rad = (rotDeg * Math.PI) / 180;
        ctx.save();
        ctx.translate(dx + dw / 2, dy + dh / 2);
        ctx.rotate(rad);
        ctx.drawImage(trimmedCanvas, 0, 0, iw, ih, -dw / 2, -dh / 2, dw, dh);
        ctx.restore();
      } else {
        ctx.drawImage(trimmedCanvas, 0, 0, iw, ih, dx, dy, dw, dh);
      }
    };
    fgImg.onerror = () => {
      message.value = "合成失败，无法读取图片数据";
    };
    fgImg.src = imageBase64.value;
  };
  bgImg.onerror = () => {
    message.value = "合成失败，无法读取模板图片数据";
  };
  bgImg.src = templateBase64.value;
};

// 当预览模板或图片变化时重新绘制
watch([() => previewTemplateId.value, imageBase64, templateBase64], () => {
  drawComposite();
});

// 开始批量合成，结果保存在当前目录的“合成/批次时间戳”目录下
const startCombine = async () => {
  if (!selectedTemplateIds.value.length) {
    message.value = "请先选择至少一个模板";
    return;
  }
  if (!selectedImagePaths.value.length) {
    message.value = "请先选择要合成的图片";
    return;
  }

  try {
    loading.value = true;
    combining.value = true;
    message.value = "";

    // 获取选中模板的完整信息
    const selectedTemplates: services.TemplateInfo[] = templates.value
      .filter((t) => t.id && selectedTemplateIds.value.includes(t.id))
      .map((t) => ({
        id: t.id!,
        name: t.name,
        width: t.width,
        height: t.height,
        scale: t.scale,
        imagePath: t.imagePath,
        url: t.url || "",
        offset_x: getOffsetX(t),
        offset_y: getOffsetY(t),
        rotation: t.rotation || 0,
      }));

    // 将图片路径转换为 MaterialInfo 格式（本地文件没有编号，使用文件名）
    const materials: services.MaterialInfo[] = selectedImagePaths.value.map((path) => {
      // 从路径中提取文件名作为编号
      const fileName = path.split("/").pop() || "";
      const code = fileName.replace(/\.[^.]+$/, ""); // 去掉扩展名
      return { url: path, code, id: "" };
    });

    const resultDir = await CombineImagesWithTemplates(
      selectedTemplates,
      materials,
    );
    message.value = `合成完成，结果已保存到: ${resultDir}`;
  } catch (error: any) {
    message.value = `合成失败: ${error.message || error}`;
  } finally {
    loading.value = false;
    combining.value = false;
  }
};

onMounted(() => {
  loadTemplates();
});
</script>

<template>
  <div class="combine-image">
    <!-- 消息提示 -->
    <div
      v-if="message"
      class="message"
      :class="{ error: message.includes('失败') }"
    >
      {{ message }}
    </div>

    <div class="content-wrapper">
      <!-- 左侧：参数选择 -->
      <div class="left-panel">
        <h2>合成参数</h2>

        <div class="form-group">
          <label>选择模板:</label>
          <div class="template-select-row">
            <el-select
              v-model="selectedTemplateIds"
              multiple
              collapse-tags
              collapse-tags-tooltip
              placeholder="请选择模板"
              :disabled="loading || templates.length === 0"
              style="flex: 1"
            >
              <el-option
                v-for="t in templates"
                :key="t.id"
                :label="`${t.name}（${t.width}×${t.height} px）`"
                :value="t.id!"
              />
            </el-select>
            <el-button :disabled="loading" @click="loadTemplates">
              刷新
            </el-button>
          </div>
          <div v-if="templates.length === 0" class="hint">
            暂无模板，请先在"添加模板"中创建，然后点击"刷新"。
          </div>
          <div v-else class="hint">
            可多选多个模板，同一批图片会分别套用所选模板。
          </div>
          <div v-if="selectedTemplateIds.length" class="preview-template-tags">
            <span class="preview-label">预览模板:</span>
            <el-tag
              v-for="id in selectedTemplateIds"
              :key="id"
              :type="id === previewTemplateId ? 'primary' : 'info'"
              class="preview-tag"
              @click="previewTemplateId = id"
            >
              {{ templates.find((t) => t.id === id)?.name || "模板" }}
            </el-tag>
          </div>
        </div>

        <div class="form-group">
          <label>选择要合成的图片:</label>
          <el-button
            type="primary"
            @click="selectImages"
            :disabled="loading || !selectedTemplate"
            :loading="loading"
          >
            选择图片（可多选）
          </el-button>
          <div v-if="!selectedTemplate" class="hint">
            请先选择一个模板，再选择图片。
          </div>
          <div v-if="selectedImagePaths.length" class="selected-path">
            已选择 {{ selectedImagePaths.length }} 张图片，示例:
            {{ selectedImagePath }}
          </div>
        </div>

        <div v-if="selectedTemplate" class="info-block">
          <p>
            模板框尺寸: {{ selectedTemplate.width }} ×
            {{ selectedTemplate.height }} px
          </p>
          <p v-if="imageBase64 && imageNaturalWidth && imageNaturalHeight">
            原图尺寸: {{ imageNaturalWidth }} × {{ imageNaturalHeight }} px
          </p>
          <p class="hint">
            模板图片作为底图，选择的图片会等比缩放并裁剪，填满中间的模板框区域。
          </p>
        </div>

        <el-button
          v-if="selectedTemplate && selectedImagePaths.length"
          type="primary"
          @click="startCombine"
          :loading="combining"
        >
          开始合成
        </el-button>

        <div v-if="combining" class="progress-wrapper">
          <div class="progress-bar">
            <div class="progress-inner"></div>
          </div>
          <div class="progress-text">
            正在合成 {{ selectedImagePaths.length }} 张图片，请稍候...
          </div>
        </div>
      </div>

      <!-- 右侧：合成预览 -->
      <div class="right-panel">
        <h2>合成预览</h2>
        <div class="preview-wrapper">
          <div
            v-if="!selectedTemplate || !imageBase64"
            class="preview-placeholder"
          >
            <p v-if="!selectedTemplate">请先选择一个模板</p>
            <p v-else>请选择要合成的图片</p>
          </div>
          <div v-else class="canvas-container">
            <canvas ref="compositeCanvas"></canvas>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.combine-image {
  width: 100%;
  padding: 16px;
}

h2 {
  margin: 0 0 12px;
  padding: 14px 16px;
  border: 1px solid var(--apple-border-soft);
  border-radius: 14px;
  background: var(--apple-surface);
  color: var(--apple-text);
  font-size: 21px;
  font-weight: 650;
}

.message {
  padding: 12px 20px;
  margin-bottom: 12px;
  border-radius: 10px;
  background-color: #ecf8f0;
  color: var(--apple-green);
  border: 1px solid #c7ebd2;
}

.message.error {
  background-color: #fff2f4;
  color: var(--apple-red);
  border-color: #ffd0d7;
}

.content-wrapper {
  display: grid;
  grid-template-columns: 1fr 1.2fr;
  gap: 12px;
}

.left-panel,
.right-panel {
  background: var(--apple-surface);
  padding: 16px;
  border: 1px solid var(--apple-border-soft);
  border-radius: 14px;
  box-shadow: var(--apple-shadow-soft);
}

.form-group {
  margin-bottom: 20px;
}

.form-group label {
  display: block;
  margin-bottom: 8px;
  font-weight: 600;
  color: var(--apple-text-secondary);
}

.template-select-row {
  display: flex;
  gap: 8px;
}

.preview-template-tags {
  margin-top: 8px;
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  align-items: center;
}

.preview-label {
  font-size: 12px;
  color: var(--apple-text-muted);
}

.preview-tag {
  cursor: pointer;
}

.selected-path {
  margin-top: 8px;
  padding: 8px 12px;
  background-color: #fbfbfd;
  border: 1px solid var(--apple-border-soft);
  border-radius: 8px;
  font-size: 13px;
  color: var(--apple-text-secondary);
  word-break: break-all;
}

.hint {
  margin-top: 6px;
  font-size: 12px;
  color: var(--apple-text-muted);
}

.info-block {
  margin-top: 10px;
  padding: 10px 12px;
  background-color: #fbfbfd;
  border: 1px solid var(--apple-border-soft);
  border-radius: 8px;
  font-size: 13px;
  color: var(--apple-text-secondary);
}

.preview-wrapper {
  width: 100%;
  min-height: 320px;
}

.preview-placeholder {
  width: 100%;
  height: 320px;
  display: flex;
  align-items: center;
  justify-content: center;
  background-color: #fbfbfd;
  border: 1px dashed var(--apple-border);
  border-radius: 10px;
  color: var(--apple-text-muted);
  font-size: 16px;
  text-align: center;
}

.canvas-container {
  width: 100%;
  max-height: 480px;
  display: flex;
  align-items: center;
  justify-content: center;
  background-color: transparent;
  border-radius: 10px;
  padding: 10px;
}

.canvas-container canvas {
  max-width: 100%;
  max-height: 460px;
  display: block;
  border: 1px solid var(--apple-border-soft);
  background-color: transparent;
}

.progress-wrapper {
  margin-top: 16px;
}

.progress-bar {
  width: 100%;
  height: 8px;
  border-radius: 999px;
  background-color: #f2f2f4;
  overflow: hidden;
}

.progress-inner {
  width: 40%;
  height: 100%;
  background: var(--apple-blue);
  animation: progress-indeterminate 1.2s linear infinite;
}

.progress-text {
  margin-top: 6px;
  font-size: 12px;
  color: var(--apple-text-secondary);
}

@keyframes progress-indeterminate {
  0% {
    transform: translateX(-100%);
  }
  50% {
    transform: translateX(0%);
  }
  100% {
    transform: translateX(100%);
  }
}
</style>
