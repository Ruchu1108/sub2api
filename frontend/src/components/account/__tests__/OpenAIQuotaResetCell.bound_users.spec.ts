import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import OpenAIQuotaResetCell from '../OpenAIQuotaResetCell.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import type { Account } from '@/types'
import { refreshOpenAIQuota, resetOpenAIQuota } from '@/api/admin/accounts'

vi.mock('@/api/admin/accounts', () => ({
  refreshOpenAIQuota: vi.fn(),
  resetOpenAIQuota: vi.fn(),
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string, params?: Record<string, unknown>) => {
        if (params?.windows !== undefined && params?.boundUsers !== undefined) {
          return `${key}:${params.windows}:${params.boundUsers}`
        }
        if (params?.windows !== undefined) return `${key}:${params.windows}`
        return key
      },
    }),
  }
})

const FUTURE_EXPIRY = '2099-07-03T04:05:06Z'

function makeAccount(overrides: Partial<Account>): Account {
  return {
    id: 1,
    name: 'acc',
    platform: 'openai',
    type: 'oauth',
    proxy_id: null,
    concurrency: 3,
    priority: 50,
    status: 'active',
    error_message: null,
    last_used_at: null,
    expires_at: null,
    auto_pause_on_expired: false,
    created_at: '2026-01-01T00:00:00Z',
    updated_at: '2026-01-01T00:00:00Z',
    schedulable: true,
    rate_limited_at: null,
    rate_limit_reset_at: null,
    overload_until: null,
    temp_unschedulable_until: null,
    temp_unschedulable_reason: null,
    session_window_start: null,
    session_window_end: null,
    session_window_status: null,
    ...overrides,
  }
}

const resetButton = (wrapper: ReturnType<typeof mount>) => wrapper.findAll('button')[1]

beforeEach(() => {
  vi.mocked(refreshOpenAIQuota).mockReset()
  vi.mocked(resetOpenAIQuota).mockReset()
})

describe('OpenAIQuotaResetCell — 绑定用户联动提示', () => {
  it('重置成功后响应带 bound_users_reset 时显示联动重置提示', async () => {
    const recoveredAccount = makeAccount({ status: 'active' })
    vi.mocked(resetOpenAIQuota).mockResolvedValue({
      code: 'success', windows_reset: 2, cache_refreshed: true,
      account_state_recovered: true, bound_users_reset: 3,
      quota: { rate_limit_reset_credits: { available_count: 0, credits: [] }, fetched_at: 1770000000 },
      account: recoveredAccount,
    })
    const account = makeAccount({
      extra: { codex_reset_credit_snapshot: { available_count: 1, credits: [{ expires_at: FUTURE_EXPIRY }] } },
    })
    const wrapper = mount(OpenAIQuotaResetCell, { props: { account } })

    await resetButton(wrapper).trigger('click')
    wrapper.findComponent(ConfirmDialog).vm.$emit('confirm')
    await flushPromises()

    expect(resetOpenAIQuota).toHaveBeenCalledWith(1)
    expect(wrapper.text()).toContain('admin.accounts.openaiQuotaReset.resetSuccessWithBoundUsers:2:3')
    expect(wrapper.text()).not.toContain('admin.accounts.openaiQuotaReset.resetSuccess:')
    wrapper.unmount()
  })

  it('重置成功但无绑定用户时保持原有成功提示', async () => {
    const recoveredAccount = makeAccount({ status: 'active' })
    vi.mocked(resetOpenAIQuota).mockResolvedValue({
      code: 'success', windows_reset: 1, cache_refreshed: true,
      account_state_recovered: true, bound_users_reset: 0,
      quota: { rate_limit_reset_credits: { available_count: 0, credits: [] }, fetched_at: 1770000000 },
      account: recoveredAccount,
    })
    const account = makeAccount({
      extra: { codex_reset_credit_snapshot: { available_count: 1, credits: [{ expires_at: FUTURE_EXPIRY }] } },
    })
    const wrapper = mount(OpenAIQuotaResetCell, { props: { account } })

    await resetButton(wrapper).trigger('click')
    wrapper.findComponent(ConfirmDialog).vm.$emit('confirm')
    await flushPromises()

    expect(wrapper.text()).toContain('admin.accounts.openaiQuotaReset.resetSuccess:1')
    expect(wrapper.text()).not.toContain('resetSuccessWithBoundUsers')
    wrapper.unmount()
  })
})
