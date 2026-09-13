import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import BudgetsView from '../views/BudgetsView.vue'

describe('BudgetsView.vue', () => {
  const formatRupiah = val => `Rp ${Number(val || 0).toLocaleString('id-ID')}`

  const mockBurnRate = {
    next_payday_remain: 12,
    next_payday_date: '2026-09-25',
    cycle_start_date: '2026-08-25',
    days_elapsed: 19,
    remain_funds: 3000000,
    grand_total_expenses: 1773000,
    cycle_income: 6000000,
    prospect_daily_limit: 250000,
    average_daily_expense: 93315.79,
    burn_rate_status: 'safe',
    pillar_breakdown: {
      needs: { total: 1000000, percentage: 56.4 },
      wants: { total: 500000, percentage: 28.2 },
      savings: { total: 273000, percentage: 15.4 },
    },
    category_breakdown: [
      {
        id: 1,
        name: 'Makan Siang & Rutin',
        pillar: 'needs',
        total: 1000000,
        percentage: 56.4,
        transaction_count: 20,
      },
      {
        id: 7,
        name: 'Kopi & Hiburan',
        pillar: 'wants',
        total: 500000,
        percentage: 28.2,
        transaction_count: 15,
      },
      {
        id: 12,
        name: 'Reksadana Makmur',
        pillar: 'savings',
        total: 273000,
        percentage: 15.4,
        transaction_count: 2,
      },
    ],
  }

  it('merender target plafon 50-30-20 dari income siklus berjalan', () => {
    const wrapper = mount(BudgetsView, {
      props: {
        burnRate: mockBurnRate,
        formatRupiah,
      },
    })

    expect(wrapper.text()).toContain('Alokasi Anggaran 50-30-20')
    // Income = 6.000.000
    // Needs 50% = 3.000.000
    // Wants 30% = 1.800.000
    // Savings 20% = 1.200.000
    expect(wrapper.text()).toContain('50% Kebutuhan Pokok')
    expect(wrapper.text()).toContain('Rp 3.000.000')
    expect(wrapper.text()).toContain('30% Kebutuhan Pribadi')
    expect(wrapper.text()).toContain('Rp 1.800.000')
    expect(wrapper.text()).toContain('20% Tabungan & Investasi')
    expect(wrapper.text()).toContain('Rp 1.200.000')

    // Total Sisa Dana = 2.000.000 + 1.300.000 + 927.000 = 4.227.000
    expect(wrapper.text()).toContain('Total Sisa Dana (Remain Funds)')
    expect(wrapper.text()).toContain('Rp 4.227.000')
  })

  it('merender 3 kolom rincian per pilar ala spreadsheet beserta sisa dana', () => {
    const wrapper = mount(BudgetsView, {
      props: {
        burnRate: mockBurnRate,
        formatRupiah,
      },
    })

    // Sisa Dana Pokok = 3.000.000 - 1.000.000 = 2.000.000
    // Sisa Dana Pribadi = 1.800.000 - 500.000 = 1.300.000
    // Sisa Target Tabungan = 1.200.000 - 273.000 = 927.000
    expect(wrapper.text()).toContain('Rincian Pengeluaran per Pilar (Spreadsheet Format)')
    expect(wrapper.text()).toContain('Makan Siang & Rutin')
    expect(wrapper.text()).toContain('Kopi & Hiburan')
    expect(wrapper.text()).toContain('Reksadana Makmur')
    expect(wrapper.text()).toContain('Rp 2.000.000')
    expect(wrapper.text()).toContain('Rp 1.300.000')
    expect(wrapper.text()).toContain('Rp 927.000')
  })

  it('merender tabel rekapitulasi kategori lengkap dengan Qty transaksi', () => {
    const wrapper = mount(BudgetsView, {
      props: {
        burnRate: mockBurnRate,
        formatRupiah,
      },
    })

    expect(wrapper.text()).toContain('Rekapitulasi Pengeluaran per Kategori')
    expect(wrapper.text()).toContain('Total Transaksi: 37')
    expect(wrapper.text()).toContain('Grand Total Pengeluaran')
    expect(wrapper.text()).toContain('Rp 1.773.000')
  })

  it('mendeteksi status overbudget jika realisasi melebihi target', () => {
    const overbudgetBurnRate = {
      ...mockBurnRate,
      pillar_breakdown: {
        needs: { total: 3500000, percentage: 70 }, // > target 3.000.000
        wants: { total: 1000000, percentage: 20 },
        savings: { total: 500000, percentage: 10 },
      },
    }

    const wrapper = mount(BudgetsView, {
      props: {
        burnRate: overbudgetBurnRate,
        formatRupiah,
      },
    })

    expect(wrapper.text()).toContain('Overbudget')
  })

  it('mengirimkan emit refresh saat tombol muat ulang diklik', async () => {
    const wrapper = mount(BudgetsView, {
      props: {
        burnRate: mockBurnRate,
        formatRupiah,
      },
    })

    const refreshBtn = wrapper.find('button[aria-label="Muat ulang data anggaran"]')
    expect(refreshBtn.exists()).toBe(true)
    await refreshBtn.trigger('click')

    expect(wrapper.emitted('refresh')).toBeTruthy()
  })
})
