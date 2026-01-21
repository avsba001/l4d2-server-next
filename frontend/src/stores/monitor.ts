import { defineStore } from 'pinia';
import { ref, computed } from 'vue';
import { api } from '../services/api';
import { message } from 'ant-design-vue';

export const useMonitorStore = defineStore('monitor', () => {
  const isMonitoring = ref(false);
  const timer = ref<any>(null);
  const historyEnabled = ref(false);
  const viewMode = ref<'realtime' | '1h' | '24h' | '3d'>('realtime');

  // Realtime Data Storage (Max 600)
  const maxPoints = 600;
  const rTimestamps = ref<string[]>([]);
  const rCpuData = ref<number[]>([]);
  const rCpuMaxCoreData = ref<number[]>([]);
  const rMemUsedData = ref<number[]>([]); // GB
  const rSwapUsedData = ref<number[]>([]); // GB
  const rNetUpData = ref<number[]>([]); // KB/s
  const rNetDownData = ref<number[]>([]); // KB/s
  const rDiskUsedData = ref<number[]>([]); // GB

  // History Data Storage
  const hRawTimestamps = ref<number[]>([]);
  const hTimestamps = ref<string[]>([]);
  const hCpuData = ref<number[]>([]);
  const hCpuMaxCoreData = ref<number[]>([]);
  const hMemUsedData = ref<number[]>([]);
  const hSwapUsedData = ref<number[]>([]);
  const hNetUpData = ref<number[]>([]);
  const hNetDownData = ref<number[]>([]);
  const hDiskUsedData = ref<number[]>([]);

  // Computed for View
  const timestamps = computed(() =>
    viewMode.value === 'realtime' ? rTimestamps.value : hTimestamps.value
  );
  const cpuData = computed(() => (viewMode.value === 'realtime' ? rCpuData.value : hCpuData.value));
  const cpuMaxCoreData = computed(() =>
    viewMode.value === 'realtime' ? rCpuMaxCoreData.value : hCpuMaxCoreData.value
  );
  const memUsedData = computed(() =>
    viewMode.value === 'realtime' ? rMemUsedData.value : hMemUsedData.value
  );
  const swapUsedData = computed(() =>
    viewMode.value === 'realtime' ? rSwapUsedData.value : hSwapUsedData.value
  );
  const netUpData = computed(() =>
    viewMode.value === 'realtime' ? rNetUpData.value : hNetUpData.value
  );
  const netDownData = computed(() =>
    viewMode.value === 'realtime' ? rNetDownData.value : hNetDownData.value
  );
  const diskUsedData = computed(() =>
    viewMode.value === 'realtime' ? rDiskUsedData.value : hDiskUsedData.value
  );

  // Disk Data (Current Only)
  const memTotal = ref(0);
  const swapTotal = ref(0);
  const diskTotal = ref(0);
  const diskUsed = ref(0);
  const diskPercent = computed(() => {
    if (diskTotal.value === 0) return 0;
    return (diskUsed.value / diskTotal.value) * 100;
  });

  const clearHistoryData = () => {
    hRawTimestamps.value = [];
    hTimestamps.value = [];
    hCpuData.value = [];
    hCpuMaxCoreData.value = [];
    hMemUsedData.value = [];
    hSwapUsedData.value = [];
    hNetUpData.value = [];
    hNetDownData.value = [];
    hDiskUsedData.value = [];
  };

  const fetchConfig = async () => {
    try {
      const cached = sessionStorage.getItem('monitor_history_enabled');
      if (cached !== null) {
        historyEnabled.value = cached === 'true';
        return;
      }
      const config = await api.getMonitorConfig();
      historyEnabled.value = config.history_enabled;
      sessionStorage.setItem('monitor_history_enabled', String(config.history_enabled));
    } catch (e) {
      console.error('Failed to fetch monitor config', e);
      historyEnabled.value = false;
    }
  };

  const fetchData = async () => {
    try {
      const response = await api.post('/monitor/status');
      const data = await response.json();

      const now = new Date(data.timestamp * 1000).toLocaleTimeString();

      memTotal.value = data.mem_total;
      swapTotal.value = data.swap_total;
      diskTotal.value = data.disk_total;
      diskUsed.value = data.disk_used;

      rTimestamps.value.push(now);
      rCpuData.value.push(parseFloat(data.cpu_percent.toFixed(1)));
      rCpuMaxCoreData.value.push(parseFloat(data.cpu_max_core_percent.toFixed(1)));
      rMemUsedData.value.push(parseFloat((data.mem_used / 1024 / 1024 / 1024).toFixed(2)));
      rSwapUsedData.value.push(parseFloat((data.swap_used / 1024 / 1024 / 1024).toFixed(2)));
      rNetUpData.value.push(parseFloat((data.net_up_speed / 1024).toFixed(1)));
      rNetDownData.value.push(parseFloat((data.net_down_speed / 1024).toFixed(1)));
      rDiskUsedData.value.push(parseFloat((data.disk_used / 1024 / 1024 / 1024).toFixed(2)));

      if (rTimestamps.value.length > maxPoints) {
        rTimestamps.value.shift();
        rCpuData.value.shift();
        rCpuMaxCoreData.value.shift();
        rMemUsedData.value.shift();
        rSwapUsedData.value.shift();
        rNetUpData.value.shift();
        rNetDownData.value.shift();
        rDiskUsedData.value.shift();
      }
    } catch (error) {
      console.error(error);
      message.error('获取监控数据失败');
      stopMonitor();
    }
  };

  const fetchCustomHistory = async (start: number, end: number) => {
    try {
      const data = await api.getMonitorHistory(start, end);
      clearHistoryData();

      data.forEach((item: any) => {
        const date = new Date(item.timestamp * 1000);
        let timeStr = '';
        // Smart formatting based on range duration
        const duration = end - start;
        if (duration <= 3600 * 2) {
          // Less than 2 hours: Show time
          timeStr = date.toLocaleTimeString();
        } else if (duration <= 86400 * 2) {
          // Less than 2 days: Show Hour:Minute
          timeStr = `${date.getHours()}:${date.getMinutes().toString().padStart(2, '0')}`;
        } else {
          // More than 2 days: Show Date Time
          timeStr = `${date.getMonth() + 1}-${date.getDate()} ${date.getHours()}:${date.getMinutes().toString().padStart(2, '0')}`;
        }

        hRawTimestamps.value.push(item.timestamp);
        hTimestamps.value.push(timeStr);
        hCpuData.value.push(item.cpu_percent);
        hCpuMaxCoreData.value.push(item.cpu_max_core_percent);
        hMemUsedData.value.push(item.mem_used);
        hSwapUsedData.value.push(item.swap_used);
        hNetUpData.value.push(item.net_up_speed);
        hNetDownData.value.push(item.net_down_speed);
        hDiskUsedData.value.push(item.disk_used);
      });
    } catch (e) {
      console.error(e);
      message.error('获取历史数据失败');
    }
  };

  const fetchHistory = async (range: '1h' | '24h' | '3d') => {
    const end = Math.floor(Date.now() / 1000);
    let start = end;
    if (range === '1h') start -= 3600;
    else if (range === '24h') start -= 86400;
    else if (range === '3d') start -= 259200;

    await fetchCustomHistory(start, end);
  };

  const startMonitor = () => {
    if (isMonitoring.value) return;
    isMonitoring.value = true;
    fetchData();
    timer.value = setInterval(fetchData, 1000);
  };

  const stopMonitor = () => {
    isMonitoring.value = false;
    if (timer.value) {
      clearInterval(timer.value);
      timer.value = null;
    }
  };

  const toggleMonitor = () => {
    if (isMonitoring.value) {
      stopMonitor();
    } else {
      if (viewMode.value !== 'realtime') {
        setViewMode('realtime');
      } else {
        startMonitor();
      }
    }
  };

  const setViewMode = (mode: 'realtime' | '1h' | '24h' | '3d') => {
    viewMode.value = mode;
    if (mode === 'realtime') {
      // Don't clear realtime data here
      // Do NOT auto-start monitor if it was stopped
    } else {
      // Don't stop monitor here, just fetch history
      fetchHistory(mode);
    }
  };

  const refreshHistory = () => {
    if (viewMode.value !== 'realtime') {
      fetchHistory(viewMode.value);
    }
  };

  return {
    isMonitoring,
    historyEnabled,
    viewMode,
    timestamps,
    hRawTimestamps,
    cpuData,
    cpuMaxCoreData,
    memUsedData,
    memTotal,
    swapUsedData,
    swapTotal,
    netUpData,
    netDownData,
    diskUsedData,
    diskTotal,
    diskUsed,
    diskPercent,
    startMonitor,
    stopMonitor,
    toggleMonitor,
    fetchConfig,
    setViewMode,
    refreshHistory,
    fetchCustomHistory,
  };
});
