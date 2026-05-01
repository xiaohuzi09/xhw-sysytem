<script setup lang="ts">
import { ref, computed, onMounted } from "vue";
import { ElMessage, ElMessageBox } from "element-plus";
import {
  apiGetUsers,
  apiCreateUser,
  apiUpdateUser,
  apiDeleteUser,
} from "../api/user";
import type { User } from "../api/user";

const allUsers = ref<User[]>([]);
const loading = ref(false);
const currentPage = ref(1);
const pageSize = ref(20);

const total = computed(() => allUsers.value.length);
const pagedUsers = computed(() => {
  const start = (currentPage.value - 1) * pageSize.value;
  return allUsers.value.slice(start, start + pageSize.value);
});

const dialogVisible = ref(false);
const dialogMode = ref<"create" | "edit">("create");
const editingId = ref<number | null>(null);
const formUsername = ref("");
const formPassword = ref("");
const formEmail = ref("");
const formNickname = ref("");
const formStatus = ref<1 | 0>(1);
const submitLoading = ref(false);

function resetForm() {
  formUsername.value = "";
  formPassword.value = "";
  formEmail.value = "";
  formNickname.value = "";
  formStatus.value = 1;
  editingId.value = null;
}

function rowIsAdmin(row: User): boolean {
  const r = row.role;
  return r === "admin" || r === "ADMIN";
}

function statusLabel(status?: number): string {
  return status === 0 ? "禁用" : "正常";
}

const loadUsers = async () => {
  try {
    loading.value = true;
    const result = await apiGetUsers();
    const raw = result.data;
    allUsers.value = Array.isArray(raw) ? raw : [];
    const maxPage = Math.max(
      1,
      Math.ceil(allUsers.value.length / pageSize.value) || 1,
    );
    if (currentPage.value > maxPage) {
      currentPage.value = maxPage;
    }
  } catch (error: unknown) {
    const msg = error instanceof Error ? error.message : String(error);
    ElMessage.error(`加载用户失败: ${msg}`);
  } finally {
    loading.value = false;
  }
};

const handlePageChange = (page: number) => {
  currentPage.value = page;
};

const handleSizeChange = (size: number) => {
  pageSize.value = size;
  currentPage.value = 1;
};

const openCreate = () => {
  dialogMode.value = "create";
  resetForm();
  dialogVisible.value = true;
};

const openEdit = (row: User) => {
  dialogMode.value = "edit";
  editingId.value = row.id;
  formUsername.value = row.username;
  formPassword.value = "";
  formEmail.value = row.email || "";
  formNickname.value = row.nickname || "";
  formStatus.value = row.status === 0 ? 0 : 1;
  dialogVisible.value = true;
};

const handleSubmit = async () => {
  const username = formUsername.value.trim();
  const email = formEmail.value.trim();
  if (!username) {
    ElMessage.warning("请输入用户名");
    return;
  }
  if (dialogMode.value === "create") {
    if (!email) {
      ElMessage.warning("请输入邮箱");
      return;
    }
    if (!formPassword.value || formPassword.value.length < 6) {
      ElMessage.warning("密码至少 6 位");
      return;
    }
  }

  try {
    submitLoading.value = true;
    if (dialogMode.value === "create") {
      await apiCreateUser({
        username,
        password: formPassword.value,
        email,
        nickname: formNickname.value.trim() || undefined,
        status: formStatus.value,
      });
      ElMessage.success("用户已创建（角色由服务端按规则分配）");
    } else if (editingId.value != null) {
      const payload: {
        username: string;
        email?: string;
        nickname?: string;
        status?: number;
        password?: string;
      } = {
        username,
        email: email || undefined,
        nickname: formNickname.value.trim() || undefined,
        status: formStatus.value,
      };
      if (formPassword.value) {
        if (formPassword.value.length < 6) {
          ElMessage.warning("新密码至少 6 位");
          submitLoading.value = false;
          return;
        }
        payload.password = formPassword.value;
      }
      await apiUpdateUser(editingId.value, payload);
      ElMessage.success("已保存");
    }
    dialogVisible.value = false;
    resetForm();
    await loadUsers();
  } catch (error: unknown) {
    const msg = error instanceof Error ? error.message : String(error);
    ElMessage.error(`操作失败: ${msg}`);
  } finally {
    submitLoading.value = false;
  }
};

const handleDelete = async (row: User) => {
  try {
    await ElMessageBox.confirm(
      `确定删除用户「${row.username}」？此操作不可恢复。`,
      "删除用户",
      {
        confirmButtonText: "删除",
        cancelButtonText: "取消",
        type: "warning",
      },
    );
    await apiDeleteUser(row.id);
    ElMessage.success("已删除");
    await loadUsers();
  } catch (error: unknown) {
    if (error === "cancel") {
      return;
    }
    const msg = error instanceof Error ? error.message : String(error);
    ElMessage.error(`删除失败: ${msg}`);
  }
};

onMounted(() => {
  loadUsers();
});
</script>

<template>
  <div class="user-manage">
    <header class="page-head">
      <div>
        <h2>用户管理</h2>
        <p class="sub">
          与 xhw-service 一致：列表一次返回全部用户；新建用户必填邮箱；角色由服务端判定。
        </p>
      </div>
      <el-button type="primary" @click="openCreate">新建用户</el-button>
    </header>

    <el-table
      v-loading="loading"
      :data="pagedUsers"
      stripe
      class="user-table"
      empty-text="暂无用户数据"
    >
      <el-table-column prop="id" label="ID" width="72" />
      <el-table-column prop="username" label="用户名" min-width="120" />
      <el-table-column prop="email" label="邮箱" min-width="160" show-overflow-tooltip />
      <el-table-column prop="nickname" label="昵称" min-width="100" show-overflow-tooltip />
      <el-table-column label="角色" width="100">
        <template #default="{ row }">
          <el-tag :type="rowIsAdmin(row) ? 'danger' : 'info'" size="small">
            {{ rowIsAdmin(row) ? "管理员" : row.role || "user" }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="状态" width="88">
        <template #default="{ row }">
          <el-tag :type="row.status === 0 ? 'warning' : 'success'" size="small">
            {{ statusLabel(row.status) }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="created_at" label="创建时间" min-width="168" />
      <el-table-column label="操作" width="160" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" @click="openEdit(row)">
            编辑
          </el-button>
          <el-button link type="danger" @click="handleDelete(row)">
            删除
          </el-button>
        </template>
      </el-table-column>
    </el-table>

    <div class="pager">
      <el-pagination
        :current-page="currentPage"
        :page-size="pageSize"
        :total="total"
        :page-sizes="[10, 20, 50, 100]"
        layout="total, sizes, prev, pager, next"
        background
        @current-change="handlePageChange"
        @size-change="handleSizeChange"
      />
    </div>

    <el-dialog
      v-model="dialogVisible"
      :title="dialogMode === 'create' ? '新建用户' : '编辑用户'"
      width="520px"
      destroy-on-close
      @closed="resetForm"
    >
      <el-form label-position="top">
        <el-form-item label="用户名" required>
          <el-input
            v-model="formUsername"
            placeholder="登录名"
            autocomplete="off"
          />
        </el-form-item>
        <el-form-item label="邮箱" :required="dialogMode === 'create'">
          <el-input
            v-model="formEmail"
            placeholder="name@example.com"
            autocomplete="off"
          />
        </el-form-item>
        <el-form-item label="昵称">
          <el-input v-model="formNickname" placeholder="选填" />
        </el-form-item>
        <el-form-item
          :label="dialogMode === 'create' ? '密码' : '新密码（留空则不修改）'"
          :required="dialogMode === 'create'"
        >
          <el-input
            v-model="formPassword"
            type="password"
            show-password
            placeholder="至少 6 位"
            autocomplete="new-password"
          />
        </el-form-item>
        <el-form-item label="状态">
          <el-radio-group v-model="formStatus">
            <el-radio :label="1">正常</el-radio>
            <el-radio :label="0">禁用</el-radio>
          </el-radio-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button
          type="primary"
          :loading="submitLoading"
          @click="handleSubmit"
        >
          确定
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.user-manage {
  width: 100%;
  padding: 16px;
}

.page-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  padding: 14px 16px;
  margin-bottom: 12px;
  border: 1px solid var(--apple-border-soft);
  border-radius: 14px;
  background: var(--apple-surface);
  box-shadow: var(--apple-shadow-soft);
}

.page-head h2 {
  margin: 0 0 6px;
  font-size: 21px;
  font-weight: 650;
  color: var(--apple-text);
}

.sub {
  margin: 0;
  font-size: 13px;
  color: var(--apple-text-muted);
}

.user-table {
  width: 100%;
  border: 1px solid var(--apple-border-soft);
  border-radius: 14px;
  overflow: hidden;
  box-shadow: var(--apple-shadow-soft);
}

.pager {
  margin-top: 16px;
  display: flex;
  justify-content: flex-end;
}

@media (max-width: 760px) {
  .user-manage {
    padding: 12px;
  }

  .page-head {
    flex-direction: column;
  }
}
</style>
