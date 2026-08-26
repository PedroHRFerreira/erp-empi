export type FeedbackTone = 'info' | 'success' | 'warning' | 'danger'

export interface SystemToast {
  id: string
  title: string
  message?: string
  tone: FeedbackTone
  duration: number
}

export interface SystemConfirmation {
  title: string
  message: string
  warning?: string
  tone: FeedbackTone
  confirmLabel: string
  cancelLabel: string
  details?: Array<{ label: string; value: string }>
  input?: { label: string; type: 'date'; value: string }
  resolve: (confirmed: boolean, inputValue?: string) => void
}

interface ConfirmationOptions {
  title: string
  message: string
  warning?: string
  tone?: FeedbackTone
  confirmLabel?: string
  cancelLabel?: string
  details?: Array<{ label: string; value: string }>
}

let toastSequence = 0

export function useSystemFeedback() {
  const toasts = useState<SystemToast[]>('system-feedback-toasts', () => [])
  const confirmation = useState<SystemConfirmation | null>('system-feedback-confirmation', () => null)

  function toast(title: string, options: { message?: string; tone?: FeedbackTone; duration?: number } = {}) {
    const id = `feedback-${Date.now()}-${++toastSequence}`
    toasts.value.push({ id, title, message: options.message, tone: options.tone || 'info', duration: options.duration ?? 5_000 })
    return id
  }

  function dismiss(id: string) {
    toasts.value = toasts.value.filter(item => item.id !== id)
  }

  function confirm(options: ConfirmationOptions) {
    if (confirmation.value) return Promise.resolve(false)
    return new Promise<boolean>((resolve) => {
      confirmation.value = {
        title: options.title,
        message: options.message,
        warning: options.warning,
        tone: options.tone || 'warning',
        confirmLabel: options.confirmLabel || 'Confirmar',
        cancelLabel: options.cancelLabel || 'Cancelar',
        details: options.details,
        resolve,
      }
    })
  }

  function answer(confirmed: boolean) {
    const current = confirmation.value
    if (!current) return
    confirmation.value = null
    current.resolve(confirmed, current.input?.value)
  }

  function confirmWithDate(options: ConfirmationOptions & { inputLabel: string; inputValue: string }) {
    if (confirmation.value) return Promise.resolve({ confirmed: false, value: options.inputValue })
    return new Promise<{ confirmed: boolean; value: string }>((resolve) => {
      confirmation.value = {
        title: options.title,
        message: options.message,
        warning: options.warning,
        tone: options.tone || 'warning',
        confirmLabel: options.confirmLabel || 'Confirmar',
        cancelLabel: options.cancelLabel || 'Cancelar',
        details: options.details,
        input: { label: options.inputLabel, type: 'date', value: options.inputValue },
        resolve: (confirmed, value) => resolve({ confirmed, value: value || options.inputValue }),
      }
    })
  }

  return {
    confirmation,
    toasts,
    answer,
    confirm,
    confirmWithDate,
    dismiss,
    error: (title: string, message?: string) => toast(title, { message, tone: 'danger' }),
    info: (title: string, message?: string) => toast(title, { message, tone: 'info' }),
    success: (title: string, message?: string) => toast(title, { message, tone: 'success' }),
    warning: (title: string, message?: string) => toast(title, { message, tone: 'warning' }),
  }
}
