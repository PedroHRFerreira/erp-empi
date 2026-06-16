import { defineStore } from 'pinia'
import type { IMetricsSummary } from '../../server/contracts/types'

export const useMetricsStore = defineStore('metrics', {
  state: () => ({
    summary: null as IMetricsSummary | null,
    loading: false
  }),
  actions: {
    async load() {
      this.loading = true
      try {
        const { data, error } = await useApiFetch<IMetricsSummary>('/metrics/summary')
        if (error.value || !data.value) {
          throw new Error('metrics load failed')
        }
        this.summary = data.value
      } finally {
        this.loading = false
      }
    }
  }
})
