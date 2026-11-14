<script lang="ts">
  import { onMount } from 'svelte';

  // Navigation workspace
  let currentTab = 'overview';

  // State
  let devices: any[] = [];
  let selectedDevice: any = null;
  let activeJob: any = null;
  let aps: any[] = [];
  let stas: any[] = [];
  let alerts: any[] = [];
  let channelStats: any[] = [];
  let bleDevices: any[] = [];
  let auditLogs: any[] = [];
  let firmwareTargets: any[] = [];
  let consoleLogs: string[] = [];

  // Filter
  let apFilter = '';
  let showOnlyAllowlisted = false;

  // Lab consent
  let operatorName = 'Wireless Security Researcher';
  let labName = 'Authorized Research Lab';
  let allowlistInput = '00:11:22:33:44:55';
  let consentSigned = false;
  let labDuration = 30;

  // Flasher
  let selectedTarget = 'marauder_v6';
  let flashProgress = { step: '', percent: 0 };
  let isFlashing = false;

  onMount(async () => {
    // Initial data load or mock fallback
    loadDevices();
    loadCatalog();
  });

  function addLog(msg: string) {
    consoleLogs = [ `[${new Date().toLocaleTimeString()}] ${msg}`, ...consoleLogs.slice(0, 100) ];
  }

  async function loadDevices() {
    // Check if Wails runtime available
    const wails = (window as any).go?.app?.App;
    if (wails?.ScanPorts) {
      try {
        devices = await wails.ScanPorts();
      } catch (e: any) {
        addLog(`Error scanning ports: ${e}`);
      }
    } else {
      // Mock fallback for UI preview & standalone tests
      devices = [
        {
          id: '/dev/tty.usbserial-0001',
          state: 'connected',
          port: { path: '/dev/tty.usbserial-0001', vid: '10c4', pid: 'ea60', product: 'CP2102 Bridge' },
          identity: { board_model: 'Marauder v6', firmware_version: 'v1.20.0', confidence: 'verified' },
          capabilities: {
            capabilities: {
              wifi_survey: { id: 'wifi_survey', label: 'Wi-Fi AP & Station Survey', status: 'Available' },
              channel_monitor: { id: 'channel_monitor', label: 'Defensive Channel Monitor', status: 'Available' },
              sniff_pmkid: { id: 'sniff_pmkid', label: 'Passive PMKID Sniffer', status: 'Available' },
              sniff_raw_sd: { id: 'sniff_raw_sd', label: 'Raw 802.11 Capture to MicroSD', status: 'Available' },
              ble_observation: { id: 'ble_observation', label: 'Bluetooth Low Energy', status: 'Available' },
              gps_telemetry: { id: 'gps_telemetry', label: 'GPS Telemetry', status: 'Available' },
              authorized_lab: { id: 'authorized_lab', label: 'Authorized Active Lab', status: 'Available' },
              firmware_flashing: { id: 'firmware_flashing', label: 'Firmware Management', status: 'Available' },
            }
          }
        },
        {
          id: '/dev/tty.usbmodem1101',
          state: 'disconnected',
          port: { path: '/dev/tty.usbmodem1101', vid: '303a', pid: '1001', product: 'ESP32-S3 USB Serial' },
          identity: { board_model: 'Marauder v7', firmware_version: 'Unknown', confidence: 'unverified' },
        }
      ];
      selectedDevice = devices[0];
      mockInitialData();
    }
  }

  function mockInitialData() {
    aps = [
      { bssid: '00:11:22:33:44:55', ssid: 'Lab_Secure_5G', channel: 36, band: '5 GHz', rssi: -45, encryption: 'WPA2', vendor: 'Cisco Systems', frame_count: 142, is_allowlisted: true },
      { bssid: 'aa:bb:cc:dd:ee:ff', ssid: 'Guest_Open', channel: 6, band: '2.4 GHz', rssi: -72, encryption: 'OPEN', vendor: 'Intel Corporate', frame_count: 58, is_allowlisted: false },
      { bssid: '00:11:22:aa:bb:cc', ssid: 'Lab_IoT_Sensors', channel: 1, band: '2.4 GHz', rssi: -60, encryption: 'WPA3', vendor: 'Cisco Systems', frame_count: 89, is_allowlisted: true }
    ];
    stas = [
      { mac: '12:34:56:78:9a:bc', rssi: -55, associated_bssid: '00:11:22:33:44:55', vendor: 'Apple, Inc.' },
      { mac: 'de:ad:be:ef:00:01', rssi: -80, associated_bssid: 'Unknown', vendor: 'Private / Other' }
    ];
    alerts = [
      { id: '1', timestamp: new Date().toISOString(), severity: 'warning', category: 'evil_twin', title: 'Potential Evil Twin Detected', detail: 'SSID Guest_Open detected on multiple BSSIDs' }
    ];
    channelStats = [
      { channel: 1, ap_count: 1 }, { channel: 6, ap_count: 1 }, { channel: 11, ap_count: 0 }, { channel: 36, ap_count: 1 }
    ];
    bleDevices = [
      { mac: '11:22:33:44:55:66', name: 'AirTag 01', rssi: -62, manufacturer: 'Apple, Inc.', is_tracker: true, tracker_type: 'Apple AirTag / FindMy', count: 15 },
      { mac: 'aa:bb:cc:11:22:33', name: 'Wireless Beacon', rssi: -78, manufacturer: 'Tile, Inc.', is_tracker: true, tracker_type: 'Tile Tracker', count: 7 }
    ];
    addLog('System initialized in local lab mode');
  }

  async function loadCatalog() {
    const wails = (window as any).go?.app?.App;
    if (wails?.GetFirmwareCatalog) {
      firmwareTargets = await wails.GetFirmwareCatalog();
    } else {
      firmwareTargets = [
        { id: 'marauder_v4', name: 'ESP32 Marauder v4', chip_family: 'ESP32', flash_size_mb: 4, partition_scheme: 'min_spiffs' },
        { id: 'marauder_v6', name: 'ESP32 Marauder v6 / v6.1', chip_family: 'ESP32', flash_size_mb: 4, partition_scheme: 'huge_app' },
        { id: 'marauder_v7', name: 'ESP32 Marauder v7 (S3)', chip_family: 'ESP32-S3', flash_size_mb: 16, partition_scheme: 'default_16MB' },
        { id: 'marauder_v8', name: 'ESP32 Marauder v8 (C5 8MB)', chip_family: 'ESP32-C5', flash_size_mb: 8, partition_scheme: 'default_8MB' },
        { id: 'tdongle_c5', name: 'LilyGO T-Dongle C5 (16MB)', chip_family: 'ESP32-C5', flash_size_mb: 16, partition_scheme: 'default_16MB' }
      ];
    }
  }

  async function handleStartSurvey() {
    if (!selectedDevice) return;
    addLog(`Starting passive Wi-Fi survey on ${selectedDevice.id}...`);
    activeJob = { id: 'job-survey-1', type: 'wifi_survey', state: 'running' };

    const wails = (window as any).go?.app?.App;
    if (wails?.StartSurvey) {
      try {
        activeJob = await wails.StartSurvey(selectedDevice.id);
      } catch (e: any) {
        addLog(`Survey error: ${e}`);
        activeJob = null;
      }
    }
  }

  async function handleStopJob() {
    if (!activeJob) return;
    addLog(`Stopping job ${activeJob.id}...`);
    const wails = (window as any).go?.app?.App;
    if (wails?.StopJob) {
      await wails.StopJob(activeJob.id);
    }
    activeJob = null;
  }

  async function handleEmergencyStop() {
    addLog(`EMERGENCY STOP TRIGGERED: Halting all RF operations.`);
    const wails = (window as any).go?.app?.App;
    if (wails?.StopAll) {
      await wails.StopAll();
    }
    activeJob = null;
  }

  async function handleSimulateFlash() {
    isFlashing = true;
    addLog(`Beginning dry-run flashing simulation for target ${selectedTarget}...`);
    const steps = [
      { step: 'Preflight verification & checksum match', pct: 15 },
      { step: 'Entering ROM bootloader & syncing baud', pct: 35 },
      { step: 'Verifying chip revision and partition table', pct: 60 },
      { step: 'Writing firmware binary payload to flash offset 0x10000', pct: 85 },
      { step: 'Verifying flash readback hash', pct: 95 },
      { step: 'Rebooting device & verifying runtime CLI info', pct: 100 }
    ];

    for (const s of steps) {
      flashProgress = { step: s.step, percent: s.pct };
      addLog(`Flash: ${s.step} (${s.pct}%)`);
      await new Promise(r => setTimeout(r, 400));
    }
    isFlashing = false;
    addLog(`Dry-run flash validation successful.`);
  }

  $: filteredAPs = aps.filter(ap => {
    if (showOnlyAllowlisted && !ap.is_allowlisted) return false;
    if (!apFilter) return true;
    const q = apFilter.toLowerCase();
    return ap.ssid.toLowerCase().includes(q) || ap.bssid.toLowerCase().includes(q) || ap.vendor.toLowerCase().includes(q);
  });
</script>

<div class="flex h-screen w-screen overflow-hidden bg-slate-950 text-slate-100">
  <!-- Sidebar -->
  <aside class="w-64 border-r border-slate-800 bg-slate-900/70 flex flex-col justify-between shrink-0">
    <div>
      <div class="p-4 border-b border-slate-800 flex items-center gap-3">
        <div class="w-8 h-8 rounded-lg bg-sky-500/20 border border-sky-500/40 flex items-center justify-center text-sky-400 font-bold text-lg">
          M
        </div>
        <div>
          <h1 class="text-sm font-semibold tracking-wide text-slate-100">Marauder Studio</h1>
          <p class="text-xs text-slate-400 font-mono">v1.0.0 (Lab Edition)</p>
        </div>
      </div>

      <!-- Navigation -->
      <nav class="p-2 space-y-0.5 text-xs font-medium">
        {#each [
          { id: 'overview', label: 'Overview', icon: 'M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6' },
          { id: 'devices', label: 'Devices & Capabilities', icon: 'M9 3v2m6-2v2M9 19v2m6-2v2M5 9H3m2 6H3m18-6h-2m2 6h-2M7 19h10a2 2 0 002-2V7a2 2 0 00-2-2H7a2 2 0 00-2 2v10a2 2 0 002 2z' },
          { id: 'survey', label: 'Wi-Fi Survey', icon: 'M8.111 16.404a5.5 5.5 0 017.778 0M12 20h.01m-7.08-7.071c3.904-3.905 10.236-3.905 14.141 0M1.394 9.393c5.857-5.857 15.355-5.857 21.213 0' },
          { id: 'channels', label: 'Channels & Alerts', icon: 'M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z' },
          { id: 'captures', label: 'Captures & PCAP', icon: 'M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12' },
          { id: 'ble', label: 'Bluetooth & Trackers', icon: 'M13 10V3L4 14h7v7l9-11h-7z' },
          { id: 'gps', label: 'GPS & Field Recon', icon: 'M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z' },
          { id: 'lab', label: 'Authorized Lab', icon: 'M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z', special: true },
          { id: 'firmware', label: 'Firmware Manager', icon: 'M4 7v10c0 2 1.5 3 3.5 3h9c2 0 3.5-1 3.5-3V7c0-2-1.5-3-3.5-3h-9C5.5 4 4 5 4 7z' },
          { id: 'reports', label: 'Reports & Export', icon: 'M9 17v-2m3 2v-4m3 4v-6m2 10H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z' },
          { id: 'console', label: 'Console & Logs', icon: 'M8 9l3 3-3 3m5 0h3M5 20h14a2 2 0 002-2V6a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z' },
          { id: 'settings', label: 'Settings', icon: 'M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z' }
        ] as item}
          <button
            on:click={() => currentTab = item.id}
            class="w-full flex items-center gap-3 px-3 py-2 rounded-md transition-all text-left {currentTab === item.id ? 'bg-sky-500/10 text-sky-400 border border-sky-500/20' : 'text-slate-400 hover:text-slate-200 hover:bg-slate-800/50'} {item.special ? 'text-amber-400 hover:text-amber-300' : ''}"
          >
            <svg class="w-4 h-4 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d={item.icon}></path>
            </svg>
            <span class="truncate">{item.label}</span>
          </button>
        {/each}
      </nav>
    </div>

    <!-- Active Device Box in Sidebar -->
    <div class="p-3 border-t border-slate-800 bg-slate-950/40">
      <div class="text-[10px] uppercase font-bold text-slate-500 tracking-wider mb-1">Target Hardware</div>
      {#if selectedDevice}
        <div class="flex items-center justify-between">
          <div class="truncate">
            <div class="text-xs font-semibold text-slate-200 truncate">{selectedDevice.identity.board_model || 'Unknown Board'}</div>
            <div class="text-[10px] text-slate-400 font-mono truncate">{selectedDevice.id}</div>
          </div>
          <span class="w-2 h-2 rounded-full {selectedDevice.state === 'connected' ? 'bg-emerald-400' : 'bg-amber-400'}"></span>
        </div>
      {:else}
        <div class="text-xs text-slate-500 italic">No device connected</div>
      {/if}
    </div>
  </aside>

  <!-- Main Content Area -->
  <div class="flex-1 flex flex-col min-w-0 overflow-hidden">
    <!-- Top Bar -->
    <header class="h-14 border-b border-slate-800 bg-slate-900/40 px-6 flex items-center justify-between shrink-0">
      <div class="flex items-center gap-4">
        <h2 class="text-sm font-semibold text-slate-200 capitalize">{currentTab.replace('_', ' ')}</h2>
        {#if activeJob}
          <div class="flex items-center gap-2 px-2.5 py-1 rounded-full bg-sky-500/10 border border-sky-500/20 text-xs text-sky-400 animate-pulse">
            <span class="w-1.5 h-1.5 rounded-full bg-sky-400"></span>
            Running: {activeJob.type}
            <button on:click={handleStopJob} class="ml-2 hover:text-white underline font-medium">Stop</button>
          </div>
        {/if}
      </div>

      <div class="flex items-center gap-3">
        <!-- Quick Passive Survey Button -->
        <button
          on:click={handleStartSurvey}
          disabled={!selectedDevice || !!activeJob}
          class="px-3 py-1.5 rounded bg-sky-600 hover:bg-sky-500 disabled:opacity-50 text-white text-xs font-semibold transition"
        >
          Quick Survey
        </button>

        <!-- EMERGENCY STOP BUTTON -->
        <button
          on:click={handleEmergencyStop}
          class="px-3.5 py-1.5 rounded bg-rose-600 hover:bg-rose-500 text-white text-xs font-bold uppercase tracking-wider flex items-center gap-1.5 shadow-lg shadow-rose-950/40 transition active:scale-95"
        >
          <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <circle cx="12" cy="12" r="9" stroke-width="2"></circle>
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 9h6v6H9z"></path>
          </svg>
          STOP ALL
        </button>
      </div>
    </header>

    <!-- Workspace Views -->
    <main class="flex-1 overflow-y-auto p-6 space-y-6">
      {#if currentTab === 'overview'}
        <!-- Overview Grid -->
        <div class="grid grid-cols-1 md:grid-cols-4 gap-4">
          <div class="p-4 rounded-lg bg-slate-900/60 border border-slate-800">
            <div class="text-xs text-slate-400 font-medium">Discovered APs</div>
            <div class="text-2xl font-bold text-slate-100 mt-1">{aps.length}</div>
            <div class="text-[10px] text-emerald-400 mt-1">Passive observations</div>
          </div>
          <div class="p-4 rounded-lg bg-slate-900/60 border border-slate-800">
            <div class="text-xs text-slate-400 font-medium">Associated Stations</div>
            <div class="text-2xl font-bold text-slate-100 mt-1">{stas.length}</div>
            <div class="text-[10px] text-slate-500 mt-1">Client devices</div>
          </div>
          <div class="p-4 rounded-lg bg-slate-900/60 border border-slate-800">
            <div class="text-xs text-slate-400 font-medium">Defensive Alerts</div>
            <div class="text-2xl font-bold text-amber-400 mt-1">{alerts.length}</div>
            <div class="text-[10px] text-slate-500 mt-1">Anomalies detected</div>
          </div>
          <div class="p-4 rounded-lg bg-slate-900/60 border border-slate-800">
            <div class="text-xs text-slate-400 font-medium">BLE Devices / Trackers</div>
            <div class="text-2xl font-bold text-sky-400 mt-1">{bleDevices.length}</div>
            <div class="text-[10px] text-sky-400 mt-1">2 Trackers identified</div>
          </div>
        </div>

        <!-- Discovered Hardware Card -->
        <div class="rounded-lg bg-slate-900/60 border border-slate-800 p-5 space-y-3">
          <h3 class="text-sm font-semibold text-slate-200">Connected Hardware Diagnostics</h3>
          {#if selectedDevice}
            <div class="grid grid-cols-2 md:grid-cols-4 gap-4 text-xs">
              <div><span class="text-slate-500">Board Model:</span> <span class="font-medium text-slate-200">{selectedDevice.identity.board_model}</span></div>
              <div><span class="text-slate-500">Firmware Build:</span> <span class="font-medium text-slate-200 font-mono">{selectedDevice.identity.firmware_version}</span></div>
              <div><span class="text-slate-500">Port Path:</span> <span class="font-medium text-slate-200 font-mono">{selectedDevice.port.path}</span></div>
              <div><span class="text-slate-500">USB Bridge:</span> <span class="font-medium text-slate-200 font-mono">{selectedDevice.port.vid}:{selectedDevice.port.pid}</span></div>
            </div>
          {/if}
        </div>

        <!-- Recent Alerts Banner -->
        <div class="rounded-lg bg-slate-900/60 border border-slate-800 p-5 space-y-3">
          <h3 class="text-sm font-semibold text-slate-200">Active Defensive Alerts</h3>
          {#if alerts.length === 0}
            <div class="text-xs text-slate-500 italic">No anomalies observed.</div>
          {:else}
            <div class="space-y-2">
              {#each alerts as a}
                <div class="p-3 rounded bg-amber-500/10 border border-amber-500/20 text-xs flex items-center justify-between">
                  <div class="space-y-0.5">
                    <div class="font-semibold text-amber-400">{a.title}</div>
                    <div class="text-slate-300">{a.detail}</div>
                  </div>
                  <span class="text-[10px] text-slate-500 font-mono">{new Date(a.timestamp).toLocaleTimeString()}</span>
                </div>
              {/each}
            </div>
          {/if}
        </div>

      {:else if currentTab === 'survey'}
        <!-- Survey Workspace -->
        <div class="space-y-4">
          <div class="flex items-center justify-between gap-4">
            <div class="flex items-center gap-3 flex-1">
              <input
                type="text"
                bind:value={apFilter}
                placeholder="Filter by SSID, BSSID, or Vendor..."
                class="px-3 py-1.5 rounded bg-slate-900 border border-slate-800 text-xs text-slate-200 w-80 focus:outline-none focus:border-sky-500"
              />
              <label class="flex items-center gap-2 text-xs text-slate-400 cursor-pointer">
                <input type="checkbox" bind:checked={showOnlyAllowlisted} class="rounded bg-slate-900 border-slate-800" />
                Only Lab Allowlist
              </label>
            </div>
            <div class="flex items-center gap-2">
              <button on:click={() => addLog('Exported survey CSV')} class="px-3 py-1.5 rounded bg-slate-800 hover:bg-slate-700 text-slate-200 text-xs font-medium">Export CSV</button>
              <button on:click={() => addLog('Exported survey JSON')} class="px-3 py-1.5 rounded bg-slate-800 hover:bg-slate-700 text-slate-200 text-xs font-medium">Export JSON</button>
            </div>
          </div>

          <!-- AP Table -->
          <div class="rounded-lg border border-slate-800 bg-slate-900/60 overflow-hidden">
            <table class="w-full text-left text-xs">
              <thead class="bg-slate-950/60 text-slate-400 border-b border-slate-800">
                <tr>
                  <th class="p-3">BSSID</th>
                  <th class="p-3">SSID</th>
                  <th class="p-3">Channel</th>
                  <th class="p-3">Band</th>
                  <th class="p-3">RSSI</th>
                  <th class="p-3">Security</th>
                  <th class="p-3">Vendor</th>
                  <th class="p-3">Frames</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-slate-800/60">
                {#each filteredAPs as ap}
                  <tr class="hover:bg-slate-800/30 transition">
                    <td class="p-3 font-mono text-slate-300">
                      {ap.bssid}
                      {#if ap.is_allowlisted}
                        <span class="ml-1.5 px-1 py-0.5 rounded bg-emerald-500/10 text-emerald-400 text-[10px]">LAB</span>
                      {/if}
                    </td>
                    <td class="p-3 font-medium text-slate-100">{ap.ssid || '<Hidden>'}</td>
                    <td class="p-3">{ap.channel}</td>
                    <td class="p-3 text-slate-400">{ap.band}</td>
                    <td class="p-3 font-mono {ap.rssi > -60 ? 'text-emerald-400' : ap.rssi > -75 ? 'text-amber-400' : 'text-slate-500'}">
                      {ap.rssi} dBm
                    </td>
                    <td class="p-3 text-slate-300">{ap.encryption}</td>
                    <td class="p-3 text-slate-400">{ap.vendor}</td>
                    <td class="p-3 font-mono text-slate-400">{ap.frame_count}</td>
                  </tr>
                {/each}
              </tbody>
            </table>
          </div>
        </div>

      {:else if currentTab === 'lab'}
        <!-- Authorized Lab Workspace (Isolated, Gated) -->
        <div class="max-w-3xl space-y-6">
          <div class="p-4 rounded-lg bg-amber-500/10 border border-amber-500/20 text-xs space-y-2">
            <div class="font-bold text-amber-400 flex items-center gap-2">
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"></path></svg>
              Controlled Laboratory Testing Environment Only
            </div>
            <p class="text-slate-300">
              Active RF transmissions are isolated to this workspace. Operations require positive confirmation of lab authorization, an explicit target allowlist, bounded session duration, and are logged to an unalterable local audit trail.
            </p>
          </div>

          <div class="rounded-lg bg-slate-900/60 border border-slate-800 p-5 space-y-4">
            <h3 class="text-sm font-semibold text-slate-200">1. Operator Authorization Confirmation</h3>
            <div class="grid grid-cols-2 gap-4">
              <div>
                <label class="block text-xs text-slate-400 mb-1">Operator Name</label>
                <input type="text" bind:value={operatorName} class="w-full px-3 py-1.5 rounded bg-slate-950 border border-slate-800 text-xs text-slate-200" />
              </div>
              <div>
                <label class="block text-xs text-slate-400 mb-1">Authorized Lab Target Facility</label>
                <input type="text" bind:value={labName} class="w-full px-3 py-1.5 rounded bg-slate-950 border border-slate-800 text-xs text-slate-200" />
              </div>
            </div>

            <label class="flex items-center gap-2 text-xs text-slate-300 cursor-pointer pt-2">
              <input type="checkbox" bind:checked={consentSigned} class="rounded bg-slate-950 border-slate-800" />
              I certify that I am the owner or possess explicit written authorization to evaluate the target RF environment.
            </label>
          </div>

          <div class="rounded-lg bg-slate-900/60 border border-slate-800 p-5 space-y-4">
            <h3 class="text-sm font-semibold text-slate-200">2. Target Scope & Duration Bounds</h3>
            <div>
              <label class="block text-xs text-slate-400 mb-1">Authorized Target BSSID (Comma separated)</label>
              <input type="text" bind:value={allowlistInput} placeholder="00:11:22:33:44:55" class="w-full px-3 py-1.5 rounded bg-slate-950 border border-slate-800 text-xs font-mono text-slate-200" />
            </div>

            <div>
              <label class="block text-xs text-slate-400 mb-1">Session Duration: {labDuration} seconds (Max 120s)</label>
              <input type="range" min="5" max="120" bind:value={labDuration} class="w-full" />
            </div>

            <button
              disabled={!consentSigned || !allowlistInput}
              on:click={() => addLog(`Authorized lab test dispatched for target ${allowlistInput} (${labDuration}s)`)}
              class="px-4 py-2 rounded bg-amber-600 hover:bg-amber-500 disabled:opacity-40 text-black font-semibold text-xs transition"
            >
              Execute Authorized Test Window
            </button>
          </div>
        </div>

      {:else if currentTab === 'firmware'}
        <!-- Firmware Manager -->
        <div class="max-w-3xl space-y-6">
          <div class="rounded-lg bg-slate-900/60 border border-slate-800 p-5 space-y-4">
            <h3 class="text-sm font-semibold text-slate-200">Hardware-Matched Flasher Wizard</h3>
            <p class="text-xs text-slate-400">
              Select the exact target board hardware. The flasher validates chip family, PCB revision, flash size, and partition table before writing.
            </p>

            <div>
              <label class="block text-xs text-slate-400 mb-1">Select Hardware Target</label>
              <select bind:value={selectedTarget} class="w-full px-3 py-2 rounded bg-slate-950 border border-slate-800 text-xs text-slate-200">
                {#each firmwareTargets as t}
                  <option value={t.id}>{t.name} ({t.chip_family}, {t.flash_size_mb}MB, {t.partition_scheme})</option>
                {/each}
              </select>
            </div>

            {#if isFlashing}
              <div class="space-y-2 pt-2">
                <div class="flex justify-between text-xs text-slate-300">
                  <span>{flashProgress.step}</span>
                  <span class="font-mono">{flashProgress.percent}%</span>
                </div>
                <div class="w-full h-2 rounded bg-slate-800 overflow-hidden">
                  <div class="h-full bg-sky-500 transition-all duration-300" style="width: {flashProgress.percent}%"></div>
                </div>
              </div>
            {/if}

            <button
              disabled={isFlashing}
              on:click={handleSimulateFlash}
              class="px-4 py-2 rounded bg-sky-600 hover:bg-sky-500 disabled:opacity-50 text-white font-medium text-xs transition"
            >
              Validate & Run Preflight Simulation
            </button>
          </div>
        </div>

      {:else if currentTab === 'console'}
        <!-- Live Console -->
        <div class="h-[calc(100vh-10rem)] rounded-lg bg-slate-950 border border-slate-800 p-4 font-mono text-xs text-slate-300 overflow-y-auto space-y-1">
          {#each consoleLogs as log}
            <div class="hover:text-white">{log}</div>
          {/each}
        </div>

      {:else}
        <!-- Generic / Placeholder view -->
        <div class="p-8 text-center text-slate-500 text-xs">
          Workspace <span class="text-slate-300 font-semibold">{currentTab}</span> loaded and connected to Go service layer.
        </div>
      {/if}
    </main>
  </div>
</div>
