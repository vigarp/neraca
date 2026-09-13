<script setup>
import { ref, computed, onMounted } from 'vue'
import {
  PieChart,
  AlertTriangle,
  CheckCircle2,
  Calendar,
  Layers,
  Sparkles,
  Coins,
  ReceiptText,
} from '@lucide/vue'

const props = defineProps({
  burnRate: {
    type: Object,
    default: () => ({
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
    }),
  },
  formatRupiah: {
    type: Function,
    default: val => `Rp ${Number(val || 0).toLocaleString('id-ID')}`,
  },
})

const emit = defineEmits(['refresh'])

// Custom Income Simulation
const isCustomIncome = ref(false)
const simulatedIncome = ref(0)
const STORAGE_KEY = 'neraca_simulated_income'

onMounted(() => {
  const saved = localStorage.getItem(STORAGE_KEY)
  if (saved) {
    const val = Number.parseFloat(saved)
    if (!Number.isNaN(val) && val > 0) {
      simulatedIncome.value = val
      isCustomIncome.value = true
    }
  }
})

const effectiveIncome = computed(() => {
  if (isCustomIncome.value && simulatedIncome.value > 0) {
    return simulatedIncome.value
  }
  if (props.burnRate.cycle_income && props.burnRate.cycle_income > 0) {
    return props.burnRate.cycle_income
  }
  // Fallback ke total pengeluaran + remain funds jika belum mencatat pemasukan
  return (props.burnRate.grand_total_expenses || 0) + (props.burnRate.remain_funds || 0)
})

const saveCustomIncome = () => {
  if (simulatedIncome.value > 0) {
    localStorage.setItem(STORAGE_KEY, simulatedIncome.value.toString())
    isCustomIncome.value = true
  } else {
    localStorage.removeItem(STORAGE_KEY)
    isCustomIncome.value = false
  }
}

const resetCustomIncome = () => {
  simulatedIncome.value = 0
  isCustomIncome.value = false
  localStorage.removeItem(STORAGE_KEY)
}

// Target Budget 50-30-20
const targetNeeds = computed(() => effectiveIncome.value * 0.5)
const targetWants = computed(() => effectiveIncome.value * 0.3)
const targetSavings = computed(() => effectiveIncome.value * 0.2)

// Realisasi per Pilar
const actualNeeds = computed(() => props.burnRate.pillar_breakdown?.needs?.total || 0)
const actualWants = computed(() => props.burnRate.pillar_breakdown?.wants?.total || 0)
const actualSavings = computed(() => props.burnRate.pillar_breakdown?.savings?.total || 0)

// Sisa Dana per Pilar
const remainNeeds = computed(() => targetNeeds.value - actualNeeds.value)
const remainWants = computed(() => targetWants.value - actualWants.value)
const remainSavings = computed(() => targetSavings.value - actualSavings.value)

// Total Sisa Dana Anggaran (Termasuk sisa tabungan jika belum dialokasikan)
const totalRemainBudget = computed(() => {
  return (
    Math.max(0, remainNeeds.value) +
    Math.max(0, remainWants.value) +
    Math.max(0, remainSavings.value)
  )
})

// Persentase Serapan per Pilar terhadap Target Ideal
const percentNeeds = computed(() => {
  if (targetNeeds.value <= 0) return 0
  return Math.round((actualNeeds.value / targetNeeds.value) * 100)
})
const percentWants = computed(() => {
  if (targetWants.value <= 0) return 0
  return Math.round((actualWants.value / targetWants.value) * 100)
})
const percentSavings = computed(() => {
  if (targetSavings.value <= 0) return 0
  return Math.round((actualSavings.value / targetSavings.value) * 100)
})

// Filter kategori per pilar untuk 3 kolom ala spreadsheet
const needsCategories = computed(() => {
  return (props.burnRate.category_breakdown || []).filter(c => c.pillar === 'needs')
})
const wantsCategories = computed(() => {
  return (props.burnRate.category_breakdown || []).filter(c => c.pillar === 'wants')
})
const savingsCategories = computed(() => {
  return (props.burnRate.category_breakdown || []).filter(c => c.pillar === 'savings')
})

// Total kuantitas transaksi di siklus ini
const totalTransactionCount = computed(() => {
  return (props.burnRate.category_breakdown || []).reduce(
    (acc, curr) => acc + (curr.transaction_count || 0),
    0
  )
})

// Format tanggal Indonesia
const formatIndonesianDate = dateStr => {
  if (!dateStr) return '-'
  const date = new Date(dateStr + 'T00:00:00')
  return date.toLocaleDateString('id-ID', {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
  })
}

// SVG Donut Chart Calculation
const chartCircumference = 2 * Math.PI * 40 // radius = 40 => circumference ~251.327
const grandTotalExpenses = computed(() => props.burnRate.grand_total_expenses || 0)

const chartSegments = computed(() => {
  const total = grandTotalExpenses.value
  if (total <= 0) return []

  const pNeeds = (actualNeeds.value / total) * chartCircumference
  const pWants = (actualWants.value / total) * chartCircumference
  const pSavings = (actualSavings.value / total) * chartCircumference

  return [
    {
      name: 'Needs (50%)',
      stroke: '#0284c7', // sky-600
      dasharray: `${pNeeds} ${chartCircumference - pNeeds}`,
      dashoffset: 0,
      total: actualNeeds.value,
      percentage: props.burnRate.pillar_breakdown?.needs?.percentage || 0,
    },
    {
      name: 'Wants (30%)',
      stroke: '#9333ea', // purple-600
      dasharray: `${pWants} ${chartCircumference - pWants}`,
      dashoffset: -pNeeds,
      total: actualWants.value,
      percentage: props.burnRate.pillar_breakdown?.wants?.percentage || 0,
    },
    {
      name: 'Savings (20%)',
      stroke: '#059669', // emerald-600
      dasharray: `${pSavings} ${chartCircumference - pSavings}`,
      dashoffset: -(pNeeds + pWants),
      total: actualSavings.value,
      percentage: props.burnRate.pillar_breakdown?.savings?.percentage || 0,
    },
  ]
})
</script>

<template>
  <div class="space-y-6 sm:space-y-8">
    <!-- Header & Info Siklus Gajian -->
    <div
      class="bg-white p-5 sm:p-6 rounded-2xl border border-slate-200 shadow-sm flex flex-col md:flex-row md:items-center md:justify-between gap-4"
    >
      <div class="space-y-1">
        <div class="flex items-center space-x-2">
          <span
            class="px-2.5 py-0.5 rounded-full text-xs font-bold uppercase tracking-wider bg-emerald-50 text-emerald-700 border border-emerald-200"
          >
            Siklus Berjalan
          </span>
          <span class="text-xs text-slate-500 font-medium">
            {{ formatIndonesianDate(burnRate.cycle_start_date) }} —
            {{ formatIndonesianDate(burnRate.next_payday_date) }}
          </span>
        </div>
        <h2 class="text-xl sm:text-2xl font-extrabold text-slate-900 tracking-tight">
          Alokasi Anggaran 50-30-20
        </h2>
        <p class="text-xs sm:text-sm text-slate-500">
          Evaluasi realisasi belanja terhadap prinsip 50% Kebutuhan Pokok, 30% Kebutuhan Pribadi,
          dan 20% Tabungan/Investasi.
        </p>
      </div>

      <div class="flex items-center space-x-3">
        <div class="text-right">
          <div class="text-xs text-slate-500 font-medium">Hitung Mundur Gajian</div>
          <div class="text-lg font-bold text-slate-900 flex items-center justify-end space-x-1.5">
            <Calendar class="w-4 h-4 text-emerald-600" />
            <span>{{ burnRate.next_payday_remain }} Hari Lagi</span>
          </div>
        </div>
        <button
          type="button"
          aria-label="Muat ulang data anggaran"
          class="p-2.5 rounded-xl border border-slate-200 hover:bg-slate-50 text-slate-600 transition min-w-[44px] min-h-[44px] flex items-center justify-center"
          @click="emit('refresh')"
        >
          <Sparkles class="w-4 h-4" />
        </button>
      </div>
    </div>

    <!-- Basis Pemasukan & Target Budget Ideal -->
    <div class="bg-white p-5 sm:p-6 rounded-2xl border border-slate-200 shadow-sm space-y-5">
      <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3">
        <div>
          <div class="flex items-center space-x-2">
            <Coins class="w-5 h-5 text-emerald-600" />
            <h3 class="text-base sm:text-lg font-bold text-slate-900">
              Basis Penghasilan & Formula 50/30/20
            </h3>
          </div>
          <p class="text-xs text-slate-500 mt-0.5">
            Plafon ideal dihitung otomatis dari total pemasukan siklus ini atau simulasi mandiri.
          </p>
        </div>

        <!-- Tombol / Form Simulasi Penghasilan -->
        <div class="flex items-center space-x-2">
          <div v-if="!isCustomIncome" class="flex items-center space-x-2">
            <div class="text-right">
              <span class="text-xs text-slate-500">Basis Digunakan:</span>
              <div class="text-sm font-bold text-slate-800">
                {{ formatRupiah(effectiveIncome) }}
              </div>
            </div>
            <button
              type="button"
              class="text-xs px-3 py-2 rounded-xl bg-slate-100 hover:bg-slate-200 text-slate-700 font-medium transition min-h-[44px] flex items-center"
              @click="isCustomIncome = true"
            >
              Ubah Basis
            </button>
          </div>

          <div v-else class="flex items-center space-x-2">
            <div>
              <label for="input-custom-income" class="sr-only">Nominal Penghasilan Simulasi</label>
              <input
                id="input-custom-income"
                v-model.number="simulatedIncome"
                type="number"
                min="0"
                step="50000"
                placeholder="Contoh: 6000000"
                class="px-3 py-2 text-xs rounded-xl border border-slate-300 focus:outline-none focus:ring-2 focus:ring-emerald-500 w-36 sm:w-44 min-h-[44px]"
              />
            </div>
            <button
              type="button"
              class="px-3 py-2 text-xs rounded-xl bg-emerald-600 hover:bg-emerald-700 text-white font-semibold transition min-h-[44px] flex items-center"
              @click="saveCustomIncome"
            >
              Simpan
            </button>
            <button
              type="button"
              class="px-2.5 py-2 text-xs rounded-xl border border-slate-200 hover:bg-slate-100 text-slate-600 transition min-h-[44px] flex items-center"
              @click="resetCustomIncome"
            >
              Reset
            </button>
          </div>
        </div>
      </div>

      <!-- 3 Kartu Plafon Ideal 50-30-20 -->
      <div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
        <!-- 50% Needs -->
        <div
          class="p-4 sm:p-5 rounded-2xl bg-sky-50/60 border border-sky-100 space-y-3 relative overflow-hidden"
        >
          <div class="flex items-center justify-between">
            <span class="text-xs font-bold uppercase tracking-wider text-sky-800">
              50% Kebutuhan Pokok
            </span>
            <span class="text-xs font-bold text-sky-700 bg-sky-100 px-2 py-0.5 rounded-full">
              Needs
            </span>
          </div>
          <div>
            <div class="text-xl sm:text-2xl font-extrabold text-sky-950 tracking-tight">
              {{ formatRupiah(targetNeeds) }}
            </div>
            <p class="text-xs text-sky-700 mt-1">
              Realisasi: <strong>{{ formatRupiah(actualNeeds) }}</strong> ({{ percentNeeds }}%)
            </p>
          </div>
          <div class="w-full bg-sky-200/60 h-2 rounded-full overflow-hidden">
            <div
              class="h-full rounded-full transition-all duration-500"
              :class="percentNeeds > 100 ? 'bg-rose-500' : 'bg-sky-600'"
              :style="{ width: `${Math.min(percentNeeds, 100)}%` }"
            ></div>
          </div>
        </div>

        <!-- 30% Wants -->
        <div
          class="p-4 sm:p-5 rounded-2xl bg-purple-50/60 border border-purple-100 space-y-3 relative overflow-hidden"
        >
          <div class="flex items-center justify-between">
            <span class="text-xs font-bold uppercase tracking-wider text-purple-800">
              30% Kebutuhan Pribadi
            </span>
            <span class="text-xs font-bold text-purple-700 bg-purple-100 px-2 py-0.5 rounded-full">
              Wants
            </span>
          </div>
          <div>
            <div class="text-xl sm:text-2xl font-extrabold text-purple-950 tracking-tight">
              {{ formatRupiah(targetWants) }}
            </div>
            <p class="text-xs text-purple-700 mt-1">
              Realisasi: <strong>{{ formatRupiah(actualWants) }}</strong> ({{ percentWants }}%)
            </p>
          </div>
          <div class="w-full bg-purple-200/60 h-2 rounded-full overflow-hidden">
            <div
              class="h-full rounded-full transition-all duration-500"
              :class="percentWants > 100 ? 'bg-rose-500' : 'bg-purple-600'"
              :style="{ width: `${Math.min(percentWants, 100)}%` }"
            ></div>
          </div>
        </div>

        <!-- 20% Savings -->
        <div
          class="p-4 sm:p-5 rounded-2xl bg-emerald-50/60 border border-emerald-100 space-y-3 relative overflow-hidden"
        >
          <div class="flex items-center justify-between">
            <span class="text-xs font-bold uppercase tracking-wider text-emerald-800">
              20% Tabungan & Investasi
            </span>
            <span
              class="text-xs font-bold text-emerald-700 bg-emerald-100 px-2 py-0.5 rounded-full"
            >
              Savings
            </span>
          </div>
          <div>
            <div class="text-xl sm:text-2xl font-extrabold text-emerald-950 tracking-tight">
              {{ formatRupiah(targetSavings) }}
            </div>
            <p class="text-xs text-emerald-700 mt-1">
              Tercapai: <strong>{{ formatRupiah(actualSavings) }}</strong> ({{ percentSavings }}%)
            </p>
          </div>
          <div class="w-full bg-emerald-200/60 h-2 rounded-full overflow-hidden">
            <div
              class="h-full rounded-full transition-all duration-500 bg-emerald-600"
              :style="{ width: `${Math.min(percentSavings, 100)}%` }"
            ></div>
          </div>
        </div>
      </div>
    </div>

    <!-- Visualisasi Grafik Alokasi vs Realisasi (Pure SVG Donut Chart) -->
    <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
      <div
        class="bg-white p-5 sm:p-6 rounded-2xl border border-slate-200 shadow-sm flex flex-col items-center justify-center space-y-4"
      >
        <div class="text-center">
          <h3 class="text-sm font-bold text-slate-900">Distribusi Realisasi Pilar</h3>
          <p class="text-xs text-slate-500">Proporsi riil pengeluaran siklus ini</p>
        </div>

        <!-- SVG Donut -->
        <div class="relative w-44 h-44 flex items-center justify-center">
          <svg
            viewBox="0 0 100 100"
            class="w-full h-full -rotate-90 transform"
            role="img"
            aria-label="Grafik Donut Alokasi Pengeluaran 50-30-20"
          >
            <title>Grafik Alokasi 50-30-20</title>
            <!-- Background circle -->
            <circle
              cx="50"
              cy="50"
              r="40"
              stroke="#f1f5f9"
              stroke-width="12"
              fill="transparent"
            ></circle>
            <!-- Segments -->
            <circle
              v-for="seg in chartSegments"
              :key="seg.name"
              cx="50"
              cy="50"
              r="40"
              :stroke="seg.stroke"
              stroke-width="12"
              fill="transparent"
              :stroke-dasharray="seg.dasharray"
              :stroke-dashoffset="seg.dashoffset"
              class="transition-all duration-700 ease-out"
            ></circle>
          </svg>
          <div
            class="absolute inset-0 flex flex-col items-center justify-center pointer-events-none"
          >
            <span class="text-xs font-semibold text-slate-400 uppercase tracking-wider">Total</span>
            <span class="text-sm sm:text-base font-extrabold text-slate-900 px-2 text-center">
              {{ formatRupiah(grandTotalExpenses) }}
            </span>
          </div>
        </div>

        <!-- Legend -->
        <div class="w-full space-y-2 pt-2 border-t border-slate-100 text-xs">
          <div class="flex items-center justify-between">
            <div class="flex items-center space-x-2">
              <span class="w-3 h-3 rounded-full bg-sky-600 flex-shrink-0"></span>
              <span class="text-slate-600 font-medium">Pokok (Needs)</span>
            </div>
            <span class="font-bold text-slate-900">
              {{ burnRate.pillar_breakdown?.needs?.percentage || 0 }}%
            </span>
          </div>
          <div class="flex items-center justify-between">
            <div class="flex items-center space-x-2">
              <span class="w-3 h-3 rounded-full bg-purple-600 flex-shrink-0"></span>
              <span class="text-slate-600 font-medium">Pribadi (Wants)</span>
            </div>
            <span class="font-bold text-slate-900">
              {{ burnRate.pillar_breakdown?.wants?.percentage || 0 }}%
            </span>
          </div>
          <div class="flex items-center justify-between">
            <div class="flex items-center space-x-2">
              <span class="w-3 h-3 rounded-full bg-emerald-600 flex-shrink-0"></span>
              <span class="text-slate-600 font-medium">Tabungan (Savings)</span>
            </div>
            <span class="font-bold text-slate-900">
              {{ burnRate.pillar_breakdown?.savings?.percentage || 0 }}%
            </span>
          </div>
        </div>
      </div>

      <!-- Panduan & Analisis Komparasi Target Ideal -->
      <div
        class="lg:col-span-2 bg-white p-5 sm:p-6 rounded-2xl border border-slate-200 shadow-sm flex flex-col justify-between space-y-4"
      >
        <div class="space-y-3">
          <div class="flex items-center justify-between">
            <h3 class="text-base font-bold text-slate-900 flex items-center space-x-2">
              <PieChart class="w-5 h-5 text-emerald-600" />
              <span>Komparasi Target Ideal vs Realisasi Riil</span>
            </h3>
            <span
              v-if="percentNeeds <= 100 && percentWants <= 100"
              class="inline-flex items-center space-x-1 text-xs font-semibold px-2.5 py-1 rounded-full bg-emerald-50 text-emerald-700 border border-emerald-200"
            >
              <CheckCircle2 class="w-3.5 h-3.5" />
              <span>Sesuai Anggaran</span>
            </span>
            <span
              v-else
              class="inline-flex items-center space-x-1 text-xs font-semibold px-2.5 py-1 rounded-full bg-rose-50 text-rose-700 border border-rose-200"
            >
              <AlertTriangle class="w-3.5 h-3.5" />
              <span>Overbudget</span>
            </span>
          </div>
          <p class="text-xs sm:text-sm text-slate-600 leading-relaxed">
            Metode 50-30-20 menjaga gaya hidup agar tidak memakan jatah tabungan masa depan maupun
            kebutuhan hidup pokok harian Anda.
          </p>
        </div>

        <!-- Tabel Perbandingan Pilar -->
        <div class="overflow-x-auto">
          <table class="w-full text-left text-xs sm:text-sm">
            <thead>
              <tr class="border-b border-slate-200 text-slate-500 font-semibold">
                <th class="pb-2.5">Pilar Anggaran</th>
                <th class="pb-2.5 text-right">Target (Ideal)</th>
                <th class="pb-2.5 text-right">Realisasi</th>
                <th class="pb-2.5 text-right">Sisa / Status</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-100 text-slate-800">
              <tr>
                <td class="py-3 font-semibold text-sky-800">🔵 Kebutuhan Pokok (50%)</td>
                <td class="py-3 text-right text-slate-600">{{ formatRupiah(targetNeeds) }}</td>
                <td class="py-3 text-right font-bold">{{ formatRupiah(actualNeeds) }}</td>
                <td class="py-3 text-right">
                  <span
                    :class="
                      remainNeeds >= 0 ? 'text-emerald-700 font-bold' : 'text-rose-600 font-bold'
                    "
                  >
                    {{
                      remainNeeds >= 0
                        ? formatRupiah(remainNeeds)
                        : `-${formatRupiah(Math.abs(remainNeeds))}`
                    }}
                  </span>
                </td>
              </tr>
              <tr>
                <td class="py-3 font-semibold text-purple-800">🟣 Kebutuhan Pribadi (30%)</td>
                <td class="py-3 text-right text-slate-600">{{ formatRupiah(targetWants) }}</td>
                <td class="py-3 text-right font-bold">{{ formatRupiah(actualWants) }}</td>
                <td class="py-3 text-right">
                  <span
                    :class="
                      remainWants >= 0 ? 'text-emerald-700 font-bold' : 'text-rose-600 font-bold'
                    "
                  >
                    {{
                      remainWants >= 0
                        ? formatRupiah(remainWants)
                        : `-${formatRupiah(Math.abs(remainWants))}`
                    }}
                  </span>
                </td>
              </tr>
              <tr>
                <td class="py-3 font-semibold text-emerald-800">🟢 Tabungan & Investasi (20%)</td>
                <td class="py-3 text-right text-slate-600">{{ formatRupiah(targetSavings) }}</td>
                <td class="py-3 text-right font-bold">{{ formatRupiah(actualSavings) }}</td>
                <td class="py-3 text-right text-emerald-700 font-bold">
                  {{ formatRupiah(remainSavings) }}
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <div
          class="bg-slate-50 p-3 rounded-xl border border-slate-100 text-xs text-slate-500 flex items-center justify-between"
        >
          <div>
            <span class="font-medium text-slate-700">Total Sisa Dana (Remain Funds):</span>
            <span class="block text-[11px] text-slate-400">Needs + Wants + Sisa Tabungan</span>
          </div>
          <span class="font-extrabold text-slate-900 text-sm">
            {{ formatRupiah(totalRemainBudget) }}
          </span>
        </div>
      </div>
    </div>

    <!-- 3 KOLOM RINCIAN ALOKASI ALA SPREADSHEET (Persis Template Pengguna) -->
    <div class="space-y-4">
      <div>
        <h3 class="text-base sm:text-lg font-bold text-slate-900 flex items-center space-x-2">
          <Layers class="w-5 h-5 text-emerald-600" />
          <span>Rincian Pengeluaran per Pilar (Spreadsheet Format)</span>
        </h3>
        <p class="text-xs text-slate-500">
          Daftar pengeluaran riil per kategori yang dikelompokkan ke dalam masing-masing pilar
          anggaran.
        </p>
      </div>

      <div class="grid grid-cols-1 md:grid-cols-3 gap-5">
        <!-- 1. KOLOM KEBUTUHAN POKOK -->
        <div
          class="bg-white rounded-2xl border border-slate-200 shadow-sm overflow-hidden flex flex-col"
        >
          <div
            class="bg-sky-500 text-white px-4 py-3 font-bold text-sm flex items-center justify-between"
          >
            <span>Kebutuhan Pokok (50%)</span>
            <span class="text-xs bg-white/20 px-2 py-0.5 rounded">Biaya</span>
          </div>

          <div class="p-4 flex-1 space-y-2 divide-y divide-slate-100">
            <div
              v-for="cat in needsCategories"
              :key="cat.id"
              class="pt-2 first:pt-0 flex items-center justify-between text-xs"
            >
              <span class="font-medium text-slate-700 truncate pr-2">{{ cat.name }}</span>
              <span class="font-bold text-slate-900 flex-shrink-0">
                {{ formatRupiah(cat.total) }}
              </span>
            </div>
            <div
              v-if="needsCategories.length === 0"
              class="py-6 text-center text-xs text-slate-400 italic"
            >
              Belum ada transaksi pokok di siklus ini
            </div>
          </div>

          <div class="bg-slate-50 border-t border-slate-200 p-4 space-y-1.5 text-xs">
            <div class="flex items-center justify-between font-bold text-slate-900">
              <span>Total Pengeluaran:</span>
              <span>{{ formatRupiah(actualNeeds) }}</span>
            </div>
            <div class="flex items-center justify-between font-semibold">
              <span :class="remainNeeds >= 0 ? 'text-slate-600' : 'text-rose-600'">Sisa Dana:</span>
              <span
                :class="
                  remainNeeds >= 0 ? 'text-emerald-700 font-bold' : 'text-rose-600 font-extrabold'
                "
              >
                {{
                  remainNeeds >= 0
                    ? formatRupiah(remainNeeds)
                    : `Overbudget (${formatRupiah(Math.abs(remainNeeds))})`
                }}
              </span>
            </div>
          </div>
        </div>

        <!-- 2. KOLOM KEBUTUHAN PRIBADI -->
        <div
          class="bg-white rounded-2xl border border-slate-200 shadow-sm overflow-hidden flex flex-col"
        >
          <div
            class="bg-rose-400 text-white px-4 py-3 font-bold text-sm flex items-center justify-between"
          >
            <span>Kebutuhan Pribadi (30%)</span>
            <span class="text-xs bg-white/20 px-2 py-0.5 rounded">Biaya</span>
          </div>

          <div class="p-4 flex-1 space-y-2 divide-y divide-slate-100">
            <div
              v-for="cat in wantsCategories"
              :key="cat.id"
              class="pt-2 first:pt-0 flex items-center justify-between text-xs"
            >
              <span class="font-medium text-slate-700 truncate pr-2">{{ cat.name }}</span>
              <span class="font-bold text-slate-900 flex-shrink-0">
                {{ formatRupiah(cat.total) }}
              </span>
            </div>
            <div
              v-if="wantsCategories.length === 0"
              class="py-6 text-center text-xs text-slate-400 italic"
            >
              Belum ada transaksi pribadi di siklus ini
            </div>
          </div>

          <div class="bg-slate-50 border-t border-slate-200 p-4 space-y-1.5 text-xs">
            <div class="flex items-center justify-between font-bold text-slate-900">
              <span>Total Pengeluaran:</span>
              <span>{{ formatRupiah(actualWants) }}</span>
            </div>
            <div class="flex items-center justify-between font-semibold">
              <span :class="remainWants >= 0 ? 'text-slate-600' : 'text-rose-600'">Sisa Dana:</span>
              <span
                :class="
                  remainWants >= 0 ? 'text-emerald-700 font-bold' : 'text-rose-600 font-extrabold'
                "
              >
                {{
                  remainWants >= 0
                    ? formatRupiah(remainWants)
                    : `Overbudget (${formatRupiah(Math.abs(remainWants))})`
                }}
              </span>
            </div>
          </div>
        </div>

        <!-- 3. KOLOM TABUNGAN & INVESTASI -->
        <div
          class="bg-white rounded-2xl border border-slate-200 shadow-sm overflow-hidden flex flex-col"
        >
          <div
            class="bg-amber-400 text-slate-900 px-4 py-3 font-bold text-sm flex items-center justify-between"
          >
            <span>Tabungan & Investasi (20%)</span>
            <span class="text-xs bg-black/10 px-2 py-0.5 rounded">Biaya</span>
          </div>

          <div class="p-4 flex-1 space-y-2 divide-y divide-slate-100">
            <div
              v-for="cat in savingsCategories"
              :key="cat.id"
              class="pt-2 first:pt-0 flex items-center justify-between text-xs"
            >
              <span class="font-medium text-slate-700 truncate pr-2">{{ cat.name }}</span>
              <span class="font-bold text-slate-900 flex-shrink-0">
                {{ formatRupiah(cat.total) }}
              </span>
            </div>
            <div
              v-if="savingsCategories.length === 0"
              class="py-6 text-center text-xs text-slate-400 italic"
            >
              Belum ada tabungan/investasi di siklus ini
            </div>
          </div>

          <div class="bg-slate-50 border-t border-slate-200 p-4 space-y-1.5 text-xs">
            <div class="flex items-center justify-between font-bold text-slate-900">
              <span>Total Tabungan:</span>
              <span>{{ formatRupiah(actualSavings) }}</span>
            </div>
            <div class="flex items-center justify-between font-semibold">
              <span class="text-slate-600">Sisa Target Tabungan:</span>
              <span class="text-emerald-700 font-bold">
                {{ formatRupiah(remainSavings) }}
              </span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- TABEL REKAPITULASI KATEGORI ALA SPREADSHEET (Persis Template Qty & Total) -->
    <div
      class="bg-white rounded-2xl border border-slate-200 shadow-sm overflow-hidden space-y-4 p-5 sm:p-6"
    >
      <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-2">
        <div>
          <h3 class="text-base sm:text-lg font-bold text-slate-900 flex items-center space-x-2">
            <ReceiptText class="w-5 h-5 text-emerald-600" />
            <span>Rekapitulasi Pengeluaran per Kategori</span>
          </h3>
          <p class="text-xs text-slate-500">
            Kuantitas transaksi, total nominal, dan persentase pengeluaran siklus ini.
          </p>
        </div>
        <div class="text-xs text-slate-500">
          Total Transaksi: <strong class="text-slate-900">{{ totalTransactionCount }}</strong>
        </div>
      </div>

      <!-- Tabel Responsif -->
      <div class="overflow-x-auto border border-slate-200 rounded-xl">
        <table class="w-full text-left text-xs sm:text-sm">
          <thead
            class="bg-slate-50 border-b border-slate-200 text-slate-600 font-bold uppercase tracking-wider text-[11px]"
          >
            <tr>
              <th class="py-3 px-4">Kategori</th>
              <th class="py-3 px-4">Pilar</th>
              <th class="py-3 px-4 text-center">Qty</th>
              <th class="py-3 px-4 text-right">Total</th>
              <th class="py-3 px-4 text-right">Percentage</th>
              <th class="py-3 px-4 w-32 hidden sm:table-cell">Visual</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100 text-slate-700">
            <tr
              v-for="cat in burnRate.category_breakdown"
              :key="cat.id"
              class="hover:bg-slate-50/70 transition"
            >
              <td class="py-3 px-4 font-semibold text-slate-900">
                <div class="flex items-center space-x-2">
                  <span
                    class="w-2.5 h-2.5 rounded-full flex-shrink-0"
                    :class="[
                      cat.pillar === 'needs'
                        ? 'bg-sky-500'
                        : cat.pillar === 'wants'
                          ? 'bg-purple-500'
                          : 'bg-emerald-500',
                    ]"
                  ></span>
                  <span>{{ cat.name }}</span>
                </div>
              </td>
              <td class="py-3 px-4">
                <span
                  class="text-[10px] font-bold px-2 py-0.5 rounded-full uppercase"
                  :class="[
                    cat.pillar === 'needs'
                      ? 'bg-sky-50 text-sky-700 border border-sky-200'
                      : cat.pillar === 'wants'
                        ? 'bg-purple-50 text-purple-700 border border-purple-200'
                        : 'bg-emerald-50 text-emerald-700 border border-emerald-200',
                  ]"
                >
                  {{ cat.pillar }}
                </span>
              </td>
              <td class="py-3 px-4 text-center font-bold text-slate-800">
                {{ cat.transaction_count || 1 }}
              </td>
              <td class="py-3 px-4 text-right font-extrabold text-slate-900">
                {{ formatRupiah(cat.total) }}
              </td>
              <td class="py-3 px-4 text-right font-bold text-slate-600">{{ cat.percentage }}%</td>
              <td class="py-3 px-4 hidden sm:table-cell">
                <div class="w-full bg-slate-100 h-2 rounded-full overflow-hidden">
                  <div
                    class="h-full rounded-full transition-all duration-300"
                    :class="[
                      cat.pillar === 'needs'
                        ? 'bg-sky-500'
                        : cat.pillar === 'wants'
                          ? 'bg-purple-500'
                          : 'bg-emerald-500',
                    ]"
                    :style="{ width: `${Math.min(cat.percentage, 100)}%` }"
                  ></div>
                </div>
              </td>
            </tr>

            <tr v-if="!burnRate.category_breakdown || burnRate.category_breakdown.length === 0">
              <td colspan="6" class="py-8 text-center text-slate-400 italic text-xs">
                Belum ada transaksi pengeluaran pada siklus ini.
              </td>
            </tr>
          </tbody>

          <!-- Footer Grand Total ala Spreadsheet -->
          <tfoot
            v-if="burnRate.category_breakdown && burnRate.category_breakdown.length > 0"
            class="bg-slate-50 border-t-2 border-slate-300 font-extrabold text-slate-900 text-xs sm:text-sm"
          >
            <tr>
              <td class="py-3 px-4" colspan="2">Grand Total Pengeluaran</td>
              <td class="py-3 px-4 text-center">{{ totalTransactionCount }}</td>
              <td class="py-3 px-4 text-right text-rose-600">
                {{ formatRupiah(grandTotalExpenses) }}
              </td>
              <td class="py-3 px-4 text-right">100%</td>
              <td class="py-3 px-4 hidden sm:table-cell"></td>
            </tr>
          </tfoot>
        </table>
      </div>
    </div>
  </div>
</template>
