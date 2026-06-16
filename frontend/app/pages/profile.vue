<script setup lang="ts">
import { Save } from '@lucide/vue'
import type { IUser } from '../../server/contracts/types'

const auth = useAuthStore()
const profile = useProfileStore()
const form = reactive<IUser>({ ...(auth.user as IUser) })
const saved = ref(false)

async function save() {
  const updated = await profile.update(form)
  auth.user = updated
  Object.assign(form, updated)
  saved.value = true
  window.setTimeout(() => {
    saved.value = false
  }, 2200)
}
</script>

<template>
  <section class="page">
    <header class="page__header">
      <div>
        <h1 class="page__title">Perfil</h1>
        <p class="page__subtitle">Dados do administrador e markup padrao da oficina.</p>
      </div>
    </header>

    <form class="profile-form panel" @submit.prevent="save">
      <label class="field">
        <span>Nome</span>
        <input v-model="form.name" required />
      </label>
      <label class="field">
        <span>CPF</span>
        <input v-model="form.cpf" required inputmode="numeric" />
      </label>
      <label class="field">
        <span>Telefone</span>
        <input v-model="form.phone" inputmode="numeric" />
      </label>
      <label class="field">
        <span>E-mail</span>
        <input v-model="form.email" type="email" />
      </label>
      <label class="field">
        <span>Markup acima do produto (%)</span>
        <input v-model.number="form.markupPercent" min="0" step="0.1" type="number" />
      </label>
      <label class="field profile-form__wide">
        <span>Endereco</span>
        <input v-model="form.address" />
      </label>
      <label class="field profile-form__wide">
        <span>Observacoes</span>
        <textarea v-model="form.notes" />
      </label>

      <p v-if="profile.error" class="form-error profile-form__wide">{{ profile.error }}</p>
      <p v-if="saved" class="form-success profile-form__wide">Perfil atualizado.</p>

      <button class="button button--primary profile-form__wide" type="submit" :disabled="profile.saving">
        <Save :size="18" />
        {{ profile.saving ? 'Salvando...' : 'Salvar perfil' }}
      </button>
    </form>
  </section>
</template>

<style scoped lang="scss">
.profile-form {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 16px;
  padding: 18px;
}

.profile-form__wide {
  grid-column: 1 / -1;
}

.form-error,
.form-success {
  margin: 0;
  font-weight: 700;
}

.form-error {
  color: var(--status-danger);
}

.form-success {
  color: var(--status-success);
}

@media (max-width: 760px) {
  .profile-form {
    grid-template-columns: 1fr;
  }
}
</style>
