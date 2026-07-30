import { defineStore } from 'pinia'
import type { IGoalsSummary, IMonthlyGoal } from '../../server/contracts/types'

export const useGoalsStore = defineStore('goals', {
  state: () => ({ summary: null as IGoalsSummary | null, loading: false, error: '' }),
  actions: {
    async load(month = previousMonth(), startDate = '', endDate = '', forceRefresh = false) {
      if (forceRefresh) invalidateApiCache(['/goals'])
      this.loading = true
      const { data, status } = await useApiFetch<IGoalsSummary>('/goals', { query: { month, startDate, endDate } })
      this.loading = false
      if (status.value === 'error' || !data.value) {
        this.error = 'Não foi possível carregar as metas.'
        return false
      }
      this.summary = data.value
      this.error = ''
      return true
    },
    async save(targets: IMonthlyGoal) {
      this.loading = true
      const { data, status } = await useApiFetch<IGoalsSummary>('/goals', {
        method: 'PUT',
        query: { month: this.summary?.month || targets.month },
        body: targets
      })
      this.loading = false
      if (status.value === 'error' || !data.value) {
        this.error = 'Não foi possível salvar as metas.'
        return false
      }
      invalidateApiCache(['/goals'])
      this.summary = data.value
      this.error = ''
      return true
    }
  }
})

function previousMonth() {
  const now = new Date()
  const previous = new Date(now.getFullYear(), now.getMonth() - 1, 1)
  return `${previous.getFullYear()}-${String(previous.getMonth() + 1).padStart(2, '0')}`
}
