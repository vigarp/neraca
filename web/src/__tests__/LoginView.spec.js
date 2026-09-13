import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import LoginView from '../views/LoginView.vue'

describe('LoginView.vue', () => {
  beforeEach(() => {
    vi.stubGlobal('fetch', vi.fn())
  })

  it('merender form login dengan username, password, dan checkbox ingat saya', () => {
    const wrapper = mount(LoginView)

    expect(wrapper.text()).toContain('Masuk ke Dashboard Finansial Pribadi Anda')
    expect(wrapper.find('input#login-username').exists()).toBe(true)
    expect(wrapper.find('input#login-password').exists()).toBe(true)
    expect(wrapper.find('input#remember-me').exists()).toBe(true)
  })

  it('menampilkan error jika kredensial salah (401)', async () => {
    global.fetch.mockResolvedValueOnce({
      ok: false,
      status: 401,
      json: () => Promise.resolve({ error: 'Unauthorized' }),
    })

    const wrapper = mount(LoginView)
    await wrapper.find('input#login-username').setValue('vigarp')
    await wrapper.find('input#login-password').setValue('wrongpassword')
    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    expect(wrapper.text()).toContain('Username atau password salah')
  })

  it('memancarkan event login-success saat kredensial valid', async () => {
    global.fetch.mockResolvedValueOnce({
      ok: true,
      status: 200,
      json: () => Promise.resolve({ initialized: true, authenticated: true, username: 'vigarp' }),
    })

    const wrapper = mount(LoginView)
    await wrapper.find('input#login-username').setValue('vigarp')
    await wrapper.find('input#login-password').setValue('secret123')
    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    expect(wrapper.emitted('login-success')).toBeTruthy()
    expect(wrapper.emitted('login-success')[0]).toEqual(['vigarp'])
  })
})
