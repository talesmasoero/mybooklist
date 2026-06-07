import { useEffect, useState } from 'react'
import ReactMarkdown from 'react-markdown'
import termsContent from '@/content/terms.md?raw'
import { Button } from '@/components/ui/Button'

interface TermsModalProps {
  open: boolean
  onClose: () => void
  onAccept: () => void
}

export function TermsModal({ open, onClose, onAccept }: TermsModalProps) {
  const [hasReachedBottom, setHasReachedBottom] = useState(false)

  useEffect(() => {
    if (!open) return
    function onKeyDown(e: KeyboardEvent) {
      if (e.key === 'Escape') onClose()
    }
    document.addEventListener('keydown', onKeyDown)
    return () => document.removeEventListener('keydown', onKeyDown)
  }, [open, onClose])

  function handleScroll(e: React.UIEvent<HTMLDivElement>) {
    const { scrollTop, scrollHeight, clientHeight } = e.currentTarget
    if (!hasReachedBottom && scrollTop + clientHeight >= scrollHeight - 8) {
      setHasReachedBottom(true)
    }
  }

  if (!open) return null

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 px-4"
      onClick={onClose}
      role="dialog"
      aria-modal="true"
      aria-labelledby="terms-modal-title"
    >
      <div
        className="flex w-full max-w-2xl flex-col rounded-2xl bg-white shadow-xl"
        style={{ maxHeight: '90vh' }}
        onClick={(e) => e.stopPropagation()}
      >
        {/* Header */}
        <div className="flex items-center justify-between border-b border-gray-100 px-6 py-4">
          <h2 id="terms-modal-title" className="text-base font-semibold text-[#162447]">
            Termos de Uso e Política de Privacidade
          </h2>
          <button
            type="button"
            onClick={onClose}
            className="text-gray-400 hover:text-gray-600"
            aria-label="Fechar"
          >
            ✕
          </button>
        </div>

        {/* Scrollable content */}
        <div
          className="flex-1 overflow-y-auto px-6 py-4"
          onScroll={handleScroll}
        >
          <ReactMarkdown
            components={{
              h1: ({ children }) => (
                <h1 className="mb-4 text-xl font-bold text-[#162447]">{children}</h1>
              ),
              h2: ({ children }) => (
                <h2 className="mb-3 mt-6 text-base font-semibold text-[#162447]">{children}</h2>
              ),
              h3: ({ children }) => (
                <h3 className="mb-2 mt-4 text-sm font-semibold text-gray-800">{children}</h3>
              ),
              p: ({ children }) => (
                <p className="mb-3 text-sm leading-relaxed text-gray-700">{children}</p>
              ),
              ul: ({ children }) => (
                <ul className="mb-3 ml-4 list-disc space-y-1 text-sm text-gray-700">{children}</ul>
              ),
              li: ({ children }) => <li className="leading-relaxed">{children}</li>,
              strong: ({ children }) => (
                <strong className="font-semibold text-gray-900">{children}</strong>
              ),
              hr: () => <hr className="my-4 border-gray-200" />,
            }}
          >
            {termsContent}
          </ReactMarkdown>
        </div>

        {/* Footer */}
        <div className="border-t border-gray-100 px-6 py-4">
          {!hasReachedBottom && (
            <p className="mb-3 text-center text-xs text-gray-400">
              Role até o final para habilitar o botão
            </p>
          )}
          <div className="flex gap-3">
            <Button
              type="button"
              variant="secondary"
              className="flex-1"
              onClick={onClose}
            >
              Cancelar
            </Button>
            <Button
              type="button"
              disabled={!hasReachedBottom}
              className="flex-1"
              onClick={onAccept}
            >
              Aceitar
            </Button>
          </div>
        </div>
      </div>
    </div>
  )
}
