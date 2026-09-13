<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
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
} from '@lucide/vue'
import AccountsView from './views/AccountsView.vue'
import TransactionsView from './views/TransactionsView.vue'

const health = ref(null)
const loading = ref(true)
const error = ref(null)
const isOnline = ref(typeof navigator !== 'undefined' ? navigator.onLine : true)
const activeTab = ref('dashboard')
const netWorth = ref(0)
const totalMonthlyExpense = ref(0)
const totalMonthlyIncome = ref(0)

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

const fetchNetWorth = async () => {
  try {
    const res = await fetch('/api/accounts')
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
    if (res.ok) {
      const data = await res.json()
      totalMonthlyExpense.value = data.total_expense || 0
      totalMonthlyIncome.value = data.total_income || 0
    }
  } catch (err) {
    console.error(err)
  }
}

const switchTab = tabId => {
  activeTab.value = tabId
  refreshAll()
}

const refreshAll = () => {
  checkHealth()
  fetchNetWorth()
  fetchMonthlyTransactions()
}

onMounted(() => {
  window.addEventListener('online', updateOnlineStatus)
  window.addEventListener('offline', updateOnlineStatus)
  refreshAll()
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
  <div class="min-h-screen bg-slate-50 flex flex-col text-slate-800">
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
          class="bg-gradient-to-r from-emerald-700 to-teal-800 rounded-2xl p-5 sm:p-8 text-white shadow-lg relative overflow-hidden"
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
            <p class="text-emerald-100/90 text-xs sm:text-base leading-relaxed">
              Aplikasi pendamping finansial pribadi berbasis prinsip 50-30-20 dan pemantauan batas
              belanja harian yang steril dari aset investasi Anda.
            </p>
          </div>
          <div class="absolute -right-8 -bottom-10 opacity-10 pointer-events-none">
            <PiggyBank class="w-72 h-72" />
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
                <span class="w-1.5 h-1.5 rounded-full bg-slate-300 mt-2 flex-shrink-0"></span>
                <span
                  ><strong>Tahap 3:</strong> Mesin Hitung Prospect Daily Limit & Countdown
                  Gajian.</span
                >
              </li>
              <li class="flex items-start space-x-2">
                <span class="w-1.5 h-1.5 rounded-full bg-slate-300 mt-2 flex-shrink-0"></span>
                <span><strong>Tahap 4:</strong> Visualisasi Dashboard Spreadsheet & Rekap.</span>
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

      <!-- TAB 4: ANGGARAN 50-30-20 (TAHAP 3/4 PLACEHOLDER) -->
      <template v-else-if="activeTab === 'budgets'">
        <div class="bg-white rounded-2xl p-8 border border-slate-200 text-center space-y-3">
          <div
            class="w-12 h-12 rounded-2xl bg-blue-50 text-blue-600 flex items-center justify-center mx-auto"
          >
            <PieChart class="w-6 h-6" />
          </div>
          <h3 class="font-bold text-slate-800 text-lg">Modul Anggaran 50-30-20 (Tahap 3 & 4)</h3>
          <p class="text-xs text-slate-500 max-w-md mx-auto">
            Pembagian pilar anggaran 50% Pokok, 30% Pribadi, dan 20% Investasi akan
            diimplementasikan pada tahap berikutnya.
          </p>
        </div>
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
  </div>
</template>
