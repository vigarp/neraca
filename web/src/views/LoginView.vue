<script setup>
import { ref } from 'vue'
import { PiggyBank, Lock, User, Eye, EyeOff, AlertCircle, LogIn } from '@lucide/vue'

const emit = defineEmits(['login-success'])

const username = ref('')
const password = ref('')
const rememberMe = ref(true)
const showPassword = ref(false)
const loading = ref(false)
const errorMessage = ref('')

const handleLogin = async () => {
  if (!username.value.trim() || !password.value) {
    errorMessage.value = 'Username dan password wajib diisi'
    return
  }

  loading.value = true
  errorMessage.value = ''

  try {
    const res = await fetch('/api/auth/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        username: username.value.trim(),
        password: password.value,
        remember_me: rememberMe.value,
      }),
    })

    if (!res.ok) {
      if (res.status === 401) {
        throw new Error('Username atau password salah')
      }
      if (res.status === 403) {
        throw new Error('Verifikasi Cloudflare diperlukan. Silakan muat ulang halaman.')
      }
      if (res.status === 429) {
        throw new Error('Terlalu banyak percobaan login gagal. Silakan tunggu 1 menit.')
      }
      throw new Error(`Gagal login (HTTP ${res.status})`)
    }

    const data = await res.json()
    emit('login-success', data.username || username.value)
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
        <h1 class="text-2xl font-extrabold tracking-tight text-slate-900">Neraca</h1>
        <p class="text-xs sm:text-sm text-slate-500 font-medium">
          Masuk ke Dashboard Finansial Pribadi Anda
        </p>
      </div>

      <!-- Login Card -->
      <div class="bg-white p-6 sm:p-8 rounded-2xl border border-slate-200 shadow-sm space-y-5">
        <!-- Error Alert -->
        <div
          v-if="errorMessage"
          class="p-3.5 rounded-xl bg-rose-50 border border-rose-200 text-rose-700 text-xs flex items-center space-x-2"
        >
          <AlertCircle class="w-4 h-4 flex-shrink-0" />
          <span>{{ errorMessage }}</span>
        </div>

        <form class="space-y-4" @submit.prevent="handleLogin">
          <!-- Username Input -->
          <div class="space-y-1.5">
            <label for="login-username" class="block text-xs font-bold text-slate-700">
              Username Pengelola
            </label>
            <div class="relative rounded-xl shadow-sm">
              <div
                class="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none text-slate-400"
              >
                <User class="w-4 h-4" />
              </div>
              <input
                id="login-username"
                v-model="username"
                type="text"
                required
                autocomplete="username"
                placeholder="Masukkan username"
                class="w-full pl-9 pr-4 py-2.5 text-sm rounded-xl border border-slate-300 focus:outline-none focus:ring-2 focus:ring-emerald-500 focus:border-emerald-500 min-h-[44px]"
              />
            </div>
          </div>

          <!-- Password Input -->
          <div class="space-y-1.5">
            <label for="login-password" class="block text-xs font-bold text-slate-700">
              Password
            </label>
            <div class="relative rounded-xl shadow-sm">
              <div
                class="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none text-slate-400"
              >
                <Lock class="w-4 h-4" />
              </div>
              <input
                id="login-password"
                v-model="password"
                :type="showPassword ? 'text' : 'password'"
                required
                autocomplete="current-password"
                placeholder="Masukkan password"
                class="w-full pl-9 pr-11 py-2.5 text-sm rounded-xl border border-slate-300 focus:outline-none focus:ring-2 focus:ring-emerald-500 focus:border-emerald-500 min-h-[44px]"
              />
              <button
                type="button"
                aria-label="Tampilkan atau sembunyikan password"
                class="absolute inset-y-0 right-0 pr-3 flex items-center text-slate-400 hover:text-slate-600 min-w-[44px] justify-center"
                @click="showPassword = !showPassword"
              >
                <component :is="showPassword ? EyeOff : Eye" class="w-4 h-4" />
              </button>
            </div>
          </div>

          <!-- Remember Me Checkbox -->
          <div class="flex items-center justify-between pt-1">
            <div class="flex items-center space-x-2">
              <input
                id="remember-me"
                v-model="rememberMe"
                type="checkbox"
                class="w-4 h-4 text-emerald-600 rounded border-slate-300 focus:ring-emerald-500"
              />
              <label for="remember-me" class="text-xs text-slate-600 select-none cursor-pointer">
                Ingat saya di perangkat ini (30 hari)
              </label>
            </div>
          </div>

          <!-- Submit Button -->
          <button
            type="submit"
            :disabled="loading"
            class="w-full py-2.5 px-4 rounded-xl bg-emerald-600 hover:bg-emerald-700 text-white font-bold text-sm shadow-md shadow-emerald-200 transition flex items-center justify-center space-x-2 min-h-[44px] disabled:opacity-60"
          >
            <LogIn class="w-4 h-4" />
            <span>{{ loading ? 'Memverifikasi...' : 'Masuk ke Neraca' }}</span>
          </button>
        </form>
      </div>

      <!-- Footer Info -->
      <p class="text-center text-[11px] text-slate-400">
        Keamanan terenkripsi dengan Session HttpOnly Cookie (Kebal XSS).
      </p>
    </div>
  </div>
</template>
