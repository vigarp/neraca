import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import App from '../App.vue'

describe('App.vue', () => {
  beforeEach(() => {
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue({
        ok: true,
        json: () =>
          Promise.resolve({
            status: 'ok',
            app: 'neraca',
            database: 'connected',
          }),
      })
    )
  })

  it('merender judul aplikasi dengan benar', async () => {
    const wrapper = mount(App)
    expect(wrapper.text()).toContain('Neraca')
    expect(wrapper.text()).toContain('Personal Financial Dashboard')
  })

  it('memanggil endpoint /api/health saat mounted dan menampilkan status online', async () => {
    const wrapper = mount(App)
    await flushPromises()

    expect(global.fetch).toHaveBeenCalledWith('/api/health')
    expect(wrapper.text()).toContain('Server Online (SQLite WAL)')
  })

  it('merender overview metric card dengan format Rupiah', async () => {
    const wrapper = mount(App)
    expect(wrapper.text()).toContain('Total Saldo (Net Worth)')
    expect(wrapper.text()).toContain('Rp')
  })
})
