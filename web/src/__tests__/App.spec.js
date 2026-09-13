import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import App from '../App.vue'

describe('App.vue', () => {
  beforeEach(() => {
    vi.stubGlobal(
      'fetch',
      vi.fn().mockImplementation(url => {
        if (url.includes('/api/auth/status')) {
          return Promise.resolve({
            ok: true,
            json: () =>
              Promise.resolve({
                initialized: true,
                authenticated: true,
                username: 'vigarp',
              }),
          })
        }
        if (url.includes('/api/health')) {
          return Promise.resolve({
            ok: true,
            json: () =>
              Promise.resolve({
                status: 'ok',
                app: 'neraca',
                database: 'connected',
              }),
          })
        }
        if (url.includes('/api/accounts')) {
          return Promise.resolve({
            ok: true,
            json: () =>
              Promise.resolve({
                total_net_worth: 20000000,
                operational_accounts: [],
                passive_accounts: [],
              }),
          })
        }
        if (url.includes('/api/transactions')) {
          return Promise.resolve({
            ok: true,
            json: () =>
              Promise.resolve({
                total_expense: 500000,
                total_income: 6000000,
              }),
          })
        }
        if (url.includes('/api/analytics/burn-rate')) {
          return Promise.resolve({
            ok: true,
            json: () =>
              Promise.resolve({
                next_payday_remain: 12,
                next_payday_date: '2026-09-25',
                cycle_start_date: '2026-08-25',
                days_elapsed: 19,
                remain_funds: 4300000,
                grand_total_expenses: 500000,
                prospect_daily_limit: 358333.33,
                average_daily_expense: 26315.79,
                burn_rate_status: 'safe',
                pillar_breakdown: {
                  needs: { total: 300000, percentage: 60 },
                  wants: { total: 200000, percentage: 40 },
                  savings: { total: 0, percentage: 0 },
                },
                category_breakdown: [],
              }),
          })
        }
        return Promise.resolve({
          ok: true,
          json: () => Promise.resolve({}),
        })
      })
    )
  })

  it('merender judul aplikasi dengan benar', async () => {
    const wrapper = mount(App)
    await flushPromises()
    expect(wrapper.text()).toContain('Neraca')
    expect(wrapper.text()).toContain('Personal Financial Dashboard')
  })

  it('merender navigasi mobile dan desktop dengan tab lengkap', async () => {
    const wrapper = mount(App)
    await flushPromises()
    expect(wrapper.text()).toContain('Dashboard')
    expect(wrapper.text()).toContain('Transaksi')
    expect(wrapper.text()).toContain('Dompet')
    expect(wrapper.text()).toContain('Anggaran')
  })

  it('memanggil endpoint /api/health saat mounted dan menampilkan status online', async () => {
    const wrapper = mount(App)
    await flushPromises()

    expect(global.fetch).toHaveBeenCalledWith('/api/health')
    expect(wrapper.text()).toContain('Online')
  })

  it('merender overview metric card dengan format Rupiah', async () => {
    const wrapper = mount(App)
    await flushPromises()
    expect(wrapper.text()).toContain('Total Net Worth')
    expect(wrapper.text()).toContain('Rp')
  })

  it('merender metrik hitung mundur gajian dan prospect daily limit', async () => {
    const wrapper = mount(App)
    await flushPromises()
    expect(wrapper.text()).toContain('Next Payday Remain')
    expect(wrapper.text()).toContain('Prospect Daily Limit')
    expect(wrapper.text()).toContain('Average Daily Expense')
    expect(wrapper.text()).toContain('Distribusi Pengeluaran Siklus Berjalan (50-30-20)')
  })

  it('dapat membuka dropdown menu pengguna dan memicu modal konfirmasi logout', async () => {
    const wrapper = mount(App)
    await flushPromises()

    const userMenuBtn = wrapper.find('button[aria-label="Buka menu pengguna"]')
    expect(userMenuBtn.exists()).toBe(true)
    await userMenuBtn.trigger('click')

    expect(wrapper.text()).toContain('Unduh Cadangan Database')
    expect(wrapper.text()).toContain('Reset Data Keuangan')
    expect(wrapper.text()).toContain('Keluar (Logout)')

    // Klik tombol Keluar
    const logoutBtn = wrapper.findAll('button').find(b => b.text().includes('Keluar (Logout)'))
    expect(logoutBtn).toBeDefined()
    await logoutBtn.trigger('click')

    expect(wrapper.text()).toContain('Konfirmasi Keluar')
    expect(wrapper.text()).toContain('Apakah Anda yakin ingin keluar dari Neraca?')
  })

  it('dapat memicu unduh cadangan database dari dropdown menu', async () => {
    const wrapper = mount(App)
    await flushPromises()

    const userMenuBtn = wrapper.find('button[aria-label="Buka menu pengguna"]')
    await userMenuBtn.trigger('click')

    const downloadBtn = wrapper
      .findAll('button')
      .find(b => b.text().includes('Unduh Cadangan Database'))
    expect(downloadBtn).toBeDefined()

    const clickSpy = vi.fn()
    const origCreateElement = document.createElement.bind(document)
    vi.spyOn(document, 'createElement').mockImplementation(tagName => {
      const el = origCreateElement(tagName)
      if (tagName === 'a') {
        el.click = clickSpy
      }
      return el
    })

    await downloadBtn.trigger('click')
    expect(clickSpy).toHaveBeenCalled()
  })

  it('dapat membuka modal konfirmasi reset data dan memanggil API reset', async () => {
    const wrapper = mount(App)
    await flushPromises()

    const userMenuBtn = wrapper.find('button[aria-label="Buka menu pengguna"]')
    await userMenuBtn.trigger('click')

    const resetBtn = wrapper.findAll('button').find(b => b.text().includes('Reset Data Keuangan'))
    expect(resetBtn).toBeDefined()
    await resetBtn.trigger('click')

    expect(wrapper.text()).toContain('Reset Seluruh Data Keuangan?')
    expect(wrapper.text()).toContain('Akun Login Anda Tetap Aman')

    // Klik Ya, Hapus & Reset Data
    const confirmResetBtn = wrapper
      .findAll('button')
      .find(b => b.text().includes('Ya, Hapus & Reset Data'))
    expect(confirmResetBtn).toBeDefined()
    await confirmResetBtn.trigger('click')
    await flushPromises()

    expect(global.fetch).toHaveBeenCalledWith(
      '/api/settings/reset-data',
      expect.objectContaining({ method: 'POST' })
    )
  })
})
