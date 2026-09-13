<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import {
  Wallet,
  ArrowDownLeft,
  ArrowUpRight,
  Activity,
  CheckCircle2,
  AlertCircle,
  PiggyBank,
  RefreshCw,
  LayoutDashboard,
  ArrowLeftRight,
  PieChart,
  WifiOff,
  Clock,
  Calendar,
  Flame,
  ShieldAlert,
  LogOut,
  ChevronDown,
  Trash2,
  AlertTriangle,
  Download,
} from '@lucide/vue'
import AccountsView from './views/AccountsView.vue'
import TransactionsView from './views/TransactionsView.vue'
import BudgetsView from './views/BudgetsView.vue'
import LoginView from './views/LoginView.vue'
import SetupView from './views/SetupView.vue'

const authStatus = ref({
  initialized: true,
  authenticated: false,
  username: '',
})
const authChecking = ref(true)

const health = ref(null)
const loading = ref(true)
const error = ref(null)
const isOnline = ref(typeof navigator !== 'undefined' ? navigator.onLine : true)
const activeTab = ref('dashboard')
const netWorth = ref(0)
const totalMonthlyExpense = ref(0)
const totalMonthlyIncome = ref(0)
const burnRate = ref({
  next_payday_remain: 0,
  next_payday_date: '',
  cycle_start_date: '',
  days_elapsed: 0,
  remain_funds: 0,
  grand_total_expenses: 0,
  cycle_income: 0,
  prospect_daily_limit: 0,
  average_daily_expense: 0,
  burn_rate_status: 'safe',
  pillar_breakdown: {
    needs: { total: 0, percentage: 0 },
    wants: { total: 0, percentage: 0 },
    savings: { total: 0, percentage: 0 },
  },
  category_breakdown: [],
})

const updateOnlineStatus = () => {
  isOnline.value = navigator.onLine
}

const checkHealth = async () => {
  if (!navigator.onLine) {
    error.value = 'Tidak ada koneksi internet'
    loading.value = false
    return
  }
  loading.value = true
  error.value = null
  try {
    const res = await fetch('/api/health')
    if (!res.ok) throw new Error(`HTTP ${res.status}`)
    health.value = await res.json()
  } catch (err) {
    error.value = err.message
  } finally {
    loading.value = false
  }
}

const checkAuthStatus = async () => {
  authChecking.value = true
  try {
    const res = await fetch('/api/auth/status')
    if (res.ok) {
      const data = await res.json()
      authStatus.value = data
      if (data.authenticated) {
        refreshAll()
      }
    }
  } catch (err) {
    console.error('Gagal memeriksa autentikasi:', err)
  } finally {
    authChecking.value = false
  }
}

const handleAuthSuccess = username => {
  authStatus.value.initialized = true
  authStatus.value.authenticated = true
  authStatus.value.username = username
  refreshAll()
}

const isUserMenuOpen = ref(false)
const showLogoutModal = ref(false)
const showResetModal = ref(false)
const isResetting = ref(false)

const openLogoutModal = () => {
  isUserMenuOpen.value = false
  showLogoutModal.value = true
}

const confirmLogout = async () => {
  showLogoutModal.value = false
  await handleLogout()
}

const downloadBackup = () => {
  isUserMenuOpen.value = false
  const link = document.createElement('a')
  link.href = '/api/settings/backup'
  link.setAttribute('download', '')
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
}

const openResetModal = () => {
  isUserMenuOpen.value = false
  showResetModal.value = true
}

const confirmResetData = async () => {
  isResetting.value = true
  try {
    const res = await fetch('/api/settings/reset-data', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
    })
    if (res.ok) {
      showResetModal.value = false
      await refreshAll()
    } else {
      const errText = await res.text()
      alert('Gagal mereset data: ' + errText)
    }
  } catch (err) {
    alert('Terjadi kesalahan saat mereset data: ' + err.message)
  } finally {
    isResetting.value = false
  }
}

const handleLogout = async () => {
  try {
    await fetch('/api/auth/logout', { method: 'POST' })
  } catch (err) {
    console.error('Logout error:', err)
  } finally {
    authStatus.value.authenticated = false
    authStatus.value.username = ''
  }
}

const fetchNetWorth = async () => {
  try {
    const res = await fetch('/api/accounts')
    if (res.status === 401) {
      authStatus.value.authenticated = false
      return
    }
    if (res.ok) {
      const data = await res.json()
      netWorth.value = data.total_net_worth || 0
    }
  } catch (err) {
    console.error(err)
  }
}

const fetchMonthlyTransactions = async () => {
  try {
    const currentMonth = new Date().toISOString().slice(0, 7)
    const res = await fetch(`/api/transactions?month=${currentMonth}`)
    if (res.status === 401) {
      authStatus.value.authenticated = false
      return
    }
    if (res.ok) {
      const data = await res.json()
      totalMonthlyExpense.value = data.total_expense || 0
      totalMonthlyIncome.value = data.total_income || 0
    }
  } catch (err) {
    console.error(err)
  }
}

const fetchBurnRate = async () => {
  try {
    const res = await fetch('/api/analytics/burn-rate')
    if (res.status === 401) {
      authStatus.value.authenticated = false
      return
    }
    if (res.ok) {
      const data = await res.json()
      if (data && data.pillar_breakdown) {
        burnRate.value = data
      }
    }
  } catch (err) {
    console.error(err)
  }
}

const burnRateUsagePercent = computed(() => {
  if (!burnRate.value.prospect_daily_limit || burnRate.value.prospect_daily_limit <= 0) return 0
  return Math.round(
    (burnRate.value.average_daily_expense / burnRate.value.prospect_daily_limit) * 100
  )
})

const formatIndonesianDate = dateStr => {
  if (!dateStr) return '-'
  const date = new Date(dateStr + 'T00:00:00')
  return date.toLocaleDateString('id-ID', {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
  })
}

const switchTab = tabId => {
  activeTab.value = tabId
  refreshAll()
}

const refreshAll = () => {
  checkHealth()
  fetchNetWorth()
  fetchMonthlyTransactions()
  fetchBurnRate()
}

onMounted(() => {
  window.addEventListener('online', updateOnlineStatus)
  window.addEventListener('offline', updateOnlineStatus)
  checkAuthStatus()
})

onUnmounted(() => {
  window.removeEventListener('online', updateOnlineStatus)
  window.removeEventListener('offline', updateOnlineStatus)
})

// Format Rupiah helper
const formatRupiah = val => {
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    minimumFractionDigits: 0,
    maximumFractionDigits: 0,
  }).format(val || 0)
}

const navTabs = [
  { id: 'dashboard', label: 'Dashboard', icon: LayoutDashboard },
  { id: 'transactions', label: 'Transaksi', icon: ArrowLeftRight },
  { id: 'accounts', label: 'Dompet', icon: Wallet },
  { id: 'budgets', label: 'Anggaran', icon: PieChart },
]
</script>

<template>
  <!-- Loading Checking Auth State -->
  <div
    v-if="authChecking"
    class="min-h-screen bg-slate-50 flex flex-col items-center justify-center space-y-3"
  >
    <div
      class="w-12 h-12 rounded-2xl bg-emerald-600 flex items-center justify-center text-white shadow-lg shadow-emerald-200 animate-pulse"
    >
      <PiggyBank class="w-7 h-7" />
    </div>
    <span class="text-xs font-semibold text-slate-500">Memeriksa Sesi Neraca...</span>
  </div>

  <!-- Screen Setup Akun Pertama -->
  <SetupView v-else-if="!authStatus.initialized" @setup-success="handleAuthSuccess" />

  <!-- Screen Login -->
  <LoginView v-else-if="!authStatus.authenticated" @login-success="handleAuthSuccess" />

  <!-- Main App Shell -->
  <div v-else class="min-h-screen bg-slate-50 flex flex-col text-slate-800">
    <!-- Offline Alert Banner -->
    <div
      v-if="!isOnline"
      class="bg-amber-500 text-white text-xs font-semibold px-4 py-2 flex items-center justify-center space-x-2 sticky top-0 z-50 shadow-sm"
    >
      <WifiOff class="w-4 h-4 animate-bounce" />
      <span>Mode Offline — Anda sedang offline. Fitur PWA tetap dapat dibuka.</span>
    </div>

    <!-- Header Navbar -->
    <header class="bg-white border-b border-slate-200 sticky top-0 z-30 shadow-sm">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 h-16 flex items-center justify-between">
        <div class="flex items-center space-x-3">
          <div
            class="w-10 h-10 rounded-xl bg-emerald-600 flex items-center justify-center text-white shadow-md shadow-emerald-200 flex-shrink-0"
          >
            <PiggyBank class="w-6 h-6" />
          </div>
          <div>
            <h1 class="text-xl font-bold tracking-tight text-slate-900 leading-tight">Neraca</h1>
            <p class="text-xs text-slate-500 font-medium hidden sm:block">
              Personal Financial Dashboard
            </p>
          </div>
        </div>

        <!-- Desktop Navigation Links -->
        <nav aria-label="Navigasi Utama" class="hidden md:flex items-center space-x-1">
          <button
            v-for="tab in navTabs"
            :key="tab.id"
            type="button"
            :class="[
              activeTab === tab.id
                ? 'bg-emerald-50 text-emerald-700 font-semibold'
                : 'text-slate-600 hover:text-slate-900 hover:bg-slate-100 font-medium',
            ]"
            class="px-3.5 py-2 rounded-lg text-sm transition flex items-center space-x-2"
            @click="switchTab(tab.id)"
          >
            <component :is="tab.icon" class="w-4 h-4" />
            <span>{{ tab.label }}</span>
          </button>
        </nav>

        <!-- Server & Network Status Badge -->
        <div class="flex items-center space-x-2 sm:space-x-3">
          <div
            v-if="health && health.status === 'ok' && isOnline"
            class="inline-flex items-center space-x-1.5 px-2.5 sm:px-3 py-1 rounded-full text-xs font-semibold bg-emerald-50 text-emerald-700 border border-emerald-200"
          >
            <span class="w-2 h-2 rounded-full bg-emerald-500 animate-pulse"></span>
            <span class="hidden sm:inline">Server Online (WAL)</span>
            <span class="sm:hidden">Online</span>
          </div>

          <div
            v-else-if="error || !isOnline"
            class="inline-flex items-center space-x-1.5 px-2.5 sm:px-3 py-1 rounded-full text-xs font-semibold bg-rose-50 text-rose-700 border border-rose-200"
          >
            <AlertCircle class="w-3.5 h-3.5" />
            <span class="hidden sm:inline">{{ error || 'Offline' }}</span>
            <span class="sm:hidden">Offline</span>
          </div>

          <button
            type="button"
            aria-label="Segarkan status server"
            :disabled="loading"
            title="Cek Status Server"
            class="p-2.5 rounded-lg text-slate-500 hover:text-slate-800 hover:bg-slate-100 transition min-w-[44px] min-h-[44px] flex items-center justify-center"
            @click="refreshAll"
          >
            <RefreshCw class="w-4 h-4" :class="{ 'animate-spin': loading }" />
          </button>

          <!-- User Dropdown Menu -->
          <div class="relative pl-2 border-l border-slate-200">
            <button
              type="button"
              aria-label="Buka menu pengguna"
              aria-haspopup="true"
              :aria-expanded="isUserMenuOpen"
              class="flex items-center space-x-1.5 px-2.5 py-1.5 rounded-xl border border-slate-200 hover:bg-slate-50 transition text-xs font-semibold text-slate-700 min-h-[38px] cursor-pointer"
              @click="isUserMenuOpen = !isUserMenuOpen"
            >
              <div
                class="w-6 h-6 rounded-full bg-emerald-100 text-emerald-700 flex items-center justify-center font-bold text-xs flex-shrink-0"
              >
                {{ (authStatus.username || 'P').charAt(0).toUpperCase() }}
              </div>
              <span class="hidden sm:inline max-w-[100px] truncate">{{
                authStatus.username || 'Pengguna'
              }}</span>
              <ChevronDown
                class="w-3.5 h-3.5 text-slate-400 transition-transform duration-200"
                :class="{ 'rotate-180': isUserMenuOpen }"
              />
            </button>

            <!-- Backdrop to close dropdown on outside click -->
            <div
              v-if="isUserMenuOpen"
              class="fixed inset-0 z-40"
              @click="isUserMenuOpen = false"
            ></div>

            <!-- Dropdown Card -->
            <div
              v-if="isUserMenuOpen"
              class="absolute right-0 mt-2 w-60 bg-white rounded-2xl shadow-xl border border-slate-200 py-1.5 z-50 overflow-hidden animate-in fade-in zoom-in-95 duration-100"
            >
              <div class="px-3.5 py-2.5 bg-slate-50 border-b border-slate-100">
                <span class="text-[10px] uppercase font-bold tracking-wider text-slate-400 block"
                  >Akun Pengelola</span
                >
                <span class="text-xs font-bold text-slate-800 truncate block">{{
                  authStatus.username || 'Pengguna'
                }}</span>
              </div>

              <div class="py-1">
                <!-- Unduh Cadangan Database -->
                <button
                  type="button"
                  class="w-full text-left px-3.5 py-2.5 text-xs text-slate-700 hover:bg-slate-50 flex items-center space-x-2.5 transition font-medium cursor-pointer"
                  @click="downloadBackup"
                >
                  <Download class="w-4 h-4 text-emerald-600 flex-shrink-0" />
                  <div>
                    <span class="font-semibold text-slate-800 block">Unduh Cadangan Database</span>
                    <span class="text-[10px] text-slate-400 block">Snapshot berkas .db SQLite</span>
                  </div>
                </button>

                <div class="my-1 border-t border-slate-100"></div>

                <!-- Reset Data Keuangan -->
                <button
                  type="button"
                  class="w-full text-left px-3.5 py-2.5 text-xs text-rose-600 hover:bg-rose-50 flex items-center space-x-2.5 transition font-medium cursor-pointer"
                  @click="openResetModal"
                >
                  <Trash2 class="w-4 h-4 text-rose-500 flex-shrink-0" />
                  <div>
                    <span class="font-semibold block">Reset Data Keuangan</span>
                    <span class="text-[10px] text-slate-400 block"
                      >Kosongkan transaksi & rekening</span
                    >
                  </div>
                </button>

                <div class="my-1 border-t border-slate-100"></div>

                <!-- Keluar / Logout -->
                <button
                  type="button"
                  class="w-full text-left px-3.5 py-2 text-xs text-slate-700 hover:bg-slate-50 flex items-center space-x-2.5 transition font-medium cursor-pointer"
                  @click="openLogoutModal"
                >
                  <LogOut class="w-4 h-4 text-slate-400 flex-shrink-0" />
                  <span>Keluar (Logout)</span>
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </header>

    <!-- Main Content Area -->
    <main
      class="flex-1 max-w-7xl w-full mx-auto px-4 sm:px-6 lg:px-8 py-5 sm:py-8 space-y-6 sm:space-y-8 pb-28 sm:pb-12"
    >
      <!-- TAB 1: DASHBOARD VIEW -->
      <template v-if="activeTab === 'dashboard'">
        <!-- Welcome & Mobile PWA Highlight Banner -->
        <div
          class="bg-gradient-to-r from-emerald-700 to-teal-800 rounded-2xl p-5 sm:p-7 text-white shadow-lg relative overflow-hidden"
        >
          <div class="relative z-10 max-w-2xl space-y-2">
            <div class="flex items-center space-x-2">
              <span
                class="inline-block px-2.5 py-0.5 rounded-full bg-emerald-600/60 text-emerald-100 text-xs font-medium tracking-wide uppercase"
              >
                PWA & Mobile Ready
              </span>
            </div>
            <h2 class="text-xl sm:text-3xl font-bold tracking-tight">Selamat Datang di Neraca</h2>
            <p class="text-emerald-100/90 text-xs sm:text-sm leading-relaxed">
              Aplikasi pendamping finansial pribadi berbasis prinsip 50-30-20 dan pemantauan batas
              belanja harian yang steril dari aset investasi Anda.
            </p>
          </div>
          <div class="absolute -right-8 -bottom-10 opacity-10 pointer-events-none">
            <PiggyBank class="w-72 h-72" />
          </div>
        </div>

        <!-- HIGHLIGHT: MESIN HITUNG PROSPECT DAILY LIMIT & COUNTDOWN GAJIAN -->
        <div class="space-y-4">
          <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-2">
            <div>
              <h2
                class="text-lg sm:text-xl font-bold tracking-tight text-slate-900 flex items-center space-x-2"
              >
                <Clock class="w-5 h-5 text-emerald-600" />
                <span>Siklus Gajian & Batas Belanja Harian</span>
              </h2>
              <p class="text-xs text-slate-500">
                Formula pembagian kas operasional riil hingga tanggal gajian berikutnya
              </p>
            </div>
            <div class="flex items-center space-x-2">
              <span
                class="text-xs px-2.5 py-1 rounded-full font-bold border"
                :class="[
                  burnRate.burn_rate_status === 'danger'
                    ? 'bg-rose-50 text-rose-700 border-rose-200'
                    : burnRate.burn_rate_status === 'warning'
                      ? 'bg-amber-50 text-amber-700 border-amber-200'
                      : 'bg-emerald-50 text-emerald-700 border-emerald-200',
                ]"
              >
                {{
                  burnRate.burn_rate_status === 'danger'
                    ? '⚠️ Status: Boros (> Ambang)'
                    : burnRate.burn_rate_status === 'warning'
                      ? '⚡ Status: Waspada'
                      : '✅ Status: Aman (< Ambang)'
                }}
              </span>
            </div>
          </div>

          <!-- 4 Kartu Metrik Inti Spreadsheet -->
          <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
            <!-- 1. Next Payday Remain -->
            <div class="bg-white p-5 rounded-2xl border border-slate-200 shadow-sm space-y-3">
              <div class="flex items-center justify-between">
                <span class="text-xs font-semibold uppercase tracking-wider text-slate-500">
                  Next Payday Remain
                </span>
                <div
                  class="w-8 h-8 rounded-lg bg-blue-50 text-blue-600 flex items-center justify-center"
                >
                  <Calendar class="w-4 h-4" />
                </div>
              </div>
              <div>
                <div class="text-2xl sm:text-3xl font-extrabold text-slate-900 tracking-tight">
                  {{ burnRate.next_payday_remain }}
                  <span class="text-sm font-semibold text-slate-500">Hari Lagi</span>
                </div>
                <p class="text-xs text-slate-500 mt-1">
                  Gajian: {{ formatIndonesianDate(burnRate.next_payday_date) }}
                </p>
              </div>
            </div>

            <!-- 2. Prospect Daily Limit -->
            <div
              class="bg-white p-5 rounded-2xl border-2 border-emerald-500/40 bg-gradient-to-br from-emerald-50/40 to-white shadow-sm space-y-3"
            >
              <div class="flex items-center justify-between">
                <span class="text-xs font-bold uppercase tracking-wider text-emerald-800">
                  Prospect Daily Limit
                </span>
                <div
                  class="w-8 h-8 rounded-lg bg-emerald-100 text-emerald-700 flex items-center justify-center"
                >
                  <ShieldAlert class="w-4 h-4" />
                </div>
              </div>
              <div>
                <div class="text-2xl sm:text-3xl font-extrabold text-emerald-700 tracking-tight">
                  {{ formatRupiah(burnRate.prospect_daily_limit) }}
                </div>
                <p class="text-xs text-emerald-800/80 mt-1 font-medium">
                  Ambang batas belanja / hari
                </p>
              </div>
            </div>

            <!-- 3. Average Daily Expense -->
            <div class="bg-white p-5 rounded-2xl border border-slate-200 shadow-sm space-y-3">
              <div class="flex items-center justify-between">
                <span class="text-xs font-semibold uppercase tracking-wider text-slate-500">
                  Average Daily Expense
                </span>
                <div
                  class="w-8 h-8 rounded-lg bg-amber-50 text-amber-600 flex items-center justify-center"
                >
                  <Flame class="w-4 h-4" />
                </div>
              </div>
              <div>
                <div
                  class="text-2xl sm:text-3xl font-extrabold tracking-tight"
                  :class="[
                    burnRate.burn_rate_status === 'danger'
                      ? 'text-rose-600'
                      : burnRate.burn_rate_status === 'warning'
                        ? 'text-amber-600'
                        : 'text-slate-900',
                  ]"
                >
                  {{ formatRupiah(burnRate.average_daily_expense) }}
                </div>
                <p class="text-xs text-slate-500 mt-1">Rata-rata riil belanja / hari</p>
              </div>
            </div>

            <!-- 4. Remain Funds -->
            <div class="bg-white p-5 rounded-2xl border border-slate-200 shadow-sm space-y-3">
              <div class="flex items-center justify-between">
                <span class="text-xs font-semibold uppercase tracking-wider text-slate-500">
                  Remain Funds
                </span>
                <div
                  class="w-8 h-8 rounded-lg bg-slate-100 text-slate-700 flex items-center justify-center"
                >
                  <Wallet class="w-4 h-4" />
                </div>
              </div>
              <div>
                <div class="text-2xl sm:text-3xl font-extrabold text-slate-900 tracking-tight">
                  {{ formatRupiah(burnRate.remain_funds) }}
                </div>
                <p class="text-xs text-slate-500 mt-1">Kas operasional aktif (steril)</p>
              </div>
            </div>
          </div>

          <!-- Bar Pembanding Laju Belanja (Burn Rate Gauge) -->
          <div
            class="bg-white p-4 sm:p-5 rounded-2xl border border-slate-200 shadow-sm space-y-2.5"
          >
            <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between text-xs gap-1">
              <span class="font-bold text-slate-800">
                Tingkat Pemakaian Ambang Harian: {{ burnRateUsagePercent }}%
              </span>
              <span class="text-slate-500">
                Rata-rata riil {{ formatRupiah(burnRate.average_daily_expense) }} / hari dari batas
                aman {{ formatRupiah(burnRate.prospect_daily_limit) }} / hari
              </span>
            </div>
            <div class="w-full bg-slate-100 h-3 rounded-full overflow-hidden flex">
              <div
                class="h-full rounded-full transition-all duration-500"
                :style="{ width: `${Math.min(burnRateUsagePercent, 100)}%` }"
                :class="[
                  burnRate.burn_rate_status === 'danger'
                    ? 'bg-rose-500'
                    : burnRate.burn_rate_status === 'warning'
                      ? 'bg-amber-500'
                      : 'bg-emerald-500',
                ]"
              ></div>
            </div>
          </div>
        </div>

        <!-- Breakdown Pilar 50-30-20 Siklus Berjalan -->
        <div class="bg-white p-5 rounded-2xl border border-slate-200 shadow-sm space-y-4">
          <div class="flex items-center justify-between">
            <h3 class="text-sm sm:text-base font-bold text-slate-900">
              Distribusi Pengeluaran Siklus Berjalan (50-30-20)
            </h3>
            <span class="text-xs font-bold text-slate-500">
              Grand Total: {{ formatRupiah(burnRate.grand_total_expenses) }}
            </span>
          </div>

          <div class="grid grid-cols-1 sm:grid-cols-3 gap-3">
            <!-- 50% Needs -->
            <div class="p-3.5 rounded-xl bg-blue-50/60 border border-blue-100 space-y-1">
              <div class="flex items-center justify-between text-xs text-blue-900 font-bold">
                <span>🔵 50% Needs (Pokok)</span>
                <span>{{ burnRate.pillar_breakdown?.needs?.percentage || 0 }}%</span>
              </div>
              <div class="text-base sm:text-lg font-extrabold text-blue-950">
                {{ formatRupiah(burnRate.pillar_breakdown?.needs?.total || 0) }}
              </div>
            </div>

            <!-- 30% Wants -->
            <div class="p-3.5 rounded-xl bg-purple-50/60 border border-purple-100 space-y-1">
              <div class="flex items-center justify-between text-xs text-purple-900 font-bold">
                <span>🟣 30% Wants (Gaya Hidup)</span>
                <span>{{ burnRate.pillar_breakdown?.wants?.percentage || 0 }}%</span>
              </div>
              <div class="text-base sm:text-lg font-extrabold text-purple-950">
                {{ formatRupiah(burnRate.pillar_breakdown?.wants?.total || 0) }}
              </div>
            </div>

            <!-- 20% Savings -->
            <div class="p-3.5 rounded-xl bg-emerald-50/60 border border-emerald-100 space-y-1">
              <div class="flex items-center justify-between text-xs text-emerald-900 font-bold">
                <span>🟢 20% Savings (Tabungan)</span>
                <span>{{ burnRate.pillar_breakdown?.savings?.percentage || 0 }}%</span>
              </div>
              <div class="text-base sm:text-lg font-extrabold text-emerald-950">
                {{ formatRupiah(burnRate.pillar_breakdown?.savings?.total || 0) }}
              </div>
            </div>
          </div>
        </div>

        <!-- Quick Metrics Overview -->
        <div class="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-4 sm:gap-5">
          <!-- Total Net Worth -->
          <div
            class="bg-white p-5 sm:p-6 rounded-2xl border border-slate-200 shadow-sm space-y-3 sm:space-y-4"
          >
            <div class="flex items-center justify-between">
              <span class="text-xs font-semibold uppercase tracking-wider text-slate-500"
                >Total Net Worth</span
              >
              <div
                class="w-9 h-9 rounded-lg bg-emerald-50 text-emerald-600 flex items-center justify-center"
              >
                <Wallet class="w-5 h-5" />
              </div>
            </div>
            <div>
              <div class="text-2xl sm:text-3xl font-extrabold text-slate-900 tracking-tight">
                {{ formatRupiah(netWorth) }}
              </div>
              <p class="text-xs text-slate-500 mt-1">Total seluruh kas & portofolio aset pasif</p>
            </div>
          </div>

          <!-- Pemasukan Bulan Ini -->
          <div
            class="bg-white p-5 sm:p-6 rounded-2xl border border-slate-200 shadow-sm space-y-3 sm:space-y-4"
          >
            <div class="flex items-center justify-between">
              <span class="text-xs font-semibold uppercase tracking-wider text-slate-500"
                >Pemasukan</span
              >
              <div
                class="w-9 h-9 rounded-lg bg-emerald-50 text-emerald-600 flex items-center justify-center"
              >
                <ArrowDownLeft class="w-5 h-5" />
              </div>
            </div>
            <div>
              <div class="text-2xl sm:text-3xl font-extrabold text-slate-900 tracking-tight">
                {{ formatRupiah(totalMonthlyIncome) }}
              </div>
              <p class="text-xs text-slate-500 mt-1">Total pemasukan bulan berjalan</p>
            </div>
          </div>

          <!-- Pengeluaran Bulan Ini -->
          <div
            class="bg-white p-5 sm:p-6 rounded-2xl border border-slate-200 shadow-sm space-y-3 sm:space-y-4 sm:col-span-2 md:col-span-1"
          >
            <div class="flex items-center justify-between">
              <span class="text-xs font-semibold uppercase tracking-wider text-slate-500"
                >Pengeluaran</span
              >
              <div
                class="w-9 h-9 rounded-lg bg-rose-50 text-rose-600 flex items-center justify-center"
              >
                <ArrowUpRight class="w-5 h-5" />
              </div>
            </div>
            <div>
              <div class="text-2xl sm:text-3xl font-extrabold text-slate-900 tracking-tight">
                {{ formatRupiah(totalMonthlyExpense) }}
              </div>
              <p class="text-xs text-slate-500 mt-1">Total pengeluaran bulan berjalan</p>
            </div>
          </div>
        </div>

        <!-- Detail Info Cards -->
        <div class="grid grid-cols-1 lg:grid-cols-2 gap-5 sm:gap-6">
          <!-- Status Stack -->
          <div class="bg-white p-5 sm:p-6 rounded-2xl border border-slate-200 shadow-sm space-y-4">
            <h3 class="text-base font-bold text-slate-900 flex items-center space-x-2">
              <Activity class="w-4 h-4 text-emerald-600" />
              <span>Konektivitas & Stack</span>
            </h3>
            <div class="space-y-3 text-sm">
              <div class="flex items-center justify-between py-2 border-b border-slate-100">
                <span class="text-slate-500">Backend Server</span>
                <span class="font-semibold text-slate-800">Golang (Chi Router v5)</span>
              </div>
              <div class="flex items-center justify-between py-2 border-b border-slate-100">
                <span class="text-slate-500">Database Engine</span>
                <span class="font-semibold text-slate-800">SQLite (Pure Go / WAL)</span>
              </div>
              <div class="flex items-center justify-between py-2 border-b border-slate-100">
                <span class="text-slate-500">Client Runtime</span>
                <span class="font-semibold text-slate-800">Vue 3 + Tailwind + PWA</span>
              </div>
              <div class="flex items-center justify-between py-2">
                <span class="text-slate-500">Status Endpoint</span>
                <span
                  v-if="health"
                  class="font-mono text-xs px-2 py-1 rounded bg-slate-100 text-slate-700"
                >
                  DB: {{ health.database }} | OK
                </span>
                <span v-else class="text-xs text-amber-600">Menghubungkan...</span>
              </div>
            </div>
          </div>

          <!-- Rencana Fitur Roadmap -->
          <div class="bg-white p-5 sm:p-6 rounded-2xl border border-slate-200 shadow-sm space-y-4">
            <h3 class="text-base font-bold text-slate-900 flex items-center space-x-2">
              <CheckCircle2 class="w-4 h-4 text-emerald-600" />
              <span>Roadmap Tahapan Fitur</span>
            </h3>
            <ul class="space-y-2.5 text-sm text-slate-600">
              <li class="flex items-start space-x-2">
                <span class="w-1.5 h-1.5 rounded-full bg-emerald-500 mt-2 flex-shrink-0"></span>
                <span
                  ><strong class="text-emerald-700">Tahap 1 (Selesai):</strong> Akun & Wealth
                  Management (Pemisahan Kas vs Aset Pasif).</span
                >
              </li>
              <li class="flex items-start space-x-2">
                <span class="w-1.5 h-1.5 rounded-full bg-emerald-500 mt-2 flex-shrink-0"></span>
                <span
                  ><strong class="text-emerald-700">Tahap 2 (Selesai):</strong> Kategori 50-30-20 &
                  Log Transaksi Harian (Pokok, Pribadi, Investasi).</span
                >
              </li>
              <li class="flex items-start space-x-2">
                <span class="w-1.5 h-1.5 rounded-full bg-emerald-500 mt-2 flex-shrink-0"></span>
                <span
                  ><strong class="text-emerald-700">Tahap 3 (Selesai):</strong> Mesin Hitung
                  Prospect Daily Limit & Countdown Gajian.</span
                >
              </li>
              <li class="flex items-start space-x-2">
                <span class="w-1.5 h-1.5 rounded-full bg-emerald-500 mt-2 flex-shrink-0"></span>
                <span
                  ><strong class="text-emerald-700">Tahap 4 (Selesai):</strong> Alokasi Anggaran
                  50-30-20 & Visualisasi Rekapitulasi Spreadsheet.</span
                >
              </li>
            </ul>
          </div>
        </div>
      </template>

      <!-- TAB 2: ACCOUNTS / DOMPET & WEALTH VIEW -->
      <template v-else-if="activeTab === 'accounts'">
        <AccountsView :format-rupiah="formatRupiah" @account-updated="refreshAll" />
      </template>

      <!-- TAB 3: TRANSAKSI (TAHAP 2) -->
      <template v-else-if="activeTab === 'transactions'">
        <TransactionsView :format-rupiah="formatRupiah" @transaction-changed="refreshAll" />
      </template>

      <!-- TAB 4: ANGGARAN 50-30-20 (TAHAP 4) -->
      <template v-else-if="activeTab === 'budgets'">
        <BudgetsView :burn-rate="burnRate" :format-rupiah="formatRupiah" @refresh="refreshAll" />
      </template>
    </main>

    <!-- Mobile Bottom Navigation Bar (Khusus Layar HP) -->
    <nav
      aria-label="Navigasi Bawah Mobile"
      class="md:hidden fixed bottom-0 left-0 right-0 z-40 bg-white/95 backdrop-blur-md border-t border-slate-200 pb-[env(safe-area-inset-bottom)] shadow-lg"
    >
      <div class="grid grid-cols-4 h-16">
        <button
          v-for="tab in navTabs"
          :key="tab.id"
          type="button"
          :class="[
            activeTab === tab.id
              ? 'text-emerald-600 font-semibold'
              : 'text-slate-500 hover:text-slate-700 font-medium',
          ]"
          class="flex flex-col items-center justify-center space-y-1 transition text-[11px]"
          @click="switchTab(tab.id)"
        >
          <component
            :is="tab.icon"
            class="w-5 h-5 transition-transform"
            :class="{ 'scale-110': activeTab === tab.id }"
          />
          <span>{{ tab.label }}</span>
        </button>
      </div>
    </nav>

    <!-- Desktop Footer -->
    <footer
      class="bg-white border-t border-slate-200 py-4 text-center text-xs text-slate-500 hidden sm:block"
    >
      Neraca &copy; {{ new Date().getFullYear() }} — Ultra-lightweight Personal Financial Dashboard
      (PWA)
    </footer>

    <!-- Modal Konfirmasi Logout -->
    <dialog
      v-if="showLogoutModal"
      open
      class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/50 backdrop-blur-sm w-full h-full max-w-none max-h-none border-0 m-0 bg-transparent text-slate-800"
      aria-labelledby="logout-title"
    >
      <div
        class="bg-white rounded-2xl shadow-2xl max-w-sm w-full p-5 space-y-4 border border-slate-100 animate-in fade-in zoom-in-95 duration-150"
      >
        <div
          class="w-11 h-11 rounded-xl bg-slate-100 text-slate-700 flex items-center justify-center"
        >
          <LogOut class="w-5 h-5" />
        </div>
        <div>
          <h3 id="logout-title" class="text-base font-bold text-slate-900">Konfirmasi Keluar</h3>
          <p class="text-xs text-slate-500 mt-1">
            Apakah Anda yakin ingin keluar dari Neraca? Sesi login Anda di perangkat ini akan
            diakhiri.
          </p>
        </div>
        <div class="flex items-center justify-end space-x-2.5 pt-2">
          <button
            type="button"
            class="px-4 py-2 rounded-xl text-xs font-semibold text-slate-600 hover:bg-slate-100 transition min-h-[40px] cursor-pointer"
            @click="showLogoutModal = false"
          >
            Batal
          </button>
          <button
            type="button"
            class="px-4 py-2 rounded-xl text-xs font-bold text-white bg-slate-900 hover:bg-slate-800 transition min-h-[40px] cursor-pointer"
            @click="confirmLogout"
          >
            Ya, Keluar
          </button>
        </div>
      </div>
    </dialog>

    <!-- Modal Konfirmasi Reset Data Keuangan -->
    <dialog
      v-if="showResetModal"
      open
      class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/50 backdrop-blur-sm w-full h-full max-w-none max-h-none border-0 m-0 bg-transparent text-slate-800"
      aria-labelledby="reset-title"
    >
      <div
        class="bg-white rounded-2xl shadow-2xl max-w-md w-full p-5 space-y-4 border border-rose-100 animate-in fade-in zoom-in-95 duration-150"
      >
        <div class="w-11 h-11 rounded-xl bg-rose-50 text-rose-600 flex items-center justify-center">
          <AlertTriangle class="w-5 h-5" />
        </div>
        <div class="space-y-1.5">
          <h3 id="reset-title" class="text-base font-bold text-slate-900">
            Reset Seluruh Data Keuangan?
          </h3>
          <p class="text-xs text-slate-600 leading-relaxed">
            Tindakan ini akan
            <strong
              >menghapus permanen seluruh rekening, transaksi, dan histori penilaian aset</strong
            >. Kategori akan dikembalikan ke format default 50-30-20.
          </p>
          <div
            class="bg-amber-50 border border-amber-200 rounded-xl p-3 text-[11px] text-amber-800 space-y-1"
          >
            <span class="font-bold block">🛡️ Akun Login Anda Tetap Aman</span>
            <span
              >Username dan sesi login Anda <strong>tidak akan terhapus</strong>. Anda akan tetap
              login ke dashboard yang bersih dari data dummy.</span
            >
          </div>
        </div>
        <div class="flex items-center justify-end space-x-2.5 pt-2">
          <button
            type="button"
            :disabled="isResetting"
            class="px-4 py-2 rounded-xl text-xs font-semibold text-slate-600 hover:bg-slate-100 transition min-h-[40px] cursor-pointer"
            @click="showResetModal = false"
          >
            Batal
          </button>
          <button
            type="button"
            :disabled="isResetting"
            class="px-4 py-2 rounded-xl text-xs font-bold text-white bg-rose-600 hover:bg-rose-700 disabled:opacity-50 transition min-h-[40px] flex items-center space-x-1.5 cursor-pointer shadow-sm shadow-rose-200"
            @click="confirmResetData"
          >
            <RefreshCw v-if="isResetting" class="w-3.5 h-3.5 animate-spin" />
            <span>{{ isResetting ? 'Mereset Data...' : 'Ya, Hapus & Reset Data' }}</span>
          </button>
        </div>
      </div>
    </dialog>
  </div>
</template>
