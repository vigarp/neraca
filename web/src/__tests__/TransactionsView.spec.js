import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import TransactionsView from '../views/TransactionsView.vue'

const formatRupiah = val =>
  new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    minimumFractionDigits: 0,
    maximumFractionDigits: 0,
  }).format(val || 0)

describe('TransactionsView.vue', () => {
  beforeEach(() => {
    vi.stubGlobal(
      'fetch',
      vi.fn().mockImplementation(url => {
        if (url.includes('/api/accounts')) {
          return Promise.resolve({
            ok: true,
            json: () =>
              Promise.resolve({
                operational_accounts: [
                  { id: 1, name: 'BCA Utama', balance: 5000000, account_group: 'operational' },
                ],
                passive_accounts: [],
                emergency_accounts: [],
              }),
          })
        }
        if (url.includes('/api/categories')) {
          return Promise.resolve({
            ok: true,
            json: () =>
              Promise.resolve({
                needs_categories: [
                  { id: 1, name: 'Makan & Minum', pillar: 'needs', type: 'expense' },
                ],
                wants_categories: [],
                savings_categories: [],
                income_categories: [],
                all_categories: [
                  { id: 1, name: 'Makan & Minum', pillar: 'needs', type: 'expense' },
                ],
              }),
          })
        }
        if (url.includes('/api/transactions')) {
          return Promise.resolve({
            ok: true,
            json: () =>
              Promise.resolve({
                transactions: [
                  {
                    id: 1,
                    account_id: 1,
                    account_name: 'BCA Utama',
                    category_id: 1,
                    category_name: 'Makan & Minum',
                    pillar: 'needs',
                    amount: 35000,
                    type: 'expense',
                    description: 'Makan Siang Nasi Padang',
                    transaction_date: '2026-09-13',
                  },
                ],
                total_expense: 35000,
                total_income: 0,
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

  it('merender judul halaman dan ringkasan transaksi', async () => {
    const wrapper = mount(TransactionsView, {
      props: { formatRupiah },
    })
    await flushPromises()

    expect(wrapper.text()).toContain('Catatan & Mutasi Harian')
    expect(wrapper.text()).toContain('Total Pengeluaran')
    expect(wrapper.text()).toContain('Total Pemasukan')
    expect(wrapper.text()).toContain('Arus Kas Bersih')
  })

  it('merender daftar transaksi dengan pengelompokan tanggal dan badge 50% Needs', async () => {
    const wrapper = mount(TransactionsView, {
      props: { formatRupiah },
    })
    await flushPromises()

    expect(wrapper.text()).toContain('Makan Siang Nasi Padang')
    expect(wrapper.text()).toContain('50% Needs')
    expect(wrapper.text()).toContain('BCA Utama')
    expect(wrapper.text()).toContain('Rp')
  })

  it('dapat membuka modal catat transaksi', async () => {
    const wrapper = mount(TransactionsView, {
      props: { formatRupiah },
    })
    await flushPromises()

    // Cari tombol "Catat Transaksi"
    const addBtn = wrapper.findAll('button').find(b => b.text().includes('Catat Transaksi'))
    expect(addBtn).toBeDefined()
    await addBtn.trigger('click')

    expect(wrapper.text()).toContain('Catat Mutasi Transaksi')
    expect(wrapper.text()).toContain('Jumlah Nominal (Rp)')
  })
})
