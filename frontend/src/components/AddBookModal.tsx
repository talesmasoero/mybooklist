import { useEffect, useRef, useState } from 'react'
import {
  searchBooks,
  addToLibrary,
  type BookSearchResult,
  type AddStatus,
  type AddToLibraryPayload,
} from '@/lib/api'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'

type Mode = 'search' | 'manual'

interface AddBookModalProps {
  open: boolean
  onClose: () => void
  onAdded: () => void
}

const STATUS_LABEL: Record<AddStatus, string> = {
  want_to_read: 'Quero ler',
  reading: 'Lendo',
}

export function AddBookModal({ open, onClose, onAdded }: AddBookModalProps) {
  const [mode, setMode] = useState<Mode>('search')
  const searchInputRef = useRef<HTMLInputElement>(null)

  useEffect(() => {
    if (!open) return
    const onKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose()
    }
    document.addEventListener('keydown', onKey)
    return () => document.removeEventListener('keydown', onKey)
  }, [open, onClose])

  useEffect(() => {
    if (!open || mode !== 'search') return
    const t = setTimeout(() => searchInputRef.current?.focus(), 30)
    return () => clearTimeout(t)
  }, [open, mode])

  if (!open) return null

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 px-4"
      onClick={onClose}
      role="dialog"
      aria-modal="true"
    >
      <div
        className="w-full max-w-2xl max-h-[90vh] overflow-y-auto rounded-2xl bg-white p-6 shadow-xl"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="mb-4 flex items-center justify-between">
          <h3 className="text-lg font-semibold text-[#162447]">Adicionar livro</h3>
          <button
            onClick={onClose}
            className="text-gray-400 hover:text-gray-600"
            aria-label="Fechar"
          >
            ✕
          </button>
        </div>

        <div className="mb-4 flex gap-2 border-b border-gray-200">
          <TabButton active={mode === 'search'} onClick={() => setMode('search')}>
            Buscar online
          </TabButton>
          <TabButton active={mode === 'manual'} onClick={() => setMode('manual')}>
            Cadastrar manualmente
          </TabButton>
        </div>

        {mode === 'search' ? (
          <SearchTab
            inputRef={searchInputRef}
            onAdded={onAdded}
          />
        ) : (
          <ManualTab onAdded={onAdded} />
        )}
      </div>
    </div>
  )
}

function TabButton({
  active,
  onClick,
  children,
}: {
  active: boolean
  onClick: () => void
  children: React.ReactNode
}) {
  return (
    <button
      onClick={onClick}
      className={`px-4 py-2 text-sm font-medium transition-colors ${
        active
          ? 'border-b-2 border-[#162447] text-[#162447]'
          : 'text-gray-500 hover:text-gray-700'
      }`}
    >
      {children}
    </button>
  )
}

function SearchTab({
  inputRef,
  onAdded,
}: {
  inputRef: React.RefObject<HTMLInputElement | null>
  onAdded: () => void
}) {
  const [query, setQuery] = useState('')
  const [results, setResults] = useState<BookSearchResult[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [submittingId, setSubmittingId] = useState<string | null>(null)

  useEffect(() => {
    const trimmed = query.trim()
    if (trimmed.length < 2) {
      setResults([])
      setError(null)
      return
    }
    setLoading(true)
    setError(null)
    const handle = setTimeout(async () => {
      try {
        const data = await searchBooks(trimmed, 10)
        setResults(data)
      } catch (err) {
        const message = err instanceof Error ? err.message : 'Erro na busca.'
        setError(message)
        setResults([])
      } finally {
        setLoading(false)
      }
    }, 400)
    return () => clearTimeout(handle)
  }, [query])

  async function handleAdd(item: BookSearchResult, status: AddStatus) {
    setSubmittingId(item.google_books_id)
    setError(null)
    try {
      const payload: AddToLibraryPayload = {
        source: 'google_books',
        google_books_id: item.google_books_id,
        book_data: {
          title: item.title,
          authors: item.authors,
          genres: item.genres,
          isbn: item.isbn,
          synopsis: item.synopsis,
          cover_url: item.cover_url,
          total_pages: item.total_pages,
        },
        status,
      }
      await addToLibrary(payload)
      onAdded()
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Erro ao adicionar.'
      setError(message)
    } finally {
      setSubmittingId(null)
    }
  }

  return (
    <div>
      <input
        ref={inputRef}
        type="text"
        value={query}
        onChange={(e) => setQuery(e.target.value)}
        placeholder="Busque por título, autor ou ISBN…"
        className="w-full rounded-lg border border-gray-300 px-4 py-3 text-sm outline-none focus:border-[#162447] focus:ring-2 focus:ring-[#162447]/20"
      />

      <div className="mt-4 min-h-[120px]">
        {loading && <p className="text-sm text-gray-400">Buscando…</p>}
        {error && <p className="text-sm text-red-500">{error}</p>}
        {!loading && !error && query.trim().length < 2 && (
          <p className="text-sm text-gray-400">Digite pelo menos 2 caracteres.</p>
        )}
        {!loading && !error && query.trim().length >= 2 && results.length === 0 && (
          <p className="text-sm text-gray-400">Nenhum resultado.</p>
        )}

        <ul className="space-y-3">
          {results.map((item) => (
            <li
              key={item.google_books_id}
              className="flex gap-3 rounded-lg border border-gray-200 p-3"
            >
              <div className="h-20 w-14 flex-shrink-0 overflow-hidden rounded bg-gray-100">
                {item.cover_url && (
                  <img
                    src={item.cover_url}
                    alt=""
                    className="h-full w-full object-cover"
                  />
                )}
              </div>
              <div className="flex flex-1 flex-col">
                <p className="text-sm font-medium text-[#162447]">{item.title}</p>
                <p className="text-xs text-gray-500">
                  {item.authors.join(', ') || 'Autor desconhecido'}
                </p>
                {item.total_pages ? (
                  <p className="text-xs text-gray-400">{item.total_pages} páginas</p>
                ) : null}
                <div className="mt-2 flex gap-2">
                  <button
                    onClick={() => handleAdd(item, 'want_to_read')}
                    disabled={submittingId !== null}
                    className="rounded border border-[#162447] px-3 py-1 text-xs text-[#162447] hover:bg-[#162447]/5 disabled:opacity-50"
                  >
                    + Quero ler
                  </button>
                  <button
                    onClick={() => handleAdd(item, 'reading')}
                    disabled={submittingId !== null}
                    className="rounded bg-[#162447] px-3 py-1 text-xs text-white hover:bg-[#1f3a6e] disabled:opacity-50"
                  >
                    + Lendo
                  </button>
                </div>
              </div>
            </li>
          ))}
        </ul>
      </div>
    </div>
  )
}

function ManualTab({ onAdded }: { onAdded: () => void }) {
  const [title, setTitle] = useState('')
  const [authors, setAuthors] = useState('')
  const [genres, setGenres] = useState('')
  const [isbn, setIsbn] = useState('')
  const [synopsis, setSynopsis] = useState('')
  const [coverUrl, setCoverUrl] = useState('')
  const [totalPages, setTotalPages] = useState('')
  const [status, setStatus] = useState<AddStatus>('want_to_read')
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)

  function parseList(value: string): string[] {
    return value
      .split(',')
      .map((s) => s.trim())
      .filter(Boolean)
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setError(null)
    const authorsList = parseList(authors)
    if (!title.trim()) {
      setError('Título é obrigatório.')
      return
    }
    if (authorsList.length === 0) {
      setError('Informe ao menos um autor.')
      return
    }
    setSubmitting(true)
    try {
      const payload: AddToLibraryPayload = {
        source: 'manual',
        book_data: {
          title: title.trim(),
          authors: authorsList,
          genres: parseList(genres),
          isbn: isbn.trim() || undefined,
          synopsis: synopsis.trim() || undefined,
          cover_url: coverUrl.trim() || undefined,
          total_pages: totalPages ? Number(totalPages) : undefined,
        },
        status,
      }
      await addToLibrary(payload)
      onAdded()
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Erro ao adicionar.'
      setError(message)
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-3">
      <Input
        label="Título *"
        value={title}
        onChange={(e) => setTitle(e.target.value)}
        required
      />
      <Input
        label="Autores * (separados por vírgula)"
        value={authors}
        onChange={(e) => setAuthors(e.target.value)}
        placeholder="Machado de Assis, Clarice Lispector"
        required
      />
      <Input
        label="Gêneros (separados por vírgula)"
        value={genres}
        onChange={(e) => setGenres(e.target.value)}
        placeholder="Romance, Ficção"
      />
      <Input label="ISBN" value={isbn} onChange={(e) => setIsbn(e.target.value)} />
      <Input
        label="URL da capa"
        value={coverUrl}
        onChange={(e) => setCoverUrl(e.target.value)}
      />
      <Input
        label="Total de páginas"
        type="number"
        min={1}
        value={totalPages}
        onChange={(e) => setTotalPages(e.target.value)}
      />
      <div className="flex flex-col gap-1">
        <label className="text-sm font-medium text-gray-700">Sinopse</label>
        <textarea
          value={synopsis}
          onChange={(e) => setSynopsis(e.target.value)}
          rows={3}
          className="rounded-lg border border-gray-300 px-4 py-3 text-sm outline-none focus:border-[#162447] focus:ring-2 focus:ring-[#162447]/20"
        />
      </div>
      <div className="flex flex-col gap-1">
        <label className="text-sm font-medium text-gray-700">Status</label>
        <select
          value={status}
          onChange={(e) => setStatus(e.target.value as AddStatus)}
          className="rounded-lg border border-gray-300 px-4 py-3 text-sm outline-none focus:border-[#162447] focus:ring-2 focus:ring-[#162447]/20"
        >
          {(Object.keys(STATUS_LABEL) as AddStatus[]).map((s) => (
            <option key={s} value={s}>
              {STATUS_LABEL[s]}
            </option>
          ))}
        </select>
      </div>

      {error && <p className="text-sm text-red-500">{error}</p>}

      <Button type="submit" loading={submitting} className="w-full">
        Adicionar à biblioteca
      </Button>
    </form>
  )
}
