<template>
  <div class="user-management">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>用户管理</span>
          <el-button type="primary" @click="showAddDialog = true">
            <el-icon><Plus /></el-icon>
            添加用户
          </el-button>
        </div>
      </template>

      <el-table :data="userStore.users" v-loading="userStore.loading" stripe>
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="username" label="用户名" width="150" />
        <el-table-column prop="email" label="邮箱" width="200" />
        <el-table-column prop="role" label="角色" width="100">
          <template #default="{ row }">
            <el-tag :type="row.role === 'admin' ? 'danger' : 'primary'">
              {{ row.role === 'admin' ? '管理员' : '普通用户' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 'active' ? 'success' : 'danger'">
              {{ row.status === 'active' ? '正常' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="180">
          <template #default="{ row }">
            {{ formatDate(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" min-width="200">
          <template #default="{ row }">
            <el-button size="small" @click="showChangePasswordDialog(row)">
              <el-icon><Key /></el-icon>
              修改密码
            </el-button>
            <el-button 
              size="small" 
              type="danger" 
              @click="deleteUser(row)"
              :disabled="row.id === authStore.user?.id"
            >
              <el-icon><Delete /></el-icon>
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 添加用户对话框 -->
    <el-dialog v-model="showAddDialog" title="添加用户" width="500px">
      <el-form :model="userForm" :rules="userRules" ref="userFormRef" label-width="80px">
        <el-form-item label="用户名" prop="username">
          <el-input v-model="userForm.username" placeholder="请输入用户名" />
        </el-form-item>
        <el-form-item label="邮箱" prop="email">
          <el-input v-model="userForm.email" placeholder="请输入邮箱" />
        </el-form-item>
        <el-form-item label="密码" prop="password">
          <el-input v-model="userForm.password" type="password" placeholder="请输入密码" show-password />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showAddDialog = false">取消</el-button>
        <el-button type="primary" @click="submitUser" :loading="submitting">确定</el-button>
      </template>
    </el-dialog>

    <!-- 修改密码对话框 -->
    <el-dialog v-model="showPasswordDialog" title="修改密码" width="400px">
      <el-form :model="passwordForm" :rules="passwordRules" ref="passwordFormRef" label-width="80px">
        <el-form-item label="用户名">
          <el-input :value="currentUser?.username" disabled />
        </el-form-item>
        <el-form-item label="新密码" prop="password">
          <el-input v-model="passwordForm.password" type="password" placeholder="请输入新密码" show-password />
        </el-form-item>
        <el-form-item label="确认密码" prop="confirmPassword">
          <el-input v-model="passwordForm.confirmPassword" type="password" placeholder="请确认新密码" show-password />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showPasswordDialog = false">取消</el-button>
        <el-button type="primary" @click="submitPassword" :loading="submitting">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Key, Delete } from '@element-plus/icons-vue'
import { useUserStore } from '@/stores/user'
import { useAuthStore } from '@/stores/auth'
import type { User, CreateUserRequest } from '@/types/auth'

const userStore = useUserStore()
const authStore = useAuthStore()

const showAddDialog = ref(false)
const showPasswordDialog = ref(false)
const submitting = ref(false)
const userFormRef = ref()
const passwordFormRef = ref()
const currentUser = ref<User | null>(null)

const userForm = ref<CreateUserRequest>({
  username: '',
  email: '',
  password: ''
})

const passwordForm = ref({
  password: '',
  confirmPassword: ''
})

const userRules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  email: [
    { required: true, message: '请输入邮箱', trigger: 'blur' },
    { type: 'email', message: '请输入正确的邮箱格式', trigger: 'blur' }
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, message: '密码长度不能少于6位', trigger: 'blur' }
  ]
}

const passwordRules = {
  password: [
    { required: true, message: '请输入新密码', trigger: 'blur' },
    { min: 6, message: '密码长度不能少于6位', trigger: 'blur' }
  ],
  confirmPassword: [
    { required: true, message: '请确认新密码', trigger: 'blur' },
    {
      validator: (_rule: any, value: string, callback: Function) => {
        if (value !== passwordForm.value.password) {
          callback(new Error('两次输入的密码不一致'))
        } else {
          callback()
        }
      },
      trigger: 'blur'
    }
  ]
}

// 格式化日期
const formatDate = (dateString: string) => {
  return new Date(dateString).toLocaleString('zh-CN')
}

// 显示修改密码对话框
const showChangePasswordDialog = (user: User) => {
  currentUser.value = user
  passwordForm.value = {
    password: '',
    confirmPassword: ''
  }
  showPasswordDialog.value = true
}

// 删除用户
const deleteUser = async (user: User) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除用户 "${user.username}" 吗？`,
      '确认删除',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    await userStore.deleteUser(user.id)
    ElMessage.success('删除成功')
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
    }
  }
}

// 提交用户信息
const submitUser = async () => {
  if (!userFormRef.value) return
  
  try {
    await userFormRef.value.validate()
    submitting.value = true
    
    await userStore.createUser(userForm.value)
    
    ElMessage.success('添加成功')
    showAddDialog.value = false
    resetUserForm()
  } catch (error) {
    ElMessage.error('添加失败')
  } finally {
    submitting.value = false
  }
}

// 提交密码修改
const submitPassword = async () => {
  if (!passwordFormRef.value || !currentUser.value) return
  
  try {
    await passwordFormRef.value.validate()
    submitting.value = true
    
    await userStore.changePassword({
      user_id: currentUser.value.id.toString(),
      new_password: passwordForm.value.password
    })
    
    ElMessage.success('密码修改成功')
    showPasswordDialog.value = false
  } catch (error) {
    ElMessage.error('密码修改失败')
  } finally {
    submitting.value = false
  }
}

// 重置用户表单
const resetUserForm = () => {
  userForm.value = {
    username: '',
    email: '',
    password: ''
  }
  userFormRef.value?.resetFields()
}

onMounted(() => {
  userStore.fetchUsers()
})
</script>

<style scoped>
.user-management {
  height: 100%;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style> 