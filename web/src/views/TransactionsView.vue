<script setup>
import { ref, computed, onMounted } from 'vue'
import {
  Plus,
  ArrowDownLeft,
  ArrowUpRight,
  ArrowLeftRight,
  Trash2,
  Calendar,
  Wallet,
  ChevronLeft,
  ChevronRight,
  X,
  Tags,
} from '@lucide/vue'

const props = defineProps({
  formatRupiah: {
    type: Function,
    required: true,
  },
})

const emit = defineEmits(['transaction-changed'])

const getTodayStr = () => new Date().toISOString().split('T')[0]
const getCurrentMonthStr = () => new Date().toISOString().slice(0, 7)

const currentMonth = ref(getCurrentMonthStr())
const transactions = ref([])
const accounts = ref([])
const categories = ref({
  needs_categories: [],
  wants_categories: [],
  savings_categories: [],
  income_categories: [],
  all_categories: [],
})
const loading = ref(true)
const filterType = ref('all')
const filterAccount = ref('all')
const filterPillar = ref('all')

const showAddModal = ref(false)
const showCategoryModal = ref(false)
const saving = ref(false)

// Form transaksi
const form = ref({
  type: 'expense',
  account_id: '',
  to_account_id: '',
  category_id: '',
  amount: null,
  transaction_date: getTodayStr(),
  description: '',
})

// Form tambah kategori kustom
const newCategory = ref({
  name: '',
  type: 'expense',
  pillar: 'needs',
})

const fetchAccounts = async () => {
  try {
    const res = await fetch('/api/accounts')
    if (res.ok) {
      const data = await res.json()
      accounts.value = [
        ...(data.operational_accounts || []),
        ...(data.passive_accounts || []),
        ...(data.emergency_accounts || []),
      ]
    }
  } catch (err) {
    console.error('Error fetching accounts:', err)
  }
}

const fetchCategories = async () => {
  try {
    const res = await fetch('/api/categories')
    if (res.ok) {
      categories.value = await res.json()
    }
  } catch (err) {
    console.error('Error fetching categories:', err)
  }
}

const fetchTransactions = async () => {
  loading.value = true
  try {
    let url = `/api/transactions?month=${currentMonth.value}`
    if (filterType.value !== 'all') url += `&type=${filterType.value}`
    if (filterAccount.value !== 'all') url += `&account_id=${filterAccount.value}`

    const res = await fetch(url)
    if (res.ok) {
      const data = await res.json()
      transactions.value = data.transactions || []
    }
  } catch (err) {
    console.error('Error fetching transactions:', err)
  } finally {
    loading.value = false
  }
}

const monthName = computed(() => {
  const [year, month] = currentMonth.value.split('-')
  const date = new Date(Number.parseInt(year, 10), Number.parseInt(month, 10) - 1, 1)
  return date.toLocaleDateString('id-ID', { month: 'long', year: 'numeric' })
})

const changeMonth = delta => {
  const [year, month] = currentMonth.value.split('-')
  const date = new Date(Number.parseInt(year, 10), Number.parseInt(month, 10) - 1 + delta, 1)
  const y = date.getFullYear()
  const m = String(date.getMonth() + 1).padStart(2, '0')
  currentMonth.value = `${y}-${m}`
  fetchTransactions()
}

const resetToCurrentMonth = () => {
  currentMonth.value = getCurrentMonthStr()
  fetchTransactions()
}

// Filtered transactions (termasuk filter pilar jika dipilih)
const filteredTransactions = computed(() => {
  return transactions.value.filter(t => {
    if (filterPillar.value !== 'all') {
      if (filterPillar.value === 'income' && t.type !== 'income') return false
      if (filterPillar.value !== 'income' && t.pillar !== filterPillar.value) return false
    }
    return true
  })
})

// Quick Monthly Summary
const monthlyTotalExpense = computed(() => {
  return transactions.value
    .filter(t => t.type === 'expense')
    .reduce((sum, t) => sum + (t.amount || 0), 0)
})

const monthlyTotalIncome = computed(() => {
  return transactions.value
    .filter(t => t.type === 'income')
    .reduce((sum, t) => sum + (t.amount || 0), 0)
})

const monthlyNetCashflow = computed(() => {
  return monthlyTotalIncome.value - monthlyTotalExpense.value
})

// Format label tanggal
const formatDateHeader = dateStr => {
  const today = getTodayStr()
  const yesterdayDate = new Date()
  yesterdayDate.setDate(yesterdayDate.getDate() - 1)
  const yesterday = yesterdayDate.toISOString().split('T')[0]

  const dateObj = new Date(dateStr + 'T00:00:00')
  const formatted = dateObj.toLocaleDateString('id-ID', {
    weekday: 'long',
    day: 'numeric',
    month: 'short',
    year: 'numeric',
  })

  if (dateStr === today) {
    return `Hari Ini — ${formatted}`
  } else if (dateStr === yesterday) {
    return `Kemarin — ${formatted}`
  }
  return formatted
}

// Group transactions by date
const groupedTransactions = computed(() => {
  const groups = {}
  for (const t of filteredTransactions.value) {
    const d = t.transaction_date
    if (!groups[d]) {
      groups[d] = {
        date: d,
        items: [],
        expenseSubtotal: 0,
        incomeSubtotal: 0,
      }
    }
    groups[d].items.push(t)
    if (t.type === 'expense') {
      groups[d].expenseSubtotal += t.amount
    } else if (t.type === 'income') {
      groups[d].incomeSubtotal += t.amount
    }
  }

  // Convert to sorted array (newest date first)
  return Object.values(groups).sort((a, b) => b.date.localeCompare(a.date))
})

const openAddModal = (type = 'expense') => {
  form.value = {
    type,
    account_id: accounts.value.length > 0 ? accounts.value[0].id : '',
    to_account_id: accounts.value.length > 1 ? accounts.value[1].id : '',
    category_id: '',
    amount: null,
    transaction_date: getTodayStr(),
    description: '',
  }
  showAddModal.value = true
}

const submitTransaction = async () => {
  if (!form.value.amount || form.value.amount <= 0) return
  saving.value = true

  try {
    let res
    if (form.value.type === 'transfer') {
      res = await fetch('/api/transactions/transfer', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          from_account_id: Number(form.value.account_id),
          to_account_id: Number(form.value.to_account_id),
          category_id: form.value.category_id ? Number(form.value.category_id) : null,
          amount: Number(form.value.amount),
          description: form.value.description,
          transaction_date: form.value.transaction_date,
        }),
      })
    } else {
      res = await fetch('/api/transactions', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          account_id: Number(form.value.account_id),
          category_id: form.value.category_id ? Number(form.value.category_id) : null,
          amount: Number(form.value.amount),
          type: form.value.type,
          description: form.value.description,
          transaction_date: form.value.transaction_date,
        }),
      })
    }

    if (res.ok) {
      showAddModal.value = false
      await fetchTransactions()
      await fetchAccounts()
      emit('transaction-changed')
    } else {
      const err = await res.text()
      alert('Gagal menyimpan transaksi: ' + err)
    }
  } catch (err) {
    console.error('Error submitting transaction:', err)
  } finally {
    saving.value = false
  }
}

const deleteTransaction = async id => {
  if (!confirm('Hapus transaksi ini? Saldo rekening akan otomatis dikembalikan.')) return

  try {
    const res = await fetch(`/api/transactions/${id}`, {
      method: 'DELETE',
    })
    if (res.ok) {
      await fetchTransactions()
      await fetchAccounts()
      emit('transaction-changed')
    }
  } catch (err) {
    console.error('Error deleting transaction:', err)
  }
}

const submitNewCategory = async () => {
  if (!newCategory.value.name.trim()) return
  try {
    const res = await fetch('/api/categories', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        name: newCategory.value.name.trim(),
        type: newCategory.value.type,
        pillar: newCategory.value.pillar,
      }),
    })
    if (res.ok) {
      newCategory.value.name = ''
      await fetchCategories()
    }
  } catch (err) {
    console.error('Error creating category:', err)
  }
}

const deleteCategory = async id => {
  if (!confirm('Hapus kategori ini?')) return
  try {
    const res = await fetch(`/api/categories/${id}`, {
      method: 'DELETE',
    })
    if (res.ok) {
      await fetchCategories()
    }
  } catch (err) {
    console.error('Error deleting category:', err)
  }
}

onMounted(() => {
  fetchAccounts()
  fetchCategories()
  fetchTransactions()
})
</script>

<template>
  <div class="space-y-6">
    <!-- Header: Bulan, Navigasi & Action Bar -->
    <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
      <div>
        <h2 class="text-xl sm:text-2xl font-bold tracking-tight text-slate-900">
          Catatan & Mutasi Harian
        </h2>
        <p class="text-xs sm:text-sm text-slate-500">
          Catat pengeluaran pos 50-30-20, pemasukan gaji, atau transfer saldo antar-rekening
        </p>
      </div>

      <div class="flex items-center space-x-2">
        <button
          type="button"
          class="inline-flex items-center space-x-1.5 px-3 py-2 rounded-xl text-xs font-semibold bg-white border border-slate-200 text-slate-700 hover:bg-slate-50 transition shadow-sm"
          @click="showCategoryModal = true"
        >
          <Tags class="w-4 h-4 text-slate-500" />
          <span>Kelola Kategori</span>
        </button>
        <button
          type="button"
          class="inline-flex items-center space-x-2 px-4 py-2 rounded-xl text-xs sm:text-sm font-bold bg-emerald-600 hover:bg-emerald-700 text-white shadow-sm shadow-emerald-200 transition"
          @click="openAddModal('expense')"
        >
          <Plus class="w-4 h-4" />
          <span>Catat Transaksi</span>
        </button>
      </div>
    </div>

    <!-- Banner Ringkasan Bulan Berjalan -->
    <div class="grid grid-cols-1 sm:grid-cols-3 gap-3 sm:gap-4">
      <!-- Total Pengeluaran Bulan Ini -->
      <div
        class="bg-white p-4 sm:p-5 rounded-2xl border border-slate-200 shadow-sm flex items-center justify-between"
      >
        <div>
          <span class="text-[11px] font-bold tracking-wider uppercase text-slate-400"
            >Total Pengeluaran</span
          >
          <div class="text-xl sm:text-2xl font-extrabold text-rose-600 mt-0.5">
            {{ props.formatRupiah(monthlyTotalExpense) }}
          </div>
          <span class="text-[10px] text-slate-400">{{ monthName }}</span>
        </div>
        <div
          class="w-10 h-10 rounded-xl bg-rose-50 text-rose-600 flex items-center justify-center flex-shrink-0"
        >
          <ArrowUpRight class="w-5 h-5" />
        </div>
      </div>

      <!-- Total Pemasukan Bulan Ini -->
      <div
        class="bg-white p-4 sm:p-5 rounded-2xl border border-slate-200 shadow-sm flex items-center justify-between"
      >
        <div>
          <span class="text-[11px] font-bold tracking-wider uppercase text-slate-400"
            >Total Pemasukan</span
          >
          <div class="text-xl sm:text-2xl font-extrabold text-emerald-600 mt-0.5">
            {{ props.formatRupiah(monthlyTotalIncome) }}
          </div>
          <span class="text-[10px] text-slate-400">{{ monthName }}</span>
        </div>
        <div
          class="w-10 h-10 rounded-xl bg-emerald-50 text-emerald-600 flex items-center justify-center flex-shrink-0"
        >
          <ArrowDownLeft class="w-5 h-5" />
        </div>
      </div>

      <!-- Arus Kas Bersih -->
      <div
        class="bg-white p-4 sm:p-5 rounded-2xl border border-slate-200 shadow-sm flex items-center justify-between"
      >
        <div>
          <span class="text-[11px] font-bold tracking-wider uppercase text-slate-400"
            >Arus Kas Bersih</span
          >
          <div
            class="text-xl sm:text-2xl font-extrabold mt-0.5"
            :class="monthlyNetCashflow >= 0 ? 'text-slate-900' : 'text-rose-600'"
          >
            {{ props.formatRupiah(monthlyNetCashflow) }}
          </div>
          <span class="text-[10px] text-slate-400">Pemasukan - Pengeluaran</span>
        </div>
        <div
          class="w-10 h-10 rounded-xl bg-slate-50 text-slate-600 flex items-center justify-center flex-shrink-0"
        >
          <Wallet class="w-5 h-5" />
        </div>
      </div>
    </div>

    <!-- Filter & Navigasi Bulan -->
    <div class="bg-white p-4 rounded-2xl border border-slate-200 shadow-sm space-y-3">
      <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3">
        <!-- Month Selector -->
        <div
          class="inline-flex items-center space-x-1.5 bg-slate-50 p-1 rounded-xl border border-slate-200 self-start"
        >
          <button
            type="button"
            class="p-1.5 hover:bg-white rounded-lg text-slate-600 transition"
            aria-label="Bulan Sebelumnya"
            @click="changeMonth(-1)"
          >
            <ChevronLeft class="w-4 h-4" />
          </button>
          <span class="px-2 text-xs sm:text-sm font-bold text-slate-800 min-w-[140px] text-center">
            {{ monthName }}
          </span>
          <button
            type="button"
            class="p-1.5 hover:bg-white rounded-lg text-slate-600 transition"
            aria-label="Bulan Selanjutnya"
            @click="changeMonth(1)"
          >
            <ChevronRight class="w-4 h-4" />
          </button>
          <button
            v-if="currentMonth !== getCurrentMonthStr()"
            type="button"
            class="ml-1 px-2 py-1 text-[11px] font-semibold bg-emerald-100 text-emerald-800 rounded-lg hover:bg-emerald-200 transition"
            @click="resetToCurrentMonth"
          >
            Bulan Ini
          </button>
        </div>

        <!-- Filter Dropdowns -->
        <div class="flex flex-wrap items-center gap-2">
          <!-- Filter Tipe Transaksi -->
          <div class="flex items-center space-x-1">
            <select
              id="filter-type"
              v-model="filterType"
              aria-label="Filter Tipe Transaksi"
              class="text-xs px-2.5 py-1.5 rounded-lg border border-slate-200 bg-slate-50 font-medium text-slate-700 focus:outline-none focus:ring-2 focus:ring-emerald-500"
              @change="fetchTransactions"
            >
              <option value="all">Semua Jenis</option>
              <option value="expense">Pengeluaran</option>
              <option value="income">Pemasukan</option>
              <option value="transfer">Transfer</option>
            </select>
          </div>

          <!-- Filter Pilar 50-30-20 -->
          <div class="flex items-center space-x-1">
            <select
              id="filter-pillar"
              v-model="filterPillar"
              aria-label="Filter Pilar Anggaran"
              class="text-xs px-2.5 py-1.5 rounded-lg border border-slate-200 bg-slate-50 font-medium text-slate-700 focus:outline-none focus:ring-2 focus:ring-emerald-500"
            >
              <option value="all">Semua Pilar</option>
              <option value="needs">50% Needs (Pokok)</option>
              <option value="wants">30% Wants (Gaya Hidup)</option>
              <option value="savings">20% Savings (Tabungan/Investasi)</option>
            </select>
          </div>

          <!-- Filter Akun -->
          <div class="flex items-center space-x-1">
            <select
              id="filter-account"
              v-model="filterAccount"
              aria-label="Filter Akun / Rekening"
              class="text-xs px-2.5 py-1.5 rounded-lg border border-slate-200 bg-slate-50 font-medium text-slate-700 focus:outline-none focus:ring-2 focus:ring-emerald-500"
              @change="fetchTransactions"
            >
              <option value="all">Semua Rekening</option>
              <option v-for="acc in accounts" :key="acc.id" :value="acc.id">
                {{ acc.name }}
              </option>
            </select>
          </div>
        </div>
      </div>
    </div>

    <!-- Riwayat Transaksi Harian (Grouped By Date) -->
    <div v-if="loading" class="text-center py-12 text-slate-400 text-sm">
      Memuat catatan transaksi...
    </div>

    <div
      v-else-if="groupedTransactions.length === 0"
      class="bg-white rounded-2xl p-10 border border-slate-200 text-center space-y-3"
    >
      <div
        class="w-12 h-12 rounded-2xl bg-slate-50 text-slate-400 flex items-center justify-center mx-auto"
      >
        <Calendar class="w-6 h-6" />
      </div>
      <h3 class="font-bold text-slate-800 text-base">Belum Ada Transaksi di Bulan Ini</h3>
      <p class="text-xs text-slate-500 max-w-sm mx-auto">
        Mulai catat transaksi pengeluaran harian, gaji masuk, atau transfer antar rekening untuk
        memantau arus kas Anda.
      </p>
      <button
        type="button"
        class="inline-flex items-center space-x-1.5 px-4 py-2 bg-emerald-600 text-white rounded-xl text-xs font-bold shadow-sm hover:bg-emerald-700 transition mt-2"
        @click="openAddModal('expense')"
      >
        <Plus class="w-4 h-4" />
        <span>Catat Transaksi Pertama</span>
      </button>
    </div>

    <div v-else class="space-y-4">
      <div
        v-for="group in groupedTransactions"
        :key="group.date"
        class="bg-white rounded-2xl border border-slate-200 shadow-sm overflow-hidden"
      >
        <!-- Header Grup Tanggal (Ala Spreadsheet) -->
        <div
          class="px-4 sm:px-5 py-3 bg-slate-50/80 border-b border-slate-100 flex items-center justify-between text-xs sm:text-sm"
        >
          <div class="flex items-center space-x-2 font-bold text-slate-800">
            <Calendar class="w-4 h-4 text-slate-500" />
            <span>{{ formatDateHeader(group.date) }}</span>
          </div>

          <div class="flex items-center space-x-3 text-xs">
            <span v-if="group.expenseSubtotal > 0" class="text-rose-600 font-bold">
              Subtotal: -{{ props.formatRupiah(group.expenseSubtotal) }}
            </span>
            <span v-if="group.incomeSubtotal > 0" class="text-emerald-600 font-bold">
              +{{ props.formatRupiah(group.incomeSubtotal) }}
            </span>
          </div>
        </div>

        <!-- Daftar Transaksi pada Tanggal Tersebut -->
        <ul class="divide-y divide-slate-100">
          <li
            v-for="item in group.items"
            :key="item.id"
            class="px-4 sm:px-5 py-3 hover:bg-slate-50/50 transition flex items-center justify-between gap-3 text-xs sm:text-sm"
          >
            <!-- Kolom Kiri: Ikon & Detail -->
            <div class="flex items-center space-x-3 min-w-0">
              <div
                class="w-8 h-8 sm:w-9 sm:h-9 rounded-xl flex items-center justify-center flex-shrink-0"
                :class="[
                  item.type === 'expense'
                    ? 'bg-rose-50 text-rose-600'
                    : item.type === 'income'
                      ? 'bg-emerald-50 text-emerald-600'
                      : 'bg-blue-50 text-blue-600',
                ]"
              >
                <ArrowUpRight v-if="item.type === 'expense'" class="w-4 h-4" />
                <ArrowDownLeft v-else-if="item.type === 'income'" class="w-4 h-4" />
                <ArrowLeftRight v-else class="w-4 h-4" />
              </div>

              <div class="min-w-0">
                <div class="flex items-center space-x-2 flex-wrap gap-y-1">
                  <span class="font-bold text-slate-900 truncate">
                    {{
                      item.description ||
                      item.category_name ||
                      (item.type === 'transfer' ? 'Transfer Saldo' : 'Transaksi')
                    }}
                  </span>

                  <!-- Badge Pilar 50-30-20 -->
                  <span
                    v-if="item.pillar === 'needs'"
                    class="px-1.5 py-0.5 rounded text-[10px] font-bold bg-blue-50 text-blue-700 border border-blue-200"
                  >
                    50% Needs
                  </span>
                  <span
                    v-else-if="item.pillar === 'wants'"
                    class="px-1.5 py-0.5 rounded text-[10px] font-bold bg-purple-50 text-purple-700 border border-purple-200"
                  >
                    30% Wants
                  </span>
                  <span
                    v-else-if="item.pillar === 'savings'"
                    class="px-1.5 py-0.5 rounded text-[10px] font-bold bg-emerald-50 text-emerald-700 border border-emerald-200"
                  >
                    20% Savings
                  </span>
                  <span
                    v-else-if="item.type === 'transfer'"
                    class="px-1.5 py-0.5 rounded text-[10px] font-bold bg-slate-100 text-slate-700"
                  >
                    Transfer
                  </span>
                </div>

                <div
                  class="text-[11px] text-slate-500 flex items-center space-x-1.5 mt-0.5 truncate"
                >
                  <span v-if="item.category_name" class="font-medium text-slate-600">
                    {{ item.category_name }} •
                  </span>
                  <span v-if="item.type === 'transfer'">
                    {{ item.account_name }} &rarr; {{ item.to_account_name }}
                  </span>
                  <span v-else>
                    {{ item.account_name }}
                  </span>
                </div>
              </div>
            </div>

            <!-- Kolom Kanan: Nominal & Tombol Hapus -->
            <div class="flex items-center space-x-3 flex-shrink-0">
              <div
                class="text-right font-extrabold"
                :class="[
                  item.type === 'expense'
                    ? 'text-rose-600'
                    : item.type === 'income'
                      ? 'text-emerald-600'
                      : 'text-slate-800',
                ]"
              >
                {{ item.type === 'expense' ? '-' : item.type === 'income' ? '+' : ''
                }}{{ props.formatRupiah(item.amount) }}
              </div>

              <button
                type="button"
                class="p-1.5 rounded-lg text-slate-400 hover:text-rose-600 hover:bg-rose-50 transition"
                title="Hapus Transaksi"
                aria-label="Hapus Transaksi"
                @click="deleteTransaction(item.id)"
              >
                <Trash2 class="w-4 h-4" />
              </button>
            </div>
          </li>
        </ul>
      </div>
    </div>

    <!-- MODAL CATAT TRANSAKSI -->
    <div
      v-if="showAddModal"
      class="fixed inset-0 z-50 bg-slate-900/60 backdrop-blur-sm flex items-center justify-center p-4"
    >
      <div
        class="bg-white rounded-2xl max-w-md w-full p-6 shadow-xl space-y-4 max-h-[90vh] overflow-y-auto"
      >
        <div class="flex items-center justify-between border-b border-slate-100 pb-3">
          <h3 class="font-bold text-slate-800 text-base">Catat Mutasi Transaksi</h3>
          <button
            type="button"
            class="text-slate-400 hover:text-slate-600"
            aria-label="Tutup Modal"
            @click="showAddModal = false"
          >
            <X class="w-5 h-5" />
          </button>
        </div>

        <!-- Tab Pemilih Jenis Transaksi -->
        <div>
          <span class="block font-semibold text-slate-700 text-xs mb-1.5">Jenis Transaksi</span>
          <div class="grid grid-cols-3 gap-2">
            <button
              type="button"
              :class="[
                form.type === 'expense'
                  ? 'bg-rose-50 border-rose-500 text-rose-700 font-bold'
                  : 'border-slate-200 text-slate-600 hover:bg-slate-50',
              ]"
              class="p-2 rounded-xl border text-center text-xs transition"
              @click="form.type = 'expense'"
            >
              Pengeluaran
            </button>
            <button
              type="button"
              :class="[
                form.type === 'income'
                  ? 'bg-emerald-50 border-emerald-500 text-emerald-700 font-bold'
                  : 'border-slate-200 text-slate-600 hover:bg-slate-50',
              ]"
              class="p-2 rounded-xl border text-center text-xs transition"
              @click="form.type = 'income'"
            >
              Pemasukan
            </button>
            <button
              type="button"
              :class="[
                form.type === 'transfer'
                  ? 'bg-blue-50 border-blue-500 text-blue-700 font-bold'
                  : 'border-slate-200 text-slate-600 hover:bg-slate-50',
              ]"
              class="p-2 rounded-xl border text-center text-xs transition"
              @click="form.type = 'transfer'"
            >
              Transfer
            </button>
          </div>
        </div>

        <form class="space-y-4 text-xs sm:text-sm" @submit.prevent="submitTransaction">
          <!-- Jumlah Nominal -->
          <div>
            <label for="tx-amount" class="block font-semibold text-slate-700 mb-1"
              >Jumlah Nominal (Rp)</label
            >
            <input
              id="tx-amount"
              v-model.number="form.amount"
              type="number"
              min="1"
              required
              placeholder="Contoh: 50000"
              class="w-full px-3.5 py-2.5 rounded-xl border border-slate-200 focus:outline-none focus:ring-2 focus:ring-emerald-500 font-bold text-base"
            />
          </div>

          <!-- Pilihan Rekening -->
          <div v-if="form.type !== 'transfer'">
            <label for="tx-account" class="block font-semibold text-slate-700 mb-1">
              {{ form.type === 'expense' ? 'Dari Rekening / Dompet' : 'Masuk Ke Rekening' }}
            </label>
            <select
              id="tx-account"
              v-model="form.account_id"
              required
              class="w-full px-3 py-2.5 rounded-xl border border-slate-200 bg-white focus:outline-none focus:ring-2 focus:ring-emerald-500"
            >
              <option value="" disabled>Pilih Rekening</option>
              <option v-for="acc in accounts" :key="acc.id" :value="acc.id">
                {{ acc.name }} (Saldo: {{ props.formatRupiah(acc.balance) }})
              </option>
            </select>
          </div>

          <!-- Pilihan Rekening Transfer (Asal & Tujuan) -->
          <div v-else class="grid grid-cols-2 gap-2">
            <div>
              <label for="tx-from-account" class="block font-semibold text-slate-700 mb-1"
                >Dari Rekening</label
              >
              <select
                id="tx-from-account"
                v-model="form.account_id"
                required
                class="w-full px-2.5 py-2 rounded-xl border border-slate-200 bg-white text-xs focus:outline-none focus:ring-2 focus:ring-blue-500"
              >
                <option value="" disabled>Pilih Asal</option>
                <option v-for="acc in accounts" :key="acc.id" :value="acc.id">
                  {{ acc.name }}
                </option>
              </select>
            </div>
            <div>
              <label for="tx-to-account" class="block font-semibold text-slate-700 mb-1"
                >Ke Rekening</label
              >
              <select
                id="tx-to-account"
                v-model="form.to_account_id"
                required
                class="w-full px-2.5 py-2 rounded-xl border border-slate-200 bg-white text-xs focus:outline-none focus:ring-2 focus:ring-blue-500"
              >
                <option value="" disabled>Pilih Tujuan</option>
                <option v-for="acc in accounts" :key="acc.id" :value="acc.id">
                  {{ acc.name }}
                </option>
              </select>
            </div>
          </div>

          <!-- Kategori 50-30-20 -->
          <div v-if="form.type !== 'transfer' || form.type === 'transfer'">
            <label for="tx-category" class="block font-semibold text-slate-700 mb-1">
              Kategori Pos Anggaran
            </label>
            <select
              id="tx-category"
              v-model="form.category_id"
              class="w-full px-3 py-2.5 rounded-xl border border-slate-200 bg-white focus:outline-none focus:ring-2 focus:ring-emerald-500"
            >
              <option :value="''">Tanpa Kategori</option>

              <!-- Kategori Expense: Needs & Wants -->
              <optgroup v-if="form.type === 'expense'" label="🔵 50% Kebutuhan Pokok (Needs)">
                <option v-for="c in categories.needs_categories" :key="c.id" :value="c.id">
                  {{ c.name }}
                </option>
              </optgroup>
              <optgroup
                v-if="form.type === 'expense'"
                label="🟣 30% Keinginan & Gaya Hidup (Wants)"
              >
                <option v-for="c in categories.wants_categories" :key="c.id" :value="c.id">
                  {{ c.name }}
                </option>
              </optgroup>

              <!-- Kategori Savings (Bisa untuk transfer tabungan atau expense investasi) -->
              <optgroup label="🟢 20% Tabungan & Investasi (Savings)">
                <option v-for="c in categories.savings_categories" :key="c.id" :value="c.id">
                  {{ c.name }}
                </option>
              </optgroup>

              <!-- Kategori Income -->
              <optgroup v-if="form.type === 'income'" label="💰 Pemasukan">
                <option v-for="c in categories.income_categories" :key="c.id" :value="c.id">
                  {{ c.name }}
                </option>
              </optgroup>
            </select>
          </div>

          <!-- Tanggal Transaksi -->
          <div>
            <label for="tx-date" class="block font-semibold text-slate-700 mb-1"
              >Tanggal Transaksi</label
            >
            <input
              id="tx-date"
              v-model="form.transaction_date"
              type="date"
              required
              class="w-full px-3.5 py-2.5 rounded-xl border border-slate-200 focus:outline-none focus:ring-2 focus:ring-emerald-500 font-medium"
            />
          </div>

          <!-- Deskripsi / Catatan -->
          <div>
            <label for="tx-description" class="block font-semibold text-slate-700 mb-1"
              >Deskripsi / Catatan</label
            >
            <input
              id="tx-description"
              v-model="form.description"
              type="text"
              placeholder="Contoh: Makan siang nasi padang / Gaji bulanan"
              class="w-full px-3.5 py-2.5 rounded-xl border border-slate-200 focus:outline-none focus:ring-2 focus:ring-emerald-500"
            />
          </div>

          <div class="pt-2 flex items-center justify-end space-x-2">
            <button
              type="button"
              class="px-4 py-2 rounded-xl text-slate-600 hover:bg-slate-100 font-medium"
              @click="showAddModal = false"
            >
              Batal
            </button>
            <button
              type="submit"
              :disabled="saving"
              class="px-5 py-2 bg-emerald-600 hover:bg-emerald-700 text-white rounded-xl font-bold shadow-sm"
            >
              {{ saving ? 'Menyimpan...' : 'Simpan Transaksi' }}
            </button>
          </div>
        </form>
      </div>
    </div>

    <!-- MODAL KELOLA KATEGORI -->
    <div
      v-if="showCategoryModal"
      class="fixed inset-0 z-50 bg-slate-900/60 backdrop-blur-sm flex items-center justify-center p-4"
    >
      <div
        class="bg-white rounded-2xl max-w-md w-full p-6 shadow-xl space-y-4 max-h-[90vh] overflow-y-auto"
      >
        <div class="flex items-center justify-between border-b border-slate-100 pb-3">
          <div>
            <h3 class="font-bold text-slate-800 text-base">Kelola Kategori Anggaran</h3>
            <p class="text-xs text-slate-500">Prinsip 50% Needs, 30% Wants, 20% Savings</p>
          </div>
          <button
            type="button"
            class="text-slate-400 hover:text-slate-600"
            aria-label="Tutup Modal"
            @click="showCategoryModal = false"
          >
            <X class="w-5 h-5" />
          </button>
        </div>

        <!-- Form Tambah Kategori Baru -->
        <form class="p-3 bg-slate-50 rounded-xl space-y-3" @submit.prevent="submitNewCategory">
          <span class="block font-bold text-xs text-slate-700">Tambah Kategori Baru</span>
          <div class="space-y-2">
            <div>
              <label for="cat-name" class="block text-[11px] font-medium text-slate-600 mb-1"
                >Nama Kategori</label
              >
              <input
                id="cat-name"
                v-model="newCategory.name"
                type="text"
                placeholder="Contoh: Perawatan Hewan"
                required
                class="w-full px-3 py-2 text-xs rounded-lg border border-slate-200 bg-white focus:outline-none focus:ring-2 focus:ring-emerald-500"
              />
            </div>

            <div class="grid grid-cols-2 gap-2">
              <div>
                <label for="cat-type" class="block text-[11px] font-medium text-slate-600 mb-1"
                  >Jenis</label
                >
                <select
                  id="cat-type"
                  v-model="newCategory.type"
                  class="w-full px-2.5 py-2 text-xs rounded-lg border border-slate-200 bg-white focus:outline-none focus:ring-2 focus:ring-emerald-500"
                >
                  <option value="expense">Pengeluaran</option>
                  <option value="income">Pemasukan</option>
                </select>
              </div>

              <div>
                <label for="cat-pillar" class="block text-[11px] font-medium text-slate-600 mb-1"
                  >Pilar 50-30-20</label
                >
                <select
                  id="cat-pillar"
                  v-model="newCategory.pillar"
                  class="w-full px-2.5 py-2 text-xs rounded-lg border border-slate-200 bg-white focus:outline-none focus:ring-2 focus:ring-emerald-500"
                >
                  <option value="needs">50% Needs</option>
                  <option value="wants">30% Wants</option>
                  <option value="savings">20% Savings</option>
                  <option value="income">Income</option>
                </select>
              </div>
            </div>
          </div>

          <button
            type="submit"
            class="w-full py-2 bg-slate-900 hover:bg-slate-800 text-white rounded-lg font-bold text-xs transition"
          >
            + Tambah Kategori
          </button>
        </form>

        <!-- Daftar Kategori Saat Ini -->
        <div class="space-y-3 text-xs max-h-60 overflow-y-auto">
          <!-- Needs -->
          <div>
            <span class="block font-bold text-blue-700 text-[11px] mb-1"
              >50% Kebutuhan Pokok (Needs)</span
            >
            <div class="space-y-1">
              <div
                v-for="c in categories.needs_categories"
                :key="c.id"
                class="flex items-center justify-between p-2 rounded-lg bg-slate-50"
              >
                <span class="font-medium text-slate-800">{{ c.name }}</span>
                <button
                  type="button"
                  class="text-slate-400 hover:text-rose-600 p-1"
                  aria-label="Hapus Kategori"
                  @click="deleteCategory(c.id)"
                >
                  <X class="w-3.5 h-3.5" />
                </button>
              </div>
            </div>
          </div>

          <!-- Wants -->
          <div>
            <span class="block font-bold text-purple-700 text-[11px] mb-1"
              >30% Keinginan (Wants)</span
            >
            <div class="space-y-1">
              <div
                v-for="c in categories.wants_categories"
                :key="c.id"
                class="flex items-center justify-between p-2 rounded-lg bg-slate-50"
              >
                <span class="font-medium text-slate-800">{{ c.name }}</span>
                <button
                  type="button"
                  class="text-slate-400 hover:text-rose-600 p-1"
                  aria-label="Hapus Kategori"
                  @click="deleteCategory(c.id)"
                >
                  <X class="w-3.5 h-3.5" />
                </button>
              </div>
            </div>
          </div>

          <!-- Savings -->
          <div>
            <span class="block font-bold text-emerald-700 text-[11px] mb-1"
              >20% Tabungan & Investasi (Savings)</span
            >
            <div class="space-y-1">
              <div
                v-for="c in categories.savings_categories"
                :key="c.id"
                class="flex items-center justify-between p-2 rounded-lg bg-slate-50"
              >
                <span class="font-medium text-slate-800">{{ c.name }}</span>
                <button
                  type="button"
                  class="text-slate-400 hover:text-rose-600 p-1"
                  aria-label="Hapus Kategori"
                  @click="deleteCategory(c.id)"
                >
                  <X class="w-3.5 h-3.5" />
                </button>
              </div>
            </div>
          </div>
        </div>

        <div class="pt-2 flex justify-end">
          <button
            type="button"
            class="px-4 py-2 bg-slate-100 hover:bg-slate-200 text-slate-700 rounded-xl font-bold text-xs"
            @click="showCategoryModal = false"
          >
            Selesai
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
