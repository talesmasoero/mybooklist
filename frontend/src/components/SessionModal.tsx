import { useEffect, useState } from 'react'
import { createSession, type CreateSessionPayload } from '@/lib/api'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'

interface SessionModalProps {
  readingId: string
  currentPage: number
  totalPages?: number
  onCreated: () => void
  onClose: () => void
}

function todayISO(): string {
  return new Date().toISOString().slice(0, 10)
}

export function SessionModal({ readingId, currentPage, totalPages, onCreated, onClose }: SessionModalProps) {
  const [startPage, setStartPage] = useState(String(currentPage))
  const [endPage, setEndPage] = useState('')
  const [sessionDate, setSessionDate] = useState(todayISO())
  const [durationMin, setDurationMin] = useState('')
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
    const start = Number(startPage)
    const end = Number(endPage)
    if (!startPage || start < 1) return 'Página inicial deve ser pelo menos 1.'
    if (!endPage || end < 1) return 'Página final é obrigatória.'
    if (end < start) return 'Página final deve ser maior ou igual à página inicial.'
    if (totalPages && end > totalPages) return `Página final não pode passar de ${totalPages}.`
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
      const payload: CreateSessionPayload = {
        start_page: Number(startPage),
        end_page: Number(endPage),
        session_date: sessionDate || undefined,
        duration_seconds: durationMin ? Math.round(Number(durationMin) * 60) : undefined,
      }
      await createSession(readingId, payload)
      onCreated()
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Erro ao registrar sessão.'
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
          <h3 className="text-lg font-semibold text-[#162447]">Registrar sessão</h3>
          <button
            onClick={onClose}
            className="text-gray-400 hover:text-gray-600"
            aria-label="Fechar"
          >
            ✕
          </button>
        </div>

        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="grid grid-cols-2 gap-3">
            <Input
              label="Página inicial"
              type="number"
              min={1}
              value={startPage}
              onChange={(e) => setStartPage(e.target.value)}
              required
            />
            <Input
              label="Página final *"
              type="number"
              min={1}
              value={endPage}
              onChange={(e) => setEndPage(e.target.value)}
              required
            />
          </div>

          <div className="flex flex-col gap-1">
            <label className="text-sm font-medium text-gray-700">Data da sessão</label>
            <input
              type="date"
              value={sessionDate}
              onChange={(e) => setSessionDate(e.target.value)}
              className="rounded-lg border border-gray-300 px-4 py-3 text-sm outline-none focus:border-[#162447] focus:ring-2 focus:ring-[#162447]/20"
            />
          </div>

          <Input
            label="Duração (minutos, opcional)"
            type="number"
            min={1}
            value={durationMin}
            onChange={(e) => setDurationMin(e.target.value)}
          />

          {error && <p className="text-sm text-red-500">{error}</p>}

          <Button type="submit" loading={submitting} className="w-full">
            Salvar sessão
          </Button>
        </form>
      </div>
    </div>
  )
}
