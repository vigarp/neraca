<script setup>
import { ref, onMounted, computed } from 'vue'
import {
  Wallet,
  TrendingUp,
  Plus,
  Edit3,
  Trash2,
  Calendar,
  X,
  CreditCard,
  Building2,
  Coins,
  ShieldCheck,
  Check,
} from '@lucide/vue'

const props = defineProps({
  formatRupiah: {
    type: Function,
    required: true,
  },
})

const loading = ref(true)
const accountsData = ref({
  operational_accounts: [],
  passive_accounts: [],
  emergency_accounts: [],
  total_operational_balance: 0,
  total_passive_assets: 0,
  total_emergency_balance: 0,
  total_net_worth: 0,
})

const settings = ref({
  payday_date: 25,
  monthly_income_budget: 6000000,
})

// Modal states
const showAddModal = ref(false)
const showRevalueModal = ref(false)
const showSettingsModal = ref(false)

// Form Tambah Akun
const formAccount = ref({
  name: '',
  account_group: 'operational',
  type: 'bank',
  balance: 0,
  institution: '',
})

// Form Revaluasi
const selectedAsset = ref(null)
const revalueForm = ref({
  new_balance: 0,
  notes: '',
})

// Form Settings
const settingForm = ref({
  payday_date: 25,
  monthly_income_budget: 6000000,
})

const fetchAccounts = async () => {
  loading.value = true
  try {
    const res = await fetch('/api/accounts')
    if (res.ok) {
      accountsData.value = await res.json()
    }
  } catch (err) {
    console.error('Error fetching accounts:', err)
  } finally {
    loading.value = false
  }
}

const fetchSettings = async () => {
  try {
    const res = await fetch('/api/settings')
    if (res.ok) {
      const data = await res.json()
      settings.value = data
      settingForm.value = { ...data }
    }
  } catch (err) {
    console.error('Error fetching settings:', err)
  }
}

onMounted(() => {
  fetchAccounts()
  fetchSettings()
})

const openAddAccount = (defaultGroup = 'operational') => {
  formAccount.value = {
    name: '',
    account_group: defaultGroup,
    type: defaultGroup === 'operational' ? 'bank' : 'mutual_fund',
    balance: 0,
    institution: '',
  }
  showAddModal.value = true
}

const submitAddAccount = async () => {
  if (!formAccount.value.name.trim()) return
  try {
    const res = await fetch('/api/accounts', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(formAccount.value),
    })
    if (res.ok) {
      showAddModal.value = false
      await fetchAccounts()
    }
  } catch (err) {
    console.error('Error adding account:', err)
  }
}

const openRevalue = asset => {
  selectedAsset.value = asset
  revalueForm.value = {
    new_balance: asset.balance,
    notes: 'Update valuasi pasar berkala',
  }
  showRevalueModal.value = true
}

const revalueDifference = computed(() => {
  if (!selectedAsset.value) return 0
  return Number(revalueForm.value.new_balance || 0) - selectedAsset.value.balance
})

const submitRevalue = async () => {
  if (!selectedAsset.value) return
  try {
    const res = await fetch(`/api/accounts/${selectedAsset.value.id}/revalue`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        new_balance: Number(revalueForm.value.new_balance),
        notes: revalueForm.value.notes,
      }),
    })
    if (res.ok) {
      showRevalueModal.value = false
      await fetchAccounts()
    }
  } catch (err) {
    console.error('Error revaluing asset:', err)
  }
}

const deleteAccount = async id => {
  if (!confirm('Apakah Anda yakin ingin menghapus akun ini?')) return
  try {
    const res = await fetch(`/api/accounts/${id}`, { method: 'DELETE' })
    if (res.ok) {
      await fetchAccounts()
    }
  } catch (err) {
    console.error('Error deleting account:', err)
  }
}

const submitSettings = async () => {
  try {
    const res = await fetch('/api/settings', {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(settingForm.value),
    })
    if (res.ok) {
      settings.value = { ...settingForm.value }
      showSettingsModal.value = false
    }
  } catch (err) {
    console.error('Error saving settings:', err)
  }
}

// Icon helper
const getAccountIcon = type => {
  switch (type) {
    case 'bank':
      return Building2
    case 'ewallet':
      return CreditCard
    case 'cash':
      return Coins
    case 'mutual_fund':
    case 'stocks':
    case 'crypto':
      return TrendingUp
    case 'gold':
      return Coins
    default:
      return Wallet
  }
}
</script>

<template>
  <div class="space-y-6">
    <!-- Top Summary Banner: Net Worth & Payday Setting -->
    <div class="grid grid-cols-1 lg:grid-cols-3 gap-4 sm:gap-5">
      <!-- Total Net Worth Card -->
      <div
        class="lg:col-span-2 bg-gradient-to-br from-slate-900 via-slate-800 to-emerald-950 rounded-2xl p-5 sm:p-6 text-white shadow-md relative overflow-hidden"
      >
        <div class="relative z-10 flex flex-col justify-between h-full space-y-4">
          <div>
            <div
              class="flex items-center space-x-2 text-emerald-400 text-xs font-semibold uppercase tracking-wider"
            >
              <ShieldCheck class="w-4 h-4" />
              <span>Total Kekayaan Bersih (Net Worth)</span>
            </div>
            <div class="text-3xl sm:text-4xl font-extrabold tracking-tight mt-1">
              {{ props.formatRupiah(accountsData.total_net_worth) }}
            </div>
            <p class="text-xs text-slate-300 mt-1">
              Akumulasi dari uang kas operasional, dana darurat, dan portofolio aset pasif.
            </p>
          </div>

          <!-- Breakdown Ringkas -->
          <div
            class="grid grid-cols-2 sm:grid-cols-3 gap-3 pt-3 border-t border-slate-700/60 text-xs"
          >
            <div>
              <span class="text-slate-400 block">Kas Operasional</span>
              <span class="font-bold text-emerald-300 text-sm">
                {{ props.formatRupiah(accountsData.total_operational_balance) }}
              </span>
            </div>
            <div>
              <span class="text-slate-400 block">Aset Pasif & Investasi</span>
              <span class="font-bold text-blue-300 text-sm">
                {{ props.formatRupiah(accountsData.total_passive_assets) }}
              </span>
            </div>
            <div class="col-span-2 sm:col-span-1">
              <span class="text-slate-400 block">Dana Darurat</span>
              <span class="font-bold text-amber-300 text-sm">
                {{ props.formatRupiah(accountsData.total_emergency_balance) }}
              </span>
            </div>
          </div>
        </div>
      </div>

      <!-- Payday Cycle Widget -->
      <div
        class="bg-white rounded-2xl p-5 sm:p-6 border border-slate-200 shadow-sm flex flex-col justify-between space-y-4"
      >
        <div>
          <div
            class="flex items-center justify-between text-xs font-semibold text-slate-500 uppercase tracking-wider"
          >
            <span class="flex items-center space-x-1.5">
              <Calendar class="w-4 h-4 text-emerald-600" />
              <span>Siklus Gajian</span>
            </span>
            <button
              type="button"
              class="text-emerald-600 hover:text-emerald-700 font-semibold lowercase text-xs"
              @click="showSettingsModal = true"
            >
              Ubah
            </button>
          </div>
          <div class="mt-3">
            <div class="text-2xl sm:text-3xl font-bold text-slate-800">
              Tiap Tgl {{ settings.payday_date }}
            </div>
            <p class="text-xs text-slate-500 mt-1">
              Basis estimasi: {{ props.formatRupiah(settings.monthly_income_budget) }} / bulan
            </p>
          </div>
        </div>

        <button
          type="button"
          class="w-full py-2 px-3 rounded-xl border border-slate-200 text-slate-700 text-xs font-semibold hover:bg-slate-50 transition flex items-center justify-center space-x-1.5 min-h-[44px]"
          @click="showSettingsModal = true"
        >
          <Calendar class="w-3.5 h-3.5 text-slate-500" />
          <span>Atur Tanggal & Pendapatan</span>
        </button>
      </div>
    </div>

    <!-- Section 1: Kas Operasional (Daily Spending Pool) -->
    <div class="bg-white rounded-2xl border border-slate-200 shadow-sm overflow-hidden">
      <div
        class="p-5 sm:p-6 border-b border-slate-100 flex flex-col sm:flex-row sm:items-center justify-between gap-3"
      >
        <div>
          <div class="flex items-center space-x-2">
            <span class="w-2.5 h-2.5 rounded-full bg-emerald-500"></span>
            <h2 class="text-base sm:text-lg font-bold text-slate-900">Kas & Dompet Operasional</h2>
          </div>
          <p class="text-xs text-slate-500 mt-0.5">
            Sumber dana belanja harian. Saldo di sinilah yang menjadi dasar hitungan
            <strong>Prospect Daily Limit</strong> Anda.
          </p>
        </div>
        <button
          type="button"
          class="inline-flex items-center justify-center space-x-1.5 px-3.5 py-2 bg-emerald-600 hover:bg-emerald-700 text-white rounded-xl text-xs font-semibold shadow-sm transition min-h-[44px] flex-shrink-0"
          @click="openAddAccount('operational')"
        >
          <Plus class="w-4 h-4" />
          <span>Tambah Dompet Kas</span>
        </button>
      </div>

      <!-- Account List: Operational -->
      <div class="divide-y divide-slate-100">
        <div
          v-for="acc in accountsData.operational_accounts"
          :key="acc.id"
          class="p-4 sm:p-5 flex items-center justify-between hover:bg-slate-50/70 transition"
        >
          <div class="flex items-center space-x-3.5">
            <div
              class="w-10 h-10 rounded-xl bg-emerald-50 text-emerald-700 flex items-center justify-center flex-shrink-0 font-bold"
            >
              <component :is="getAccountIcon(acc.type)" class="w-5 h-5" />
            </div>
            <div>
              <div class="font-bold text-slate-800 text-sm sm:text-base">{{ acc.name }}</div>
              <div class="text-xs text-slate-500 flex items-center space-x-2">
                <span class="capitalize">{{ acc.type }}</span>
                <span v-if="acc.institution">&bull; {{ acc.institution }}</span>
              </div>
            </div>
          </div>
          <div class="flex items-center space-x-3">
            <div class="text-right">
              <div class="text-sm sm:text-base font-extrabold text-slate-900">
                {{ props.formatRupiah(acc.balance) }}
              </div>
              <span
                class="text-[10px] uppercase tracking-wider text-emerald-600 bg-emerald-50 font-bold px-2 py-0.5 rounded-full"
              >
                Kas Aktif
              </span>
            </div>
            <button
              type="button"
              aria-label="Hapus akun"
              class="p-2 text-slate-400 hover:text-rose-600 rounded-lg hover:bg-rose-50 transition"
              @click="deleteAccount(acc.id)"
            >
              <Trash2 class="w-4 h-4" />
            </button>
          </div>
        </div>

        <div
          v-if="accountsData.operational_accounts.length === 0 && !loading"
          class="p-8 text-center text-slate-400 text-xs"
        >
          Belum ada dompet kas operasional. Klik "Tambah Dompet Kas" di atas (misal: BCA, GoPay).
        </div>
      </div>
    </div>

    <!-- Section 2: Aset Pasif & Investasi (Wealth Portfolio) -->
    <div class="bg-white rounded-2xl border border-slate-200 shadow-sm overflow-hidden">
      <div
        class="p-5 sm:p-6 border-b border-slate-100 flex flex-col sm:flex-row sm:items-center justify-between gap-3"
      >
        <div>
          <div class="flex items-center space-x-2">
            <span class="w-2.5 h-2.5 rounded-full bg-blue-500"></span>
            <h2 class="text-base sm:text-lg font-bold text-slate-900">
              Aset Pasif & Portofolio Investasi
            </h2>
          </div>
          <p class="text-xs text-slate-500 mt-0.5">
            Aset kekayaan pasif. Nilai dapat di-update berkala dan
            <strong>TIDAK akan mengganggu jatah belanja harian</strong>.
          </p>
        </div>
        <button
          type="button"
          class="inline-flex items-center justify-center space-x-1.5 px-3.5 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-xl text-xs font-semibold shadow-sm transition min-h-[44px] flex-shrink-0"
          @click="openAddAccount('passive_asset')"
        >
          <Plus class="w-4 h-4" />
          <span>Tambah Aset Pasif</span>
        </button>
      </div>

      <!-- Account List: Passive Assets -->
      <div class="divide-y divide-slate-100">
        <div
          v-for="acc in accountsData.passive_accounts"
          :key="acc.id"
          class="p-4 sm:p-5 flex flex-col sm:flex-row sm:items-center justify-between gap-3 hover:bg-slate-50/70 transition"
        >
          <div class="flex items-center space-x-3.5">
            <div
              class="w-10 h-10 rounded-xl bg-blue-50 text-blue-700 flex items-center justify-center flex-shrink-0 font-bold"
            >
              <component :is="getAccountIcon(acc.type)" class="w-5 h-5" />
            </div>
            <div>
              <div class="font-bold text-slate-800 text-sm sm:text-base">{{ acc.name }}</div>
              <div class="text-xs text-slate-500 flex items-center space-x-2">
                <span class="capitalize">{{ acc.type }}</span>
                <span v-if="acc.institution">&bull; {{ acc.institution }}</span>
              </div>
            </div>
          </div>

          <div class="flex items-center justify-between sm:justify-end space-x-3 w-full sm:w-auto">
            <div class="text-left sm:text-right">
              <div class="text-sm sm:text-base font-extrabold text-slate-900">
                {{ props.formatRupiah(acc.balance) }}
              </div>
              <span
                class="text-[10px] uppercase tracking-wider text-blue-600 bg-blue-50 font-bold px-2 py-0.5 rounded-full"
              >
                Aset Pasif
              </span>
            </div>

            <div class="flex items-center space-x-1.5">
              <button
                type="button"
                class="px-3 py-1.5 rounded-lg text-xs font-semibold bg-slate-100 text-slate-700 hover:bg-slate-200 transition flex items-center space-x-1 min-h-[36px]"
                @click="openRevalue(acc)"
              >
                <Edit3 class="w-3.5 h-3.5" />
                <span>Update Nilai</span>
              </button>
              <button
                type="button"
                aria-label="Hapus aset"
                class="p-2 text-slate-400 hover:text-rose-600 rounded-lg hover:bg-rose-50 transition"
                @click="deleteAccount(acc.id)"
              >
                <Trash2 class="w-4 h-4" />
              </button>
            </div>
          </div>
        </div>

        <div
          v-if="accountsData.passive_accounts.length === 0 && !loading"
          class="p-8 text-center text-slate-400 text-xs"
        >
          Belum ada aset pasif / investasi. Klik "Tambah Aset Pasif" (misal: Reksadana Makmur,
          Pluang, Emas).
        </div>
      </div>
    </div>

    <!-- Modal Tambah Akun -->
    <div
      v-if="showAddModal"
      class="fixed inset-0 z-50 bg-slate-900/60 backdrop-blur-sm flex items-center justify-center p-4"
    >
      <div class="bg-white rounded-2xl max-w-md w-full p-6 shadow-xl space-y-4">
        <div class="flex items-center justify-between border-b border-slate-100 pb-3">
          <h3 class="font-bold text-slate-800 text-base">Tambah Akun / Dompet Baru</h3>
          <button
            type="button"
            class="text-slate-400 hover:text-slate-600"
            @click="showAddModal = false"
          >
            <X class="w-5 h-5" />
          </button>
        </div>

        <form class="space-y-4 text-xs sm:text-sm" @submit.prevent="submitAddAccount">
          <!-- Kelompok Akun -->
          <div>
            <span class="block font-semibold text-slate-700 mb-1.5">Kelompok Akun</span>
            <div class="grid grid-cols-2 gap-2">
              <button
                type="button"
                :class="[
                  formAccount.account_group === 'operational'
                    ? 'border-emerald-600 bg-emerald-50 text-emerald-800 font-bold'
                    : 'border-slate-200 text-slate-600 hover:bg-slate-50',
                ]"
                class="p-2.5 rounded-xl border text-center transition"
                @click="formAccount.account_group = 'operational'"
              >
                Kas Operasional
                <span class="block text-[10px] font-normal text-slate-500 mt-0.5"
                  >Uang belanja harian</span
                >
              </button>
              <button
                type="button"
                :class="[
                  formAccount.account_group === 'passive_asset'
                    ? 'border-blue-600 bg-blue-50 text-blue-800 font-bold'
                    : 'border-slate-200 text-slate-600 hover:bg-slate-50',
                ]"
                class="p-2.5 rounded-xl border text-center transition"
                @click="formAccount.account_group = 'passive_asset'"
              >
                Aset Pasif / Investasi
                <span class="block text-[10px] font-normal text-slate-500 mt-0.5"
                  >Tidak ganggu limit belanja</span
                >
              </button>
            </div>
          </div>

          <!-- Nama Akun -->
          <div>
            <label for="account-name" class="block font-semibold text-slate-700 mb-1"
              >Nama Akun</label
            >
            <input
              id="account-name"
              v-model="formAccount.name"
              type="text"
              placeholder="Contoh: BCA Utama / Reksadana Makmur"
              required
              class="w-full px-3.5 py-2.5 rounded-xl border border-slate-200 focus:outline-none focus:ring-2 focus:ring-emerald-500"
            />
          </div>

          <!-- Tipe & Institusi -->
          <div class="grid grid-cols-2 gap-3">
            <div>
              <label for="account-type" class="block font-semibold text-slate-700 mb-1">Tipe</label>
              <select
                id="account-type"
                v-model="formAccount.type"
                class="w-full px-3 py-2.5 rounded-xl border border-slate-200 bg-white focus:outline-none focus:ring-2 focus:ring-emerald-500"
              >
                <option value="bank">Bank</option>
                <option value="ewallet">E-Wallet</option>
                <option value="cash">Tunai / Dompet</option>
                <option value="mutual_fund">Reksadana</option>
                <option value="stocks">Saham</option>
                <option value="gold">Emas</option>
                <option value="crypto">Kripto</option>
                <option value="other">Lainnya</option>
              </select>
            </div>
            <div>
              <label for="account-institution" class="block font-semibold text-slate-700 mb-1"
                >Institusi / Platform</label
              >
              <input
                id="account-institution"
                v-model="formAccount.institution"
                type="text"
                placeholder="Contoh: BCA, Makmur, GoPay"
                class="w-full px-3.5 py-2.5 rounded-xl border border-slate-200 focus:outline-none focus:ring-2 focus:ring-emerald-500"
              />
            </div>
          </div>

          <!-- Saldo Awal -->
          <div>
            <label for="account-balance" class="block font-semibold text-slate-700 mb-1"
              >Saldo Awal (Rp)</label
            >
            <input
              id="account-balance"
              v-model.number="formAccount.balance"
              type="number"
              min="0"
              placeholder="0"
              class="w-full px-3.5 py-2.5 rounded-xl border border-slate-200 focus:outline-none focus:ring-2 focus:ring-emerald-500 font-bold"
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
              class="px-5 py-2 bg-emerald-600 hover:bg-emerald-700 text-white rounded-xl font-bold shadow-sm"
            >
              Simpan Akun
            </button>
          </div>
        </form>
      </div>
    </div>

    <!-- Modal Revaluasi Aset Pasif -->
    <div
      v-if="showRevalueModal && selectedAsset"
      class="fixed inset-0 z-50 bg-slate-900/60 backdrop-blur-sm flex items-center justify-center p-4"
    >
      <div class="bg-white rounded-2xl max-w-md w-full p-6 shadow-xl space-y-4">
        <div class="flex items-center justify-between border-b border-slate-100 pb-3">
          <div>
            <h3 class="font-bold text-slate-800 text-base">Update Nilai / Valuasi Pasar</h3>
            <p class="text-xs text-slate-500">{{ selectedAsset.name }}</p>
          </div>
          <button
            type="button"
            class="text-slate-400 hover:text-slate-600"
            @click="showRevalueModal = false"
          >
            <X class="w-5 h-5" />
          </button>
        </div>

        <form class="space-y-4 text-xs sm:text-sm" @submit.prevent="submitRevalue">
          <div class="p-3 bg-slate-50 rounded-xl space-y-1">
            <span class="text-slate-500 text-xs">Saldo Tercatat Saat Ini:</span>
            <div class="font-bold text-slate-800 text-base">
              {{ props.formatRupiah(selectedAsset.balance) }}
            </div>
          </div>

          <div>
            <label for="revalue-new-balance" class="block font-semibold text-slate-700 mb-1"
              >Nilai Terkini / Saldo Baru (Rp)</label
            >
            <input
              id="revalue-new-balance"
              v-model.number="revalueForm.new_balance"
              type="number"
              min="0"
              required
              class="w-full px-3.5 py-2.5 rounded-xl border border-slate-200 focus:outline-none focus:ring-2 focus:ring-blue-500 font-bold text-base"
            />
          </div>

          <!-- Selisih Pertumbuhan -->
          <div class="text-xs">
            <span class="text-slate-500">Selisih Pertumbuhan: </span>
            <span
              :class="[
                revalueDifference >= 0 ? 'text-emerald-600 font-bold' : 'text-rose-600 font-bold',
              ]"
            >
              {{ revalueDifference >= 0 ? '+' : '' }}{{ props.formatRupiah(revalueDifference) }}
            </span>
          </div>

          <div>
            <label for="revalue-notes" class="block font-semibold text-slate-700 mb-1"
              >Catatan (Opsional)</label
            >
            <input
              id="revalue-notes"
              v-model="revalueForm.notes"
              type="text"
              placeholder="Contoh: NAB akhir bulan"
              class="w-full px-3.5 py-2 rounded-xl border border-slate-200 focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>

          <div class="pt-2 flex items-center justify-end space-x-2">
            <button
              type="button"
              class="px-4 py-2 rounded-xl text-slate-600 hover:bg-slate-100 font-medium"
              @click="showRevalueModal = false"
            >
              Batal
            </button>
            <button
              type="submit"
              class="px-5 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-xl font-bold shadow-sm"
            >
              Simpan Nilai Baru
            </button>
          </div>
        </form>
      </div>
    </div>

    <!-- Modal Pengaturan Siklus Gajian -->
    <div
      v-if="showSettingsModal"
      class="fixed inset-0 z-50 bg-slate-900/60 backdrop-blur-sm flex items-center justify-center p-4"
    >
      <div class="bg-white rounded-2xl max-w-md w-full p-6 shadow-xl space-y-4">
        <div class="flex items-center justify-between border-b border-slate-100 pb-3">
          <h3 class="font-bold text-slate-800 text-base">Atur Siklus Gajian & Pendapatan</h3>
          <button
            type="button"
            class="text-slate-400 hover:text-slate-600"
            @click="showSettingsModal = false"
          >
            <X class="w-5 h-5" />
          </button>
        </div>

        <form class="space-y-4 text-xs sm:text-sm" @submit.prevent="submitSettings">
          <div>
            <label for="setting-payday-date" class="block font-semibold text-slate-700 mb-1"
              >Tanggal Gajian Bulanan (1 - 31)</label
            >
            <input
              id="setting-payday-date"
              v-model.number="settingForm.payday_date"
              type="number"
              min="1"
              max="31"
              required
              class="w-full px-3.5 py-2.5 rounded-xl border border-slate-200 focus:outline-none focus:ring-2 focus:ring-emerald-500 font-bold"
            />
            <p class="text-[11px] text-slate-500 mt-1">
              Tanggal ini menjadi patokan hitung mundur sisa hari (*Next Payday Remain*) dan
              pembagian jatah harian.
            </p>
          </div>

          <div>
            <label for="setting-monthly-income" class="block font-semibold text-slate-700 mb-1"
              >Estimasi Pendapatan per Bulan (Rp)</label
            >
            <input
              id="setting-monthly-income"
              v-model.number="settingForm.monthly_income_budget"
              type="number"
              min="0"
              required
              class="w-full px-3.5 py-2.5 rounded-xl border border-slate-200 focus:outline-none focus:ring-2 focus:ring-emerald-500 font-bold"
            />
            <p class="text-[11px] text-slate-500 mt-1">Digunakan sebagai acuan rasio 50-30-20.</p>
          </div>

          <div class="pt-2 flex items-center justify-end space-x-2">
            <button
              type="button"
              class="px-4 py-2 rounded-xl text-slate-600 hover:bg-slate-100 font-medium"
              @click="showSettingsModal = false"
            >
              Batal
            </button>
            <button
              type="submit"
              class="px-5 py-2 bg-emerald-600 hover:bg-emerald-700 text-white rounded-xl font-bold shadow-sm flex items-center space-x-1.5"
            >
              <Check class="w-4 h-4" />
              <span>Simpan Pengaturan</span>
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>
