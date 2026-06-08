import { useEffect, useState } from 'react'
import { createGoal, updateGoal, type Goal, type GoalProgress } from '@/lib/api'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'

interface GoalModalProps {
  initialGoal?: Goal
  onCreated: (goal: GoalProgress) => void
  onClose: () => void
}

export function GoalModal({ initialGoal, onCreated, onClose }: GoalModalProps) {
  const year = initialGoal?.year ?? new Date().getFullYear()
  const [targetBooks, setTargetBooks] = useState(initialGoal ? String(initialGoal.target_books) : '12')
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose()
    }
    document.addEventListener('keydown', onKey)
    return () => document.removeEventListener('keydown', onKey)
  }, [onClose])

  function validate(): string | null {
    const n = Number(targetBooks)
    if (!targetBooks || !Number.isInteger(n) || n < 1 || n > 1000) {
      return 'Informe um número inteiro entre 1 e 1000.'
    }
    return null
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setError(null)
    const validationError = validate()
    if (validationError) {
      setError(validationError)
      return
    }
    setSubmitting(true)
    try {
      const goal = initialGoal
        ? await updateGoal(initialGoal.id, Number(targetBooks))
        : await createGoal(year, Number(targetBooks))
      onCreated(goal)
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Erro ao salvar meta.'
      setError(message)
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 px-4"
      onClick={onClose}
      role="dialog"
      aria-modal="true"
    >
      <div
        className="w-full max-w-md rounded-2xl bg-white p-6 shadow-xl"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="mb-5 flex items-center justify-between">
          <h3 className="text-lg font-semibold text-[#162447]">
            {initialGoal ? 'Editar meta' : `Meta para ${year}`}
          </h3>
          <button
            onClick={onClose}
            className="text-gray-400 hover:text-gray-600"
            aria-label="Fechar"
          >
            ✕
          </button>
        </div>

        <form onSubmit={handleSubmit} className="space-y-4">
          <Input
            label={`Quantos livros quero ler em ${year}`}
            type="number"
            min={1}
            max={1000}
            value={targetBooks}
            onChange={(e) => setTargetBooks(e.target.value)}
            required
          />

          {error && <p className="text-sm text-red-500">{error}</p>}

          <Button type="submit" loading={submitting} className="w-full">
            {initialGoal ? 'Salvar alterações' : 'Salvar meta'}
          </Button>
        </form>
      </div>
    </div>
  )
}
