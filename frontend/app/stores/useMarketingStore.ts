import { defineStore } from 'pinia'
import type { MarketingBrand, MarketingCreation, MarketingCreationInput, MarketingImage, MarketingTemplate } from '~/types/marketing'

export const useMarketingStore = defineStore('marketing', {
  state: () => ({
    brand: null as MarketingBrand | null,
    creations: [] as MarketingCreation[],
    templates: [] as MarketingTemplate[],
    total: 0,
    loading: false,
    error: ''
  }),
  actions: {
    async loadTemplates(forceRefresh = false) {
      if (this.templates.length && !forceRefresh) return true
      if (forceRefresh) invalidateApiCache(['/marketing/templates'])
      const { data, status } = await useApiFetch<MarketingTemplate[]>('/marketing/templates')
      if (status.value === 'error' || !data.value) { this.error = 'Não foi possível carregar os modelos.'; return false }
      this.templates = data.value; this.error = ''; return true
    },
    async loadBrand(forceRefresh = false) {
      if (forceRefresh) invalidateApiCache(['/marketing/brand'])
      const { data, status } = await useApiFetch<MarketingBrand>('/marketing/brand')
      if (status.value === 'error' || !data.value) { this.error = 'Não foi possível carregar a identidade visual.'; return false }
      this.brand = data.value; this.error = ''; return true
    },
    async saveBrand(input: MarketingBrand) {
      const { data, status } = await useApiFetch<MarketingBrand>('/marketing/brand', { method: 'PUT', body: input })
      if (status.value === 'error' || !data.value) { this.error = 'Não foi possível salvar a identidade visual.'; return false }
      invalidateApiCache(['/marketing/brand']); this.brand = data.value; this.error = ''; return true
    },
    async loadCreations(offset = 0, forceRefresh = false) {
      if (forceRefresh) invalidateApiCache(['/marketing/creations'])
      this.loading = true
      const { data, status } = await useApiFetch<{ data: MarketingCreation[]; total: number }>('/marketing/creations', { query: { limit: 24, offset } })
      this.loading = false
      if (status.value === 'error' || !data.value) { this.error = 'Não foi possível carregar as criações.'; return false }
      this.creations = data.value.data; this.total = data.value.total; this.error = ''; return true
    },
    async saveCreation(input: MarketingCreationInput, id?: string): Promise<MarketingCreation | null> {
      const { data, status } = await useApiFetch<MarketingCreation>(id ? `/marketing/creations/${id}` : '/marketing/creations', { method: id ? 'PUT' : 'POST', body: input })
      if (status.value === 'error' || !data.value) { this.error = 'Não foi possível salvar a criação. Seu conteúdo continua disponível.'; return null }
      invalidateApiCache(['/marketing/creations']); this.error = ''; return data.value
    },
    async deleteCreation(id: string) {
      const { status } = await useApiFetch(`/marketing/creations/${id}`, { method: 'DELETE' })
      if (status.value === 'error') { this.error = 'Não foi possível excluir a criação.'; return false }
      invalidateApiCache(['/marketing/creations']); this.creations = this.creations.filter(item => item.id !== id); this.total = Math.max(0, this.total - 1); this.error = ''; return true
    },
    async uploadImage(file: File): Promise<MarketingImage | null> {
      const body = new FormData(); body.append('file', file)
      const { data, status } = await useApiFetch<MarketingImage>('/marketing/images', { method: 'POST', body })
      if (status.value === 'error' || !data.value) { this.error = 'Não foi possível enviar a imagem. Tente novamente; sua criação foi mantida.'; return null }
      this.error = ''; return data.value
    },
    async loadImage(id: string): Promise<string | null> {
      const config = useRuntimeConfig()
      const authorization = useAuthToken().getAuthorizationHeader()
      try {
        const response = await fetch(`${config.public.apiBaseUrl || '/api'}/marketing/images/${encodeURIComponent(id)}`, { headers: authorization ? { Authorization: authorization } : {} })
        if (!response.ok) throw new Error('Image request failed')
        return URL.createObjectURL(await response.blob())
      } catch {
        return null
      }
    }
  }
})
