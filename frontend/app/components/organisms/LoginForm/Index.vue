<script lang="ts">
import { defineComponent, reactive } from 'vue'
import BrandMark from '../../atoms/BrandMark/Index.vue'
import { maskCpf } from '../../../utils/masks'

export default defineComponent({
  name: 'LoginForm',
  components: {
    BrandMark
  },
  setup() {
    const auth = useAuthStore()
    const form = reactive({ cpf: '', password: '' })

    function maskLoginCpf() {
      auth.clearFieldError('cpf')
      auth.clearFieldError('password')
      form.cpf = maskCpf(form.cpf)
    }

    function clearPasswordError() {
      auth.clearFieldError('password')
    }

    async function submit() {
      const result = await auth.login(form)

      if (result.status === 'success') {
        await navigateTo('/')
      }
    }

    return {
      auth,
      clearPasswordError,
      form,
      maskLoginCpf,
      submit
    }
  }
})
</script>

<template>
  <section class="login-form panel">
    <BrandMark initials="EA" subtitle="Controle operacional" title="EMPI ERP" />

    <form class="login-form__form" novalidate @submit.prevent="submit">
      <label class="field" :class="{ 'field--error': auth.fieldErrors.cpf }">
        <span>CPF</span>
        <input
          v-model="form.cpf"
          autocomplete="username"
          inputmode="numeric"
          placeholder="000.000.000-00"
          required
          @input="maskLoginCpf"
        />
        <small v-if="auth.fieldErrors.cpf" class="field__error">{{ auth.fieldErrors.cpf }}</small>
      </label>

      <label class="field" :class="{ 'field--error': auth.fieldErrors.password }">
        <span>Senha</span>
        <input v-model="form.password" autocomplete="current-password" type="password" placeholder="Sua senha" required @input="clearPasswordError" />
        <small v-if="auth.fieldErrors.password" class="field__error">{{ auth.fieldErrors.password }}</small>
      </label>

      <p v-if="auth.error" class="login-form__error">{{ auth.error }}</p>

      <button class="button button--primary" type="submit" :disabled="auth.loading">
        <span v-if="auth.loading" class="login-form__spinner" aria-hidden="true" />
        {{ auth.loading ? 'Entrando...' : 'Entrar' }}
      </button>
    </form>
  </section>
</template>

<style scoped lang="scss">
@use "styles.module.scss";
</style>
