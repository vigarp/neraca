<script setup>
import { ref } from 'vue'
import { PiggyBank, ShieldCheck, Lock, User, AlertCircle, Sparkles } from '@lucide/vue'

const emit = defineEmits(['setup-success'])

const username = ref('')
const password = ref('')
const confirmPassword = ref('')
const loading = ref(false)
const errorMessage = ref('')

const handleSetup = async () => {
  if (!username.value.trim() || !password.value) {
    errorMessage.value = 'Username dan password wajib diisi'
    return
  }

  if (username.value.trim().length < 3) {
    errorMessage.value = 'Username minimal 3 karakter'
    return
  }

  if (password.value.length < 6) {
    errorMessage.value = 'Password minimal 6 karakter'
    return
  }

  if (password.value !== confirmPassword.value) {
    errorMessage.value = 'Konfirmasi password tidak cocok'
    return
  }

  loading.value = true
  errorMessage.value = ''

  try {
    const res = await fetch('/api/auth/setup', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        username: username.value.trim(),
        password: password.value,
      }),
    })

    if (!res.ok) {
      const errText = await res.text()
      throw new Error(errText || `Gagal inisialisasi akun (HTTP ${res.status})`)
    }

    const data = await res.json()
    emit('setup-success', data.username || username.value)
  } catch (err) {
    errorMessage.value = err.message || 'Terjadi kesalahan sistem'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="min-h-screen bg-slate-100 flex flex-col justify-center items-center px-4 py-8">
    <div class="w-full max-w-md space-y-6">
      <!-- App Header & Logo -->
      <div class="text-center space-y-2">
        <div
          class="w-14 h-14 mx-auto rounded-2xl bg-emerald-600 flex items-center justify-center text-white shadow-lg shadow-emerald-200"
        >
          <PiggyBank class="w-8 h-8" />
        </div>
        <h1 class="text-2xl font-extrabold tracking-tight text-slate-900">Selamat Datang di Neraca</h1>
        <p class="text-xs sm:text-sm text-slate-500 font-medium">
          Inisialisasi Akun Pengelola Utama
        </p>
      </div>

      <!-- Setup Card -->
      <div class="bg-white p-6 sm:p-8 rounded-2xl border border-slate-200 shadow-sm space-y-5">
        <div class="p-3.5 rounded-xl bg-emerald-50 border border-emerald-100 text-emerald-800 text-xs leading-relaxed flex items-start space-x-2.5">
          <ShieldCheck class="w-4 h-4 text-emerald-600 flex-shrink-0 mt-0.5" />
          <span>
            Karena ini pertama kali dibuka, buat akun master Anda. Setelah akun dibuat, pendaftaran
            publik akan ditutup permanen.
          </span>
        </div>

        <!-- Error Alert -->
        <div
          v-if="errorMessage"
          class="p-3.5 rounded-xl bg-rose-50 border border-rose-200 text-rose-700 text-xs flex items-center space-x-2"
        >
          <AlertCircle class="w-4 h-4 flex-shrink-0" />
          <span>{{ errorMessage }}</span>
        </div>

        <form class="space-y-4" @submit.prevent="handleSetup">
          <!-- Username Input -->
          <div class="space-y-1.5">
            <label for="setup-username" class="block text-xs font-bold text-slate-700">
              Username Pengelola
            </label>
            <div class="relative rounded-xl shadow-sm">
              <div class="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none text-slate-400">
                <User class="w-4 h-4" />
              </div>
              <input
                id="setup-username"
                v-model="username"
                type="text"
                required
                autocomplete="username"
                placeholder="Contoh: admin atau nama Anda"
                class="w-full pl-9 pr-4 py-2.5 text-sm rounded-xl border border-slate-300 focus:outline-none focus:ring-2 focus:ring-emerald-500 focus:border-emerald-500 min-h-[44px]"
              />
            </div>
          </div>

          <!-- Password Input -->
          <div class="space-y-1.5">
            <label for="setup-password" class="block text-xs font-bold text-slate-700">
              Password (Minimal 6 Karakter)
            </label>
            <div class="relative rounded-xl shadow-sm">
              <div class="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none text-slate-400">
                <Lock class="w-4 h-4" />
              </div>
              <input
                id="setup-password"
                v-model="password"
                type="password"
                required
                autocomplete="new-password"
                placeholder="Buat password yang kuat"
                class="w-full pl-9 pr-4 py-2.5 text-sm rounded-xl border border-slate-300 focus:outline-none focus:ring-2 focus:ring-emerald-500 focus:border-emerald-500 min-h-[44px]"
              />
            </div>
          </div>

          <!-- Confirm Password Input -->
          <div class="space-y-1.5">
            <label for="setup-confirm-password" class="block text-xs font-bold text-slate-700">
              Konfirmasi Password
            </label>
            <div class="relative rounded-xl shadow-sm">
              <div class="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none text-slate-400">
                <Lock class="w-4 h-4" />
              </div>
              <input
                id="setup-confirm-password"
                v-model="confirmPassword"
                type="password"
                required
                autocomplete="new-password"
                placeholder="Ulangi password"
                class="w-full pl-9 pr-4 py-2.5 text-sm rounded-xl border border-slate-300 focus:outline-none focus:ring-2 focus:ring-emerald-500 focus:border-emerald-500 min-h-[44px]"
              />
            </div>
          </div>

          <!-- Submit Button -->
          <button
            type="submit"
            :disabled="loading"
            class="w-full py-2.5 px-4 rounded-xl bg-emerald-600 hover:bg-emerald-700 text-white font-bold text-sm shadow-md shadow-emerald-200 transition flex items-center justify-center space-x-2 min-h-[44px] disabled:opacity-60"
          >
            <Sparkles class="w-4 h-4" />
            <span>{{ loading ? 'Menginisialisasi...' : 'Mulai Gunakan Neraca' }}</span>
          </button>
        </form>
      </div>

      <!-- Footer Info -->
      <p class="text-center text-[11px] text-slate-400">
        Password di-hash dengan standar industri Bcrypt.
      </p>
    </div>
  </div>
</template>

