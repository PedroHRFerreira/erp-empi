<script lang="ts">
import { defineComponent, reactive, ref } from 'vue'
import type { IUser } from '../../../../server/contracts/types'
import PageHeader from '../../molecules/PageHeader/Index.vue'
import ProfileForm from '../../organisms/ProfileForm/Index.vue'
import { maskCpf, maskPhone } from '../../../utils/masks'

function emptyUser(): IUser {
  return {
    id: '',
    name: '',
    cpf: '',
    type: 'admin',
    email: '',
    phone: '',
    markupPercent: 0,
    machineFeePercent: 0,
    address: '',
    notes: '',
    createdAt: '',
    updatedAt: ''
  }
}

export default defineComponent({
  name: 'ProfileTemplate',
  components: {
    PageHeader,
    ProfileForm
  },
  setup() {
    const auth = useAuthStore()
    const profile = useProfileStore()
    const form = reactive<IUser>({ ...emptyUser(), ...(auth.user || {}) })
    const saved = ref(false)

    form.cpf = maskCpf(form.cpf || '')
    form.phone = maskPhone(form.phone || '')

    async function save() {
      const result = await profile.update(form)

      if (result.status === 'error' || !result.data) {
        return
      }

      auth.user = result.data
      Object.assign(form, result.data)
      form.cpf = maskCpf(form.cpf || '')
      form.phone = maskPhone(form.phone || '')
      saved.value = true
      window.setTimeout(() => {
        saved.value = false
      }, 2200)
    }

    return {
      form,
      profile,
      save,
      saved
    }
  }
})
</script>

<template>
  <section class="page">
    <PageHeader title="Perfil" subtitle="Dados do administrador e markup padrão da oficina." />

    <ProfileForm
      :error="profile.error"
      :field-errors="profile.fieldErrors"
      :form="form"
      :saved="saved"
      :saving="profile.saving"
      @clear-field-error="profile.clearFieldError"
      @save="save"
    />
  </section>
</template>

<style scoped lang="scss">
@use "styles.module.scss";
</style>
