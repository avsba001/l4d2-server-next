<template>
  <div>
    <div class="flex flex-col">
      <div class="flex justify-between items-center mb-6">
        <h1 class="text-2xl font-bold text-gray-900 dark:text-gray-100">服务器配置</h1>
        <a-button type="primary" size="large" @click="save" :loading="saving" v-if="isAdmin"
          >保存修改</a-button
        >
      </div>

      <div v-if="loading" class="flex justify-center items-center h-64">
        <a-spin size="large" />
      </div>

      <div v-else class="flex flex-col gap-6">
        <a-alert
          v-if="!isAdmin"
          message="只读模式"
          description="您当前使用的是授权码登录，仅具有查看权限。如需修改配置，请使用管理员密码登录。"
          type="info"
          show-icon
        />

        <!-- Settings Card -->
        <div class="flex flex-col bg-white dark:bg-gray-800 p-4 rounded-lg shadow shrink-0">
          <h2 class="text-lg font-semibold mb-4 text-gray-800 dark:text-gray-200 shrink-0">
            基本设置
            <span class="text-xs font-normal text-gray-500 ml-2 block sm:inline">server.cfg</span>
          </h2>

          <div class="flex flex-col gap-4">
            <!-- Hidden Switch -->
            <div class="flex items-center justify-between">
              <span class="text-gray-700 dark:text-gray-300">隐藏服务器 (sv_tags hidden)</span>
              <a-switch v-model:checked="form.hidden" :disabled="!isAdmin" />
            </div>

            <!-- Lobby Only Switch -->
            <div class="flex items-center justify-between">
              <span class="text-gray-700 dark:text-gray-300"
                >开启匹配 (sv_allow_lobby_connect_only)</span
              >
              <a-switch v-model:checked="form.lobby_connect_only" :disabled="!isAdmin" />
            </div>

            <!-- Steam Group Input -->
            <div class="flex flex-col gap-2">
              <span class="text-gray-700 dark:text-gray-300">Steam 组 ID (sv_steamgroup)</span>
              <a-input
                v-model:value="form.steam_group"
                placeholder="输入Steam组ID，留空则删除该设置"
                :disabled="!isAdmin"
              />
            </div>
          </div>
        </div>

        <!-- Custom Config Card -->
        <div class="flex flex-col bg-white dark:bg-gray-800 p-4 rounded-lg shadow shrink-0">
          <h2 class="text-lg font-semibold mb-4 text-gray-800 dark:text-gray-200 shrink-0">
            自定义参数
            <span class="text-xs font-normal text-gray-500 ml-2 block sm:inline"
              >配置文件底部统一区域</span
            >
          </h2>

          <div class="flex flex-col gap-3">
            <div
              v-for="(_, index) in form.custom_config"
              :key="index"
              class="flex items-center gap-2"
            >
              <a-input
                v-model:value="form.custom_config[index]"
                placeholder="例如: exec banned_user.cfg"
                class="font-mono text-sm"
                :disabled="!isAdmin"
              />
              <a-button danger @click="removeCustomConfig(index)" :disabled="!isAdmin">
                删除
              </a-button>
            </div>

            <div
              v-if="form.custom_config.length === 0"
              class="text-gray-500 text-center py-4 bg-gray-50 dark:bg-gray-900 rounded border border-dashed border-gray-300 dark:border-gray-700"
            >
              暂无自定义参数
            </div>

            <a-button type="dashed" block @click="addCustomConfig" :disabled="!isAdmin">
              + 添加一行
            </a-button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
  import { ref, onMounted, reactive, computed } from 'vue';
  import { api } from '../services/api';
  import { message } from 'ant-design-vue';
  import { useAuthStore } from '../stores/auth';

  const authStore = useAuthStore();
  const isAdmin = computed(() => authStore.isAdmin);

  const loading = ref(true);
  const saving = ref(false);

  const form = reactive({
    hidden: false,
    lobby_connect_only: false,
    steam_group: '',
    custom_config: [] as string[],
  });

  const fetchData = async () => {
    try {
      loading.value = true;
      const data = await api.getServerConfig();
      form.hidden = data.hidden;
      form.lobby_connect_only = data.lobby_connect_only;
      form.steam_group = data.steam_group || '';
      form.custom_config = data.custom_config || [];
    } catch (e: any) {
      message.error('获取服务器配置失败: ' + e.message);
    } finally {
      loading.value = false;
    }
  };

  const addCustomConfig = () => {
    form.custom_config.push('');
  };

  const removeCustomConfig = (index: number) => {
    form.custom_config.splice(index, 1);
  };

  const save = async () => {
    try {
      saving.value = true;
      // Filter out empty lines
      const cleanConfig = {
        ...form,
        custom_config: form.custom_config.filter((line) => line.trim() !== ''),
      };
      await api.updateServerConfig(cleanConfig);
      message.success('保存成功，请重启服务器以应用更改');
      await fetchData(); // Refresh data
    } catch (e: any) {
      message.error('保存失败: ' + e.message);
    } finally {
      saving.value = false;
    }
  };

  onMounted(() => {
    fetchData();
  });
</script>
