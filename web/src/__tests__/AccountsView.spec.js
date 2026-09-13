import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import AccountsView from '../views/AccountsView.vue'

describe('AccountsView.vue', () => {
  const formatRupiah = val => `Rp ${Number(val || 0).toLocaleString('id-ID')}`

  beforeEach(() => {
    vi.stubGlobal(
      'fetch',
      vi.fn().mockImplementation(url => {
        if (url === '/api/accounts') {
          return Promise.resolve({
            ok: true,
            json: () =>
              Promise.resolve({
                operational_accounts: [
                  {
                    id: 1,
                    name: 'BCA Utama',
                    account_group: 'operational',
                    type: 'bank',
                    balance: 4800000,
                    institution: 'BCA',
                  },
                ],
                passive_accounts: [
                  {
                    id: 2,
                    name: 'Reksadana Makmur',
                    account_group: 'passive_asset',
                    type: 'mutual_fund',
                    balance: 15000000,
                    institution: 'Makmur',
                  },
                ],
                emergency_accounts: [],
                total_operational_balance: 4800000,
                total_passive_assets: 15000000,
                total_emergency_balance: 0,
                total_net_worth: 19800000,
              }),
          })
        }
        if (url === '/api/settings') {
          return Promise.resolve({
            ok: true,
            json: () =>
              Promise.resolve({
                payday_date: 25,
                monthly_income_budget: 6000000,
              }),
          })
        }
        return Promise.resolve({ ok: true, json: () => Promise.resolve({}) })
      })
    )
  })

  it('merender ringkasan Net Worth, Kas Operasional, dan Aset Pasif', async () => {
    const wrapper = mount(AccountsView, {
      props: { formatRupiah },
    })
    await flushPromises()

    expect(wrapper.text()).toContain('Total Kekayaan Bersih (Net Worth)')
    expect(wrapper.text()).toContain('Rp 19.800.000')
    expect(wrapper.text()).toContain('BCA Utama')
    expect(wrapper.text()).toContain('Reksadana Makmur')
    expect(wrapper.text()).toContain('Tiap Tgl 25')
  })

  it('menyediakan tombol Update Nilai untuk aset pasif', async () => {
    const wrapper = mount(AccountsView, {
      props: { formatRupiah },
    })
    await flushPromises()

    const updateButtons = wrapper.findAll('button').filter(b => b.text().includes('Update Nilai'))
    expect(updateButtons.length).toBeGreaterThan(0)
  })
})
