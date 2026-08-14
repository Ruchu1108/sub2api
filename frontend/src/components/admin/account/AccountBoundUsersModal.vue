<template>
  <BaseDialog
    :show="show"
    :title="t('admin.accounts.boundUsersTitle', { name: account?.name || '' })"
    width="normal"
    @close="emit('close')"
  >
    <div v-if="account" class="space-y-4">
      <p class="text-sm text-gray-600 dark:text-gray-400">
        {{ t('admin.accounts.boundUsersHint') }}
      </p>
      <input
        v-model="search"
        type="text"
        class="input"
        :placeholder="t('admin.users.searchUsers')"
        @input="debouncedSearch"
      />
      <div class="max-h-72 overflow-y-auto rounded-lg border border-gray-200 dark:border-dark-600">
        <div v-if="loading" class="p-4 text-center text-sm text-gray-400">
          {{ t('common.loading') }}
        </div>
        <div v-else-if="candidates.length === 0" class="p-4 text-center text-sm text-gray-400">
          {{ t('admin.accounts.boundUsersNoResults') }}
        </div>
        <label
          v-for="user in candidates"
          :key="user.id"
          class="flex cursor-pointer items-center gap-3 px-3 py-2.5 transition-colors hover:bg-gray-50 dark:hover:bg-dark-700"
        >
          <input
            v-model="selectedIds"
            type="checkbox"
            :value="user.id"
            class="h-4 w-4 rounded border-gray-300 text-primary-600 focus:ring-primary-500"
          />
          <div class="min-w-0 flex-1">
            <div class="truncate text-sm font-medium text-gray-800 dark:text-gray-200">
              {{ user.username || user.email }}
            </div>
            <div class="truncate text-xs text-gray-500 dark:text-dark-400">
              {{ user.email }} · {{ t('admin.users.columns.defaultAmount') }}: ${{ user.default_amount.toFixed(2) }}
            </div>
          </div>
        </label>
      </div>
      <p class="text-xs text-gray-500 dark:text-dark-400">
        {{ t('admin.accounts.boundUsersSelectedCount', { count: selectedIds.length }) }}
      </p>
    </div>

    <template #footer>
      <div class="flex justify-end gap-3">
        <button type="button" class="btn btn-secondary" @click="emit('close')">
          {{ t('common.cancel') }}
        </button>
        <button
          type="button"
          class="btn btn-primary"
          :disabled="saving || !account"
          data-test="bound-users-save"
          @click="save"
        >
          {{ saving ? t('admin.accounts.boundUsersSaving') : t('common.save') }}
        </button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { adminAPI } from '@/api/admin'
import { useAppStore } from '@/stores/app'
import type { Account } from '@/types'
import type { BoundUserSummary } from '@/api/admin/accounts'
import BaseDialog from '@/components/common/BaseDialog.vue'

const props = defineProps<{ show: boolean; account: Account | null }>()
const emit = defineEmits<{ close: []; success: [] }>()
const { t } = useI18n()
const appStore = useAppStore()

const search = ref('')
const loading = ref(false)
const saving = ref(false)
const candidates = ref<BoundUserSummary[]>([])
const selectedIds = ref<number[]>([])
let searchTimer: ReturnType<typeof setTimeout> | undefined

const loadCandidates = async () => {
  if (!props.account) return
  loading.value = true
  try {
    const response = await adminAPI.users.list(1, 100, {
      search: search.value.trim() || undefined,
      sort_by: 'created_at',
      sort_order: 'desc'
    })
    candidates.value = response.items.map((u) => ({
      id: u.id,
      email: u.email,
      username: u.username,
      balance: u.balance,
      default_amount: u.default_amount
    }))
  } catch (error: any) {
    appStore.showError(error?.message || error?.response?.data?.detail || t('admin.accounts.boundUsersLoadFailed'))
    candidates.value = []
  } finally {
    loading.value = false
  }
}

const loadBound = async () => {
  if (!props.account) return
  try {
    selectedIds.value = (await adminAPI.accounts.getBoundUsers(props.account.id)).map((u) => u.id)
  } catch (error: any) {
    appStore.showError(error?.message || error?.response?.data?.detail || t('admin.accounts.boundUsersLoadFailed'))
    selectedIds.value = []
  }
}

const debouncedSearch = () => {
  if (searchTimer) clearTimeout(searchTimer)
  searchTimer = setTimeout(() => void loadCandidates(), 300)
}

const save = async () => {
  if (!props.account || saving.value) return
  saving.value = true
  try {
    const result = await adminAPI.accounts.setBoundUsers(props.account.id, [...selectedIds.value])
    appStore.showSuccess(t('admin.accounts.boundUsersSaved', { count: result.bound_count }))
    emit('success')
    emit('close')
  } catch (error: any) {
    appStore.showError(error?.message || error?.response?.data?.detail || t('admin.accounts.boundUsersSaveFailed'))
  } finally {
    saving.value = false
  }
}

watch(() => props.show, (visible) => {
  if (!visible) return
  search.value = ''
  candidates.value = []
  selectedIds.value = []
  void loadBound()
  void loadCandidates()
})
</script>
