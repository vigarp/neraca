import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import App from '../App.vue'

describe('App.vue', () => {
  beforeEach(() => {
    vi.stubGlobal(
      'fetch',
      vi.fn().mockImplementation(url => {
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
    expect(wrapper.text()).toContain('Neraca')
    expect(wrapper.text()).toContain('Personal Financial Dashboard')
  })

  it('merender navigasi mobile dan desktop dengan tab lengkap', async () => {
    const wrapper = mount(App)
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
    expect(wrapper.text()).toContain('Total Net Worth')
    expect(wrapper.text()).toContain('Rp')
  })

  it('merender metrik hitung mundur gajian dan prospect daily limit', async () => {
    const wrapper = mount(App)
    expect(wrapper.text()).toContain('Next Payday Remain')
    expect(wrapper.text()).toContain('Prospect Daily Limit')
    expect(wrapper.text()).toContain('Average Daily Expense')
    expect(wrapper.text()).toContain('Distribusi Pengeluaran Siklus Berjalan (50-30-20)')
  })
})
