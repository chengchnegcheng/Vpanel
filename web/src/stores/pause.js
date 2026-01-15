import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { 
  getPauseStatus, 
  pauseSubscription, 
  resumeSubscription, 
  getPauseHistory 
} from '@/api/modules/pause'

export const usePauseStore = defineStore('pause', () => {
  // State
  const status = ref(null)
  const history = ref([])
  const historyTotal = ref(0)
  const loading = ref(false)
  const error = ref(null)

  // Computed
  const isPaused = computed(() => status.value?.is_paused || false)
  const canPause = computed(() => status.value?.can_pause || false)
  const cannotPauseReason = computed(() => status.value?.cannot_pause_reason || '')
  const remainingPauses = computed(() => status.value?.remaining_pauses || 0)
  const maxDuration = computed(() => status.value?.max_duration_days || 30)
  const activePause = computed(() => status.value?.pause || null)
  const autoResumeAt = computed(() => {
    if (activePause.value?.auto_resume_at) {
      return new Date(activePause.value.auto_resume_at)
    }
    return null
  })

  // Actions
  async function fetchPauseStatus() {
    loading.value = true
    error.value = null
    try {
      const response = await getPauseStatus()
      status.value = response.data || response
      return status.value
    } catch (err) {
      error.value = err.message || 'Failed to fetch pause status'
      throw err
    } finally {
      loading.value = false
    }
  }

  async function pause() {
    loading.value = true
    error.value = null
    try {
      const response = await pauseSubscription()
      // Refresh status after pausing
      await fetchPauseStatus()
      return response.data || response
    } catch (err) {
      error.value = err.message || 'Failed to pause subscription'
      throw err
    } finally {
      loading.value = false
    }
  }

  async function resume() {
    loading.value = true
    error.value = null
    try {
      const response = await resumeSubscription()
      // Refresh status after resuming
      await fetchPauseStatus()
      return response.data || response
    } catch (err) {
      error.value = err.message || 'Failed to resume subscription'
      throw err
    } finally {
      loading.value = false
    }
  }

  async function fetchHistory(page = 1, pageSize = 10) {
    loading.value = true
    error.value = null
    try {
      const response = await getPauseHistory({ page, page_size: pageSize })
      const data = response.data || response
      history.value = data.pauses || []
      historyTotal.value = data.total || 0
      return data
    } catch (err) {
      error.value = err.message || 'Failed to fetch pause history'
      throw err
    } finally {
      loading.value = false
    }
  }

  function reset() {
    status.value = null
    history.value = []
    historyTotal.value = 0
    loading.value = false
    error.value = null
  }

  return {
    // State
    status,
    history,
    historyTotal,
    loading,
    error,
    // Computed
    isPaused,
    canPause,
    cannotPauseReason,
    remainingPauses,
    maxDuration,
    activePause,
    autoResumeAt,
    // Actions
    fetchPauseStatus,
    pause,
    resume,
    fetchHistory,
    reset
  }
})
