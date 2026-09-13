<script setup>
import { ref, onMounted } from 'vue'
import {
  Wallet,
  ArrowDownLeft,
  ArrowUpRight,
  Activity,
  CheckCircle2,
  AlertCircle,
  PiggyBank,
  RefreshCw,
} from '@lucide/vue'

const health = ref(null)
const loading = ref(true)
const error = ref(null)

const checkHealth = async () => {
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

onMounted(() => {
  checkHealth()
})

// Format Rupiah helper
const formatRupiah = val => {
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    minimumFractionDigits: 0,
    maximumFractionDigits: 0,
  }).format(val)
}
</script>

<template>
  <div class="min-h-screen bg-slate-50 flex flex-col text-slate-800">
    <!-- Navbar -->
    <header class="bg-white border-b border-slate-200 sticky top-0 z-30 shadow-sm">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 h-16 flex items-center justify-between">
        <div class="flex items-center space-x-3">
          <div
            class="w-10 h-10 rounded-xl bg-emerald-600 flex items-center justify-center text-white shadow-md shadow-emerald-200"
          >
            <PiggyBank class="w-6 h-6" />
          </div>
          <div>
            <h1 class="text-xl font-bold tracking-tight text-slate-900 leading-tight">Neraca</h1>
            <p class="text-xs text-slate-500 font-medium">Personal Financial Dashboard</p>
          </div>
        </div>

        <!-- Health / Server Status Badge -->
        <div class="flex items-center space-x-3">
          <div
            v-if="health && health.status === 'ok'"
            class="inline-flex items-center space-x-1.5 px-3 py-1 rounded-full text-xs font-semibold bg-emerald-50 text-emerald-700 border border-emerald-200"
          >
            <span class="w-2 h-2 rounded-full bg-emerald-500 animate-pulse"></span>
            <span>Server Online (SQLite WAL)</span>
          </div>

          <div
            v-else-if="error"
            class="inline-flex items-center space-x-1.5 px-3 py-1 rounded-full text-xs font-semibold bg-rose-50 text-rose-700 border border-rose-200"
          >
            <AlertCircle class="w-3.5 h-3.5" />
            <span>Terputus: {{ error }}</span>
          </div>

          <button
            :disabled="loading"
            title="Cek Status Server"
            class="p-2 rounded-lg text-slate-500 hover:text-slate-800 hover:bg-slate-100 transition"
            @click="checkHealth"
          >
            <RefreshCw class="w-4 h-4" :class="{ 'animate-spin': loading }" />
          </button>
        </div>
      </div>
    </header>

    <!-- Main Content -->
    <main class="flex-1 max-w-7xl w-full mx-auto px-4 sm:px-6 lg:px-8 py-8 space-y-8">
      <!-- Welcome & Action Banner -->
      <div
        class="bg-gradient-to-r from-emerald-700 to-teal-800 rounded-2xl p-6 sm:p-8 text-white shadow-lg relative overflow-hidden"
      >
        <div class="relative z-10 max-w-2xl space-y-2">
          <span
            class="inline-block px-3 py-0.5 rounded-full bg-emerald-600/60 text-emerald-100 text-xs font-medium tracking-wide uppercase"
          >
            Pondasi Siap
          </span>
          <h2 class="text-2xl sm:text-3xl font-bold tracking-tight">Selamat Datang di Neraca</h2>
          <p class="text-emerald-100/90 text-sm sm:text-base leading-relaxed">
            Struktur sistem Go + SQLite + Vue 3 berhasil diinisialisasi. Server backend dan database
            SQLite siap untuk tahap pengembangan fitur finansial Anda.
          </p>
        </div>
        <div class="absolute -right-8 -bottom-10 opacity-10 pointer-events-none">
          <PiggyBank class="w-72 h-72" />
        </div>
      </div>

      <!-- Quick Metrics Overview (Skeleton Preview) -->
      <div class="grid grid-cols-1 md:grid-cols-3 gap-5">
        <!-- Total Net Worth -->
        <div class="bg-white p-6 rounded-2xl border border-slate-200 shadow-sm space-y-4">
          <div class="flex items-center justify-between">
            <span class="text-xs font-semibold uppercase tracking-wider text-slate-500"
              >Total Saldo (Net Worth)</span
            >
            <div
              class="w-9 h-9 rounded-lg bg-emerald-50 text-emerald-600 flex items-center justify-center"
            >
              <Wallet class="w-5 h-5" />
            </div>
          </div>
          <div>
            <div class="text-3xl font-extrabold text-slate-900 tracking-tight">
              {{ formatRupiah(0) }}
            </div>
            <p class="text-xs text-slate-500 mt-1">Akumulasi dari seluruh kantong/rekening</p>
          </div>
        </div>

        <!-- Pemasukan Bulan Ini -->
        <div class="bg-white p-6 rounded-2xl border border-slate-200 shadow-sm space-y-4">
          <div class="flex items-center justify-between">
            <span class="text-xs font-semibold uppercase tracking-wider text-slate-500"
              >Pemasukan Bulan Ini</span
            >
            <div
              class="w-9 h-9 rounded-lg bg-blue-50 text-blue-600 flex items-center justify-center"
            >
              <ArrowDownLeft class="w-5 h-5" />
            </div>
          </div>
          <div>
            <div class="text-3xl font-extrabold text-slate-900 tracking-tight">
              {{ formatRupiah(0) }}
            </div>
            <p class="text-xs text-slate-500 mt-1">Belum ada transaksi pemasukan</p>
          </div>
        </div>

        <!-- Pengeluaran Bulan Ini -->
        <div class="bg-white p-6 rounded-2xl border border-slate-200 shadow-sm space-y-4">
          <div class="flex items-center justify-between">
            <span class="text-xs font-semibold uppercase tracking-wider text-slate-500"
              >Pengeluaran Bulan Ini</span
            >
            <div
              class="w-9 h-9 rounded-lg bg-rose-50 text-rose-600 flex items-center justify-center"
            >
              <ArrowUpRight class="w-5 h-5" />
            </div>
          </div>
          <div>
            <div class="text-3xl font-extrabold text-slate-900 tracking-tight">
              {{ formatRupiah(0) }}
            </div>
            <p class="text-xs text-slate-500 mt-1">Belum ada transaksi pengeluaran</p>
          </div>
        </div>
      </div>

      <!-- Detail System Health & Next Development Steps -->
      <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <!-- Status Panel -->
        <div class="bg-white p-6 rounded-2xl border border-slate-200 shadow-sm space-y-4">
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
              <span class="text-slate-500">Database Driver</span>
              <span class="font-semibold text-slate-800">SQLite (Pure Go / WAL Mode)</span>
            </div>
            <div class="flex items-center justify-between py-2 border-b border-slate-100">
              <span class="text-slate-500">Frontend Client</span>
              <span class="font-semibold text-slate-800">Vue 3 + Tailwind CSS + Vite</span>
            </div>
            <div class="flex items-center justify-between py-2">
              <span class="text-slate-500">Status API Endpoint</span>
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

        <!-- Rencana Fitur Selanjutnya -->
        <div class="bg-white p-6 rounded-2xl border border-slate-200 shadow-sm space-y-4">
          <h3 class="text-base font-bold text-slate-900 flex items-center space-x-2">
            <CheckCircle2 class="w-4 h-4 text-emerald-600" />
            <span>Siap untuk Pengembangan Fitur</span>
          </h3>
          <ul class="space-y-2.5 text-sm text-slate-600">
            <li class="flex items-start space-x-2">
              <span class="w-1.5 h-1.5 rounded-full bg-emerald-500 mt-2 flex-shrink-0"></span>
              <span
                ><strong>Akun & Dompet:</strong> Manajemen rekening (BCA, Mandiri, Jago, GoPay,
                Tunai, dsb).</span
              >
            </li>
            <li class="flex items-start space-x-2">
              <span class="w-1.5 h-1.5 rounded-full bg-emerald-500 mt-2 flex-shrink-0"></span>
              <span
                ><strong>Transaksi:</strong> Form catat pemasukan, pengeluaran, dan transfer antar
                akun.</span
              >
            </li>
            <li class="flex items-start space-x-2">
              <span class="w-1.5 h-1.5 rounded-full bg-emerald-500 mt-2 flex-shrink-0"></span>
              <span
                ><strong>Kategori & Budget:</strong> Pengelompokan pos pengeluaran & monitoring
                batas bulanan.</span
              >
            </li>
            <li class="flex items-start space-x-2">
              <span class="w-1.5 h-1.5 rounded-full bg-emerald-500 mt-2 flex-shrink-0"></span>
              <span><strong>Visualisasi & Laporan:</strong> Grafik cashflow dan tren bulanan.</span>
            </li>
          </ul>
        </div>
      </div>
    </main>

    <!-- Footer -->
    <footer class="bg-white border-t border-slate-200 py-4 text-center text-xs text-slate-400">
      Neraca &copy; {{ new Date().getFullYear() }} — Ultra-lightweight Personal Financial Dashboard
    </footer>
  </div>
</template>
