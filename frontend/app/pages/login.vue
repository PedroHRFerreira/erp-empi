<script setup lang="ts">
definePageMeta({ layout: 'auth' })

const auth = useAuthStore()
const form = reactive({ cpf: '', password: '' })

async function submit() {
  await auth.login(form)
  await navigateTo('/')
}
</script>

<template>
  <section class="login panel">
    <div class="login__brand">
      <span>EA</span>
      <div>
        <strong>EMPI ERP</strong>
        <small>Controle operacional</small>
      </div>
    </div>

    <form class="login__form" @submit.prevent="submit">
      <label class="field">
        <span>CPF</span>
        <input v-model="form.cpf" autocomplete="username" inputmode="numeric" placeholder="000.000.000-00" required />
      </label>

      <label class="field">
        <span>Senha</span>
        <input v-model="form.password" autocomplete="current-password" type="password" placeholder="Sua senha" required />
      </label>

      <p v-if="auth.error" class="login__error">{{ auth.error }}</p>

      <button class="button button--primary" type="submit" :disabled="auth.loading">
        <span v-if="auth.loading" class="login__spinner" aria-hidden="true" />
        {{ auth.loading ? 'Entrando...' : 'Entrar' }}
      </button>
    </form>
  </section>
</template>

<style scoped lang="scss">
.login {
  width: min(100%, 420px);
  padding: 28px;
}

.login__brand {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 28px;
}

.login__brand span {
  display: grid;
  width: 48px;
  height: 48px;
  place-items: center;
  border-radius: 8px;
  color: var(--braip-neutral-0);
  background: var(--braip-brand-primary);
  font-weight: 900;
}

.login__brand strong,
.login__brand small {
  display: block;
}

.login__brand small {
  color: var(--braip-theme-text-muted);
}

.login__form {
  display: grid;
  gap: 16px;
}

.login__error {
  margin: 0;
  color: var(--status-danger);
  font-weight: 700;
}

.login__spinner {
  width: 16px;
  height: 16px;
  border: 2px solid rgba(255, 255, 255, 0.55);
  border-top-color: var(--braip-neutral-0);
  border-radius: 50%;
  animation: spin 800ms linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}
</style>
