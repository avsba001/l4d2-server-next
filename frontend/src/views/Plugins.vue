<script setup lang="ts">
  import { ref, computed, onMounted, onErrorCaptured, reactive, watch } from 'vue';
  import {
    message,
    Card as ACard,
    Tabs as ATabs,
    TabPane as ATabPane,
    Input as AInput,
    Button as AButton,
    Table as ATable,
    Popconfirm as APopconfirm,
    Upload as AUpload,
    RadioGroup as ARadioGroup,
    Alert as AAlert,
  } from 'ant-design-vue';
  import {
    UploadOutlined,
    DeleteOutlined,
    PoweroffOutlined,
    SearchOutlined,
    ReloadOutlined,
    SettingOutlined,
    SyncOutlined,
  } from '@ant-design/icons-vue';
  import { api } from '../services/api';
  import type { UploadProps, TablePaginationConfig } from 'ant-design-vue';
  import PluginConfigModal from '../components/PluginConfigModal.vue';
  import { useAuthStore } from '../stores/auth';

  const authStore = useAuthStore();

  interface Plugin {
    name: string;
    status: 'enabled' | 'disabled';
    description?: string;
  }

  const plugins = ref<Plugin[]>([]);
  const loading = ref(false);
  const uploading = ref(false);
  const fileList = ref<UploadProps['fileList']>([]);
  const activeTab = ref('enabled');
  const selectedRowKeys = ref<string[]>([]);
  const searchText = ref('');
  const filterText = ref('');

  const configModalVisible = ref(false);
  const currentConfigPlugin = ref('');
  const pendingFiles = ref<File[]>([]);
  let uploadTimer: any = null;

  const presetModalVisible = ref(false);
  const presets = ref<any[]>([]);
  const selectedPreset = ref('');
  const applyingPreset = ref(false);
  const footerContainerRef = ref<HTMLElement | null>(null);

  const getPopupContainer = (trigger: HTMLElement) => {
    return footerContainerRef.value || trigger || document.body;
  };

  const openPresetModal = async () => {
    try {
      const data = await api.getPresets();
      presets.value = data || [];
      selectedPreset.value = '';
      presetModalVisible.value = true;
    } catch (error: any) {
      message.error('获取预设列表失败: ' + error.message);
    }
  };

  const confirmApplyPreset = async () => {
    if (!selectedPreset.value) {
      message.warning('请选择一个预设');
      return;
    }

    applyingPreset.value = true;
    try {
      await api.applyPreset(selectedPreset.value);
      message.success('预设应用成功');
      presetModalVisible.value = false;
      fetchPlugins();
    } catch (error: any) {
      message.error('应用预设失败: ' + error.message);
    } finally {
      applyingPreset.value = false;
    }
  };

  onErrorCaptured((err) => {
    console.error('Plugins.vue Error:', err);
    message.error('插件管理页面发生错误');
    return false;
  });

  const enabledPlugins = computed(() =>
    plugins.value.filter((p) => p.status === 'enabled').sort((a, b) => a.name.localeCompare(b.name))
  );
  const disabledPlugins = computed(() =>
    plugins.value
      .filter((p) => p.status === 'disabled')
      .sort((a, b) => a.name.localeCompare(b.name))
  );

  const filteredDisabledPlugins = computed(() => {
    if (!filterText.value) return disabledPlugins.value;
    const lower = filterText.value.toLowerCase();
    return disabledPlugins.value.filter(
      (p) =>
        p.name.toLowerCase().includes(lower) ||
        (p.description && p.description.toLowerCase().includes(lower))
    );
  });

  const filteredEnabledPlugins = computed(() => {
    if (!filterText.value) return enabledPlugins.value;
    const lower = filterText.value.toLowerCase();
    return enabledPlugins.value.filter(
      (p) =>
        p.name.toLowerCase().includes(lower) ||
        (p.description && p.description.toLowerCase().includes(lower))
    );
  });

  const fetchPlugins = async () => {
    loading.value = true;
    try {
      plugins.value = await api.getPlugins();
    } catch (error: any) {
      message.error('加载插件失败: ' + error.message);
    } finally {
      loading.value = false;
    }
  };

  const processPendingUploads = async () => {
    const filesToUpload = [...pendingFiles.value];
    pendingFiles.value = [];

    if (filesToUpload.length === 0) return;

    const hide = message.loading(`正在上传 ${filesToUpload.length} 个插件...`, 0);
    uploading.value = true;
    try {
      await api.uploadPlugin(filesToUpload);
      message.success('插件上传成功');
      fileList.value = [];
      fetchPlugins();
    } catch (error: any) {
      message.error('上传失败: ' + error.message);
    } finally {
      uploading.value = false;
      hide();
    }
  };

  const handleUpload = (file: File) => {
    if (!file.name.endsWith('.zip')) {
      message.error('只允许上传 .zip 格式的文件');
      return false;
    }

    pendingFiles.value.push(file);

    if (uploadTimer) clearTimeout(uploadTimer);
    uploadTimer = setTimeout(() => {
      processPendingUploads();
    }, 100);

    return false; // Prevent default upload behavior
  };

  const togglePlugin = async (plugin: Plugin) => {
    const actionText = plugin.status === 'enabled' ? '禁用' : '启用';
    const hide = message.loading(`正在${actionText}插件...`, 0);

    try {
      if (plugin.status === 'enabled') {
        await api.disablePlugin(plugin.name);
      } else {
        await api.enablePlugin(plugin.name);
      }
      message.success(`插件${actionText}成功`);
      fetchPlugins();
    } catch (error: any) {
      message.error(`${actionText}插件失败: ` + error.message);
    } finally {
      hide();
    }
  };

  const deletePlugin = async (plugin: Plugin) => {
    if (plugin.status === 'enabled') {
      message.warning('请先禁用插件');
      return;
    }

    const hide = message.loading('正在删除插件...', 0);
    try {
      await api.deletePlugin(plugin.name);
      message.success('插件删除成功');
      fetchPlugins();
    } catch (error: any) {
      message.error('删除插件失败: ' + error.message);
    } finally {
      hide();
    }
  };

  const handleSearch = () => {
    filterText.value = searchText.value;
  };

  const handleReset = () => {
    searchText.value = '';
    filterText.value = '';
  };

  watch(activeTab, () => {
    selectedRowKeys.value = [];
  });

  const handleBatchEnable = async () => {
    if (selectedRowKeys.value.length === 0) return;

    const hide = message.loading('正在批量启用插件...', 0);
    try {
      await api.enablePlugins(selectedRowKeys.value);
      message.success(`成功启用 ${selectedRowKeys.value.length} 个插件`);
      selectedRowKeys.value = [];
      fetchPlugins();
    } catch (error: any) {
      message.error('批量启用失败: ' + error.message);
      fetchPlugins();
    } finally {
      hide();
    }
  };

  const handleBatchDisable = async () => {
    if (selectedRowKeys.value.length === 0) return;

    const hide = message.loading('正在批量禁用插件...', 0);
    try {
      await api.disablePlugins(selectedRowKeys.value);
      message.success(`成功禁用 ${selectedRowKeys.value.length} 个插件`);
      selectedRowKeys.value = [];
      fetchPlugins();
    } catch (error: any) {
      message.error('批量禁用失败: ' + error.message);
      fetchPlugins();
    } finally {
      hide();
    }
  };

  const rowSelection = computed(() => {
    return {
      selectedRowKeys: selectedRowKeys.value,
      onChange: onSelectChange,
    };
  });

  const handleBatchDelete = async () => {
    if (selectedRowKeys.value.length === 0) return;

    const hide = message.loading('正在批量删除插件...', 0);
    try {
      // Execute deletions sequentially to avoid potential conflicts or backend overload
      // Or concurrent if backend supports it. Here sequential for safety.
      for (const name of selectedRowKeys.value) {
        await api.deletePlugin(name);
      }
      message.success(`成功删除 ${selectedRowKeys.value.length} 个插件`);
      selectedRowKeys.value = [];
      fetchPlugins();
    } catch (error: any) {
      message.error('批量删除部分失败: ' + error.message);
      // Refresh to see what was actually deleted
      fetchPlugins();
    } finally {
      hide();
    }
  };

  const onSelectChange = (keys: any[]) => {
    selectedRowKeys.value = keys;
  };

  const openConfig = async (plugin: Plugin) => {
    currentConfigPlugin.value = plugin.name;
    configModalVisible.value = true;
  };

  onMounted(() => {
    fetchPlugins();
  });

  const enabledColumns = [
    {
      title: '插件名称',
      dataIndex: 'name',
      key: 'name',
      sorter: (a: Plugin, b: Plugin) => a.name.localeCompare(b.name),
    },
    {
      title: '操作',
      key: 'actions',
      width: 200,
    },
  ];

  const disabledColumns = [
    {
      title: '插件名称',
      dataIndex: 'name',
      key: 'name',
      sorter: (a: Plugin, b: Plugin) => a.name.localeCompare(b.name),
    },
    {
      title: '操作',
      key: 'actions',
      width: 200,
    },
  ];

  const enabledPagination = reactive<TablePaginationConfig>({
    current: 1,
    pageSize: 10,
    showSizeChanger: true,
    pageSizeOptions: ['10', '20', '50', '100'],
    showTotal: (total: number) => `共 ${total} 条`,
  });

  const disabledPagination = reactive<TablePaginationConfig>({
    current: 1,
    pageSize: 10,
    showSizeChanger: true,
    pageSizeOptions: ['10', '20', '50', '100'],
    showTotal: (total: number) => `共 ${total} 条`,
  });

  const handleEnabledTableChange = (pag: TablePaginationConfig) => {
    enabledPagination.current = pag.current;
    enabledPagination.pageSize = pag.pageSize;
  };

  const handleDisabledTableChange = (pag: TablePaginationConfig) => {
    disabledPagination.current = pag.current;
    disabledPagination.pageSize = pag.pageSize;
  };
</script>

<template>
  <div class="space-y-6">
    <div class="flex flex-col md:flex-row justify-between items-start md:items-center gap-4">
      <div>
        <h1 class="text-2xl font-bold text-gray-800 dark:text-gray-100">插件管理</h1>
        <p class="text-gray-500 dark:text-gray-400 mt-1">管理服务器插件和模组</p>
      </div>
      <div class="flex gap-2">
        <a-button
          v-if="authStore.isAdmin"
          type="primary"
          ghost
          @click="openPresetModal"
          class="!flex !items-center !justify-center"
        >
          <template #icon><SettingOutlined /></template>
          应用预设
        </a-button>
        <a-button
          type="default"
          @click="fetchPlugins"
          :loading="loading"
          class="!flex !items-center !justify-center"
        >
          <template #icon><SyncOutlined /></template>
          刷新列表
        </a-button>
      </div>
    </div>

    <a-card :bordered="false" class="shadow-sm dark:bg-gray-800">
      <a-tabs v-model:activeKey="activeTab">
        <!-- Enabled Plugins Tab -->
        <a-tab-pane key="enabled" tab="已启用插件">
          <div
            class="mb-4 flex flex-col lg:flex-row justify-between items-start lg:items-center gap-4"
          >
            <div class="flex flex-col sm:flex-row gap-2 w-full lg:w-auto">
              <a-input
                v-model:value="searchText"
                placeholder="搜索插件..."
                class="w-full sm:w-[200px]"
                allow-clear
                @pressEnter="handleSearch"
              />
              <div class="flex gap-2 w-full sm:w-auto">
                <a-button
                  type="primary"
                  @click="handleSearch"
                  class="!flex !items-center !justify-center flex-1 sm:flex-none"
                >
                  <template #icon><SearchOutlined /></template>
                  搜索
                </a-button>
                <a-button
                  @click="handleReset"
                  class="!flex !items-center !justify-center flex-1 sm:flex-none"
                >
                  <template #icon><ReloadOutlined /></template>
                  重置
                </a-button>
              </div>
            </div>

            <div v-if="selectedRowKeys.length > 0 && authStore.isAdmin" class="flex gap-2">
              <div>
                <a-popconfirm
                  title="确定要禁用选中的插件吗？"
                  ok-text="确定"
                  cancel-text="取消"
                  @confirm="handleBatchDisable"
                >
                  <a-button danger>批量禁用 ({{ selectedRowKeys.length }})</a-button>
                </a-popconfirm>
              </div>
            </div>
          </div>

          <a-table
            :columns="enabledColumns"
            :data-source="filteredEnabledPlugins"
            :loading="loading"
            :pagination="enabledPagination"
            @change="handleEnabledTableChange"
            row-key="name"
            :scroll="{ x: 'max-content' }"
            :row-selection="rowSelection"
          >
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'name'">
                <div class="font-medium text-gray-700 dark:text-gray-200">{{ record.name }}</div>
                <div v-if="record.description" class="text-xs text-gray-400 dark:text-gray-500">
                  {{ record.description }}
                </div>
              </template>

              <template v-else-if="column.key === 'actions'">
                <div class="flex gap-2">
                  <a-button
                    type="default"
                    size="small"
                    class="!flex !items-center !justify-center"
                    @click="openConfig(record as Plugin)"
                  >
                    <template #icon><SettingOutlined /></template>
                    配置
                  </a-button>
                  <a-popconfirm
                    v-if="authStore.isAdmin"
                    title="确定要禁用这个插件吗？"
                    ok-text="确定"
                    cancel-text="取消"
                    @confirm="togglePlugin(record as Plugin)"
                  >
                    <a-button
                      type="default"
                      danger
                      size="small"
                      class="!flex !items-center !justify-center"
                    >
                      <template #icon><PoweroffOutlined /></template>
                      禁用
                    </a-button>
                  </a-popconfirm>
                </div>
              </template>
            </template>
          </a-table>
        </a-tab-pane>

        <!-- Disabled Plugins Tab -->
        <a-tab-pane key="disabled" tab="未启用插件" v-if="authStore.isAdmin">
          <div
            class="mb-4 flex flex-col lg:flex-row justify-between items-start lg:items-center gap-4"
          >
            <div class="flex flex-col sm:flex-row gap-2 w-full lg:w-auto">
              <a-input
                v-model:value="searchText"
                placeholder="搜索插件..."
                class="w-full sm:w-[200px]"
                allow-clear
                @pressEnter="handleSearch"
              />
              <div class="flex gap-2 w-full sm:w-auto">
                <a-button
                  type="primary"
                  @click="handleSearch"
                  class="!flex !items-center !justify-center flex-1 sm:flex-none"
                >
                  <template #icon><SearchOutlined /></template>
                  搜索
                </a-button>
                <a-button
                  @click="handleReset"
                  class="!flex !items-center !justify-center flex-1 sm:flex-none"
                >
                  <template #icon><ReloadOutlined /></template>
                  重置
                </a-button>
              </div>
            </div>

            <div class="flex flex-col sm:flex-row gap-2 w-full lg:w-auto lg:items-center">
              <div
                v-if="selectedRowKeys.length > 0 && authStore.isAdmin"
                class="flex gap-2 w-full sm:w-auto"
              >
                <div>
                  <a-popconfirm
                    title="确定要启用选中的插件吗？"
                    ok-text="确定"
                    cancel-text="取消"
                    @confirm="handleBatchEnable"
                  >
                    <a-button type="primary">批量启用 ({{ selectedRowKeys.length }})</a-button>
                  </a-popconfirm>
                </div>
                <div>
                  <a-popconfirm
                    title="确定要删除选中的插件吗？"
                    ok-text="确定"
                    cancel-text="取消"
                    @confirm="handleBatchDelete"
                  >
                    <a-button danger>批量删除 ({{ selectedRowKeys.length }})</a-button>
                  </a-popconfirm>
                </div>
              </div>

              <a-upload
                v-if="authStore.isAdmin"
                v-model:file-list="fileList"
                :before-upload="handleUpload"
                accept=".zip"
                :show-upload-list="false"
                :disabled="uploading"
                multiple
                class="flex-1 sm:flex-none"
              >
                <a-button
                  type="primary"
                  :loading="uploading"
                  class="!flex !items-center !justify-center w-full"
                >
                  <template #icon><UploadOutlined /></template>
                  上传插件 (.zip)
                </a-button>
              </a-upload>
            </div>
          </div>

          <a-table
            :columns="disabledColumns"
            :data-source="filteredDisabledPlugins"
            :loading="loading"
            :pagination="disabledPagination"
            @change="handleDisabledTableChange"
            row-key="name"
            :scroll="{ x: 'max-content' }"
            :row-selection="rowSelection"
          >
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'name'">
                <div class="font-medium text-gray-700 dark:text-gray-200">{{ record.name }}</div>
                <div v-if="record.description" class="text-xs text-gray-400 dark:text-gray-500">
                  {{ record.description }}
                </div>
              </template>

              <template v-else-if="column.key === 'actions'">
                <div class="flex items-center gap-2">
                  <a-popconfirm
                    title="确定要启用这个插件吗？"
                    ok-text="确定"
                    cancel-text="取消"
                    @confirm="togglePlugin(record as Plugin)"
                    :disabled="!authStore.isAdmin"
                  >
                    <a-button
                      type="primary"
                      size="small"
                      class="!flex !items-center !justify-center"
                      :disabled="!authStore.isAdmin"
                    >
                      <template #icon><PoweroffOutlined /></template>
                      启用
                    </a-button>
                  </a-popconfirm>

                  <a-popconfirm
                    v-if="authStore.isAdmin"
                    title="确定要删除这个插件吗？"
                    ok-text="确定"
                    cancel-text="取消"
                    @confirm="deletePlugin(record as Plugin)"
                  >
                    <a-button
                      type="text"
                      danger
                      size="small"
                      class="!flex !items-center !justify-center"
                    >
                      <template #icon><DeleteOutlined /></template>
                      删除
                    </a-button>
                  </a-popconfirm>
                </div>
              </template>
            </template>
          </a-table>
        </a-tab-pane>
      </a-tabs>
    </a-card>

    <a-modal
      v-model:open="presetModalVisible"
      title="选择插件预设"
      :confirm-loading="applyingPreset"
      ok-text="应用"
      cancel-text="取消"
      :width="600"
    >
      <template #footer>
        <div class="flex justify-end gap-2" ref="footerContainerRef">
          <a-button @click="presetModalVisible = false">取消</a-button>
          <a-popconfirm
            title="此操作将重置所有插件状态，确定要继续吗？"
            ok-text="确定"
            cancel-text="取消"
            @confirm="confirmApplyPreset"
            :getPopupContainer="getPopupContainer"
          >
            <a-button type="primary" :loading="applyingPreset">应用</a-button>
          </a-popconfirm>
        </div>
      </template>
      <a-alert
        message="注意"
        description="应用预设将重置所有插件状态，禁用当前所有插件并按预设启用。配置项也会被覆盖。"
        type="warning"
        show-icon
        class="mb-6"
      />
      <div v-if="presets.length === 0" class="text-center py-4 text-gray-500">暂无可用预设</div>
      <div class="flex flex-col gap-2 max-h-[60vh] overflow-y-auto mt-6">
        <a-radio-group v-model:value="selectedPreset" button-style="solid" class="w-full">
          <div class="flex flex-col gap-3">
            <a-radio-button
              v-for="preset in presets"
              :key="preset.name"
              :value="preset.name"
              class="!h-auto !py-3 !px-4 !flex !items-center !rounded-md !border !border-gray-200 dark:!border-gray-700 hover:!border-blue-500 transition-all"
            >
              <div class="flex flex-col text-left">
                <span class="font-medium text-base">{{ preset.name }}</span>
                <span
                  class="text-xs mt-1"
                  :class="selectedPreset === preset.name ? 'text-blue-100' : 'text-gray-500'"
                >
                  包含 {{ preset.plugin_count || 0 }} 个插件
                </span>
              </div>
            </a-radio-button>
          </div>
        </a-radio-group>
      </div>
    </a-modal>

    <PluginConfigModal v-model:open="configModalVisible" :plugin-name="currentConfigPlugin" />
  </div>
</template>

<style scoped>
  /* 修复 Popconfirm 按钮在 flex 容器中换行的问题 */
  :deep(.ant-popconfirm-buttons) {
    display: flex;
    justify-content: flex-end; /* 按钮靠右对齐 */
    flex-wrap: nowrap;
    gap: 8px;
    white-space: nowrap;
  }

  :deep(.ant-popconfirm-buttons button) {
    margin-left: 0 !important;
  }

  /* 防止 Popconfirm 文字换行跳动 */
  :deep(.ant-popconfirm-message) {
    white-space: nowrap;
  }

  /* 修复 RadioButton 垂直排列时的边框问题 */
  :deep(.ant-radio-button-wrapper) {
    margin-right: 0 !important;
    border-left-width: 1px !important;
  }

  :deep(.ant-radio-button-wrapper:not(:first-child)::before) {
    display: none !important;
  }
</style>
