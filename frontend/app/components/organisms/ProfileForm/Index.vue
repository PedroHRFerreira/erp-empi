<script lang="ts">
import { Save } from '@lucide/vue'
import { defineComponent, type PropType } from 'vue'
import type { IUser } from '../../../../server/contracts/types'
import { maskCpf, maskPhone } from '../../../utils/masks'

export default defineComponent({
  name: 'ProfileForm',
  components: {
    Save
  },
  props: {
    form: {
      type: Object as PropType<IUser>,
      required: true
    },
    fieldErrors: {
      type: Object as PropType<Record<string, string>>,
      required: true
    },
    error: {
      type: String,
      default: ''
    },
    saved: {
      type: Boolean,
      default: false
    },
    saving: {
      type: Boolean,
      default: false
    }
  },
  emits: ['clear-field-error', 'save'],
  setup(props, { emit }) {
    function clearFieldError(field: string) {
      emit('clear-field-error', field)
    }

    function maskProfileCpf() {
      clearFieldError('cpf')
      props.form.cpf = maskCpf(props.form.cpf)
    }

    function maskProfilePhone() {
      clearFieldError('phone')
      props.form.phone = maskPhone(props.form.phone)
    }

    function save() {
      emit('save')
    }

    return {
      clearFieldError,
      maskProfileCpf,
      maskProfilePhone,
      save
    }
  }
})
</script>

<template>
  <form class="profile-form panel" novalidate @submit.prevent="save">
    <label class="field" :class="{ 'field--error': fieldErrors.name }">
      <span>Nome</span>
      <input v-model="form.name" required @input="clearFieldError('name')" />
      <small v-if="fieldErrors.name" class="field__error">{{ fieldErrors.name }}</small>
    </label>

    <label class="field" :class="{ 'field--error': fieldErrors.cpf }">
      <span>CPF</span>
      <input v-model="form.cpf" required inputmode="numeric" @input="maskProfileCpf" />
      <small v-if="fieldErrors.cpf" class="field__error">{{ fieldErrors.cpf }}</small>
    </label>

    <label class="field" :class="{ 'field--error': fieldErrors.phone }">
      <span>Telefone</span>
      <input v-model="form.phone" inputmode="numeric" placeholder="(33) 98735-1922" @input="maskProfilePhone" />
      <small v-if="fieldErrors.phone" class="field__error">{{ fieldErrors.phone }}</small>
    </label>

    <label class="field">
      <span>E-mail</span>
      <input v-model="form.email" type="email" @input="clearFieldError('email')" />
    </label>

    <label class="field" :class="{ 'field--error': fieldErrors.markupPercent }">
      <span>Margem de revenda padrão (%)</span>
      <input v-model.number="form.markupPercent" min="0" step="0.1" type="number" @input="clearFieldError('markupPercent')" />
      <small v-if="!fieldErrors.markupPercent" class="field__hint">Usado como padrão ao adicionar novos produtos no estoque.</small>
      <small v-if="fieldErrors.markupPercent" class="field__error">{{ fieldErrors.markupPercent }}</small>
    </label>

    <label class="field" :class="{ 'field--error': fieldErrors.machineFeePercent }">
      <span>Juros da maquininha (%)</span>
      <input v-model.number="form.machineFeePercent" min="0" step="0.1" type="number" @input="clearFieldError('machineFeePercent')" />
      <small v-if="!fieldErrors.machineFeePercent" class="field__hint">Aplicado ao débito e ao crédito à vista.</small>
      <small v-if="fieldErrors.machineFeePercent" class="field__error">{{ fieldErrors.machineFeePercent }}</small>
    </label>

    <label class="field" :class="{ 'field--error': fieldErrors.installmentFeePercent }">
      <span>Juros de parcelamento (%)</span>
      <input
        v-model.number="form.installmentFeePercent"
        min="0"
        step="0.1"
        type="number"
        @input="clearFieldError('installmentFeePercent')"
      />
      <small v-if="!fieldErrors.installmentFeePercent" class="field__hint">Aplicado ao crédito em 2x ou mais.</small>
      <small v-if="fieldErrors.installmentFeePercent" class="field__error">{{ fieldErrors.installmentFeePercent }}</small>
    </label>

    <label class="field profile-form__wide">
      <span>Endereço</span>
      <input v-model="form.address" @input="clearFieldError('address')" />
    </label>

    <label class="field profile-form__wide">
      <span>Observações</span>
      <textarea v-model="form.notes" @input="clearFieldError('notes')" />
    </label>

    <p v-if="error" class="form-error profile-form__wide">{{ error }}</p>
    <p v-if="saved" class="form-success profile-form__wide">Perfil atualizado.</p>

    <button class="button button--primary profile-form__wide" type="submit" :disabled="saving">
      <Save :size="18" />
      {{ saving ? 'Salvando...' : 'Salvar perfil' }}
    </button>
  </form>
</template>

<style scoped lang="scss">
@use "styles.module.scss";
</style>
