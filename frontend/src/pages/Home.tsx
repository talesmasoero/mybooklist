import { useCallback, useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { getSession, clearSession } from '@/lib/auth'
import { Button } from '@/components/ui/Button'
import { AddBookModal } from '@/components/AddBookModal'
import { listLibrary, type Reading, type ReadingStatus } from '@/lib/api'

type LibraryFilter = 'all' | ReadingStatus

const FILTER_LABEL: Record<LibraryFilter, string> = {
  all: 'Todos',
  want_to_read: 'Quero ler',
  reading: 'Lendo',
  read: 'Lido',
  abandoned: 'Abandonado',
}

const STATUS_BADGE: Record<ReadingStatus, string> = {
  want_to_read: 'Quero ler',
  reading: 'Lendo',
  read: 'Lido',
  abandoned: 'Abandonado',
}

export function Home() {
  const navigate = useNavigate()
  const session = getSession()
  const userName = session?.user.name ?? 'Leitor'

  const [modalOpen, setModalOpen] = useState(false)
  const [readings, setReadings] = useState<Reading[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [filter, setFilter] = useState<LibraryFilter>('all')

  const reload = useCallback(async () => {
    setLoading(true)
    setError(null)
    try {
      const data = await listLibrary()
      setReadings(data)
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Erro ao carregar.'
      setError(message)
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    void reload()
  }, [reload])

  function handleLogout() {
    clearSession()
    navigate('/', { replace: true })
  }

  function handleAdded() {
    setModalOpen(false)
    void reload()
  }

  const reading = readings.filter((r) => r.status === 'reading')
  const filteredLibrary =
    filter === 'all' ? readings : readings.filter((r) => r.status === filter)

  return (
    <div className="min-h-screen bg-gray-50">
      <header className="bg-white shadow-sm">
        <div className="mx-auto flex max-w-7xl items-center justify-between px-4 py-4">
          <span className="text-lg font-bold text-[#162447]">MyBookList</span>
          <div className="flex items-center gap-4 text-sm text-gray-600">
            <span>
              Olá, <strong className="text-[#162447]">{userName}</strong>
            </span>
            <button
              onClick={handleLogout}
              className="text-gray-400 hover:text-red-500 transition-colors"
            >
              Sair
            </button>
          </div>
        </div>
      </header>

      <main className="mx-auto max-w-7xl px-4 py-10">
        <section className="mb-8">
          <div className="mb-4 flex items-center justify-between">
            <h2 className="text-xl font-semibold text-gray-800">Continuar lendo</h2>
            <Button
              variant="secondary"
              className="px-5 py-2.5 text-sm"
              onClick={() => setModalOpen(true)}
            >
              + Adicionar livro
            </Button>
          </div>

          {loading ? (
            <Placeholder>Carregando…</Placeholder>
          ) : error ? (
            <Placeholder>{error}</Placeholder>
          ) : reading.length === 0 ? (
            <div className="rounded-2xl border border-dashed border-gray-200 bg-white px-6 py-12 text-center text-gray-400">
              <p className="text-sm">Você ainda não está lendo nenhum livro.</p>
              <p className="mt-1 text-xs text-gray-300">
                Use "+ Adicionar livro" para começar.
              </p>
            </div>
          ) : (
            <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
              {reading.map((r) => (
                <ReadingCard key={r.id} reading={r} variant="large" />
              ))}
            </div>
          )}
        </section>

        <section className="mb-8">
          <h2 className="mb-4 text-xl font-semibold text-gray-800">Meta anual</h2>
          <div className="rounded-2xl border border-dashed border-gray-200 bg-white px-6 py-10 text-center text-gray-400">
            <p className="text-sm">
              Defina sua meta de leitura para {new Date().getFullYear()}.
            </p>
            <p className="mt-1 text-xs text-gray-300">Em breve</p>
          </div>
        </section>

        <section>
          <h2 className="mb-4 text-xl font-semibold text-gray-800">Minha biblioteca</h2>

          <div className="mb-4 flex flex-wrap gap-2">
            {(Object.keys(FILTER_LABEL) as LibraryFilter[]).map((key) => (
              <button
                key={key}
                onClick={() => setFilter(key)}
                className={`rounded-full border px-4 py-1.5 text-xs font-medium transition-colors ${
                  filter === key
                    ? 'border-[#162447] bg-[#162447] text-white'
                    : 'border-gray-300 text-gray-600 hover:bg-gray-100'
                }`}
              >
                {FILTER_LABEL[key]}
              </button>
            ))}
          </div>

          {loading ? (
            <Placeholder>Carregando…</Placeholder>
          ) : filteredLibrary.length === 0 ? (
            <Placeholder>Nenhum livro nesse filtro.</Placeholder>
          ) : (
            <div className="grid grid-cols-2 gap-3 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5">
              {filteredLibrary.map((r) => (
                <ReadingCard key={r.id} reading={r} variant="compact" />
              ))}
            </div>
          )}
        </section>
      </main>

      <AddBookModal
        open={modalOpen}
        onClose={() => setModalOpen(false)}
        onAdded={handleAdded}
      />
    </div>
  )
}

function Placeholder({ children }: { children: React.ReactNode }) {
  return (
    <div className="rounded-2xl border border-dashed border-gray-200 bg-white px-6 py-10 text-center text-sm text-gray-400">
      {children}
    </div>
  )
}

function ReadingCard({
  reading,
  variant,
}: {
  reading: Reading
  variant: 'large' | 'compact'
}) {
  const book = reading.book
  if (!book) return null

  if (variant === 'compact') {
    return (
      <div className="flex flex-col rounded-xl border border-gray-200 bg-white p-2">
        <div className="mb-2 aspect-[2/3] w-full overflow-hidden rounded bg-gray-100">
          {book.cover_url && (
            <img
              src={book.cover_url}
              alt=""
              className="h-full w-full object-cover"
            />
          )}
        </div>
        <p className="line-clamp-2 text-xs font-medium text-[#162447]">{book.title}</p>
        <p className="line-clamp-1 text-[11px] text-gray-500">
          {book.authors.join(', ')}
        </p>
        <span className="mt-1 inline-block rounded bg-gray-100 px-1.5 py-0.5 text-[10px] text-gray-600">
          {STATUS_BADGE[reading.status]}
        </span>
      </div>
    )
  }

  return (
    <div className="flex gap-3 rounded-2xl border border-gray-200 bg-white p-4">
      <div className="h-28 w-20 flex-shrink-0 overflow-hidden rounded bg-gray-100">
        {book.cover_url && (
          <img src={book.cover_url} alt="" className="h-full w-full object-cover" />
        )}
      </div>
      <div className="flex flex-1 flex-col">
        <p className="text-sm font-semibold text-[#162447]">{book.title}</p>
        <p className="text-xs text-gray-500">{book.authors.join(', ')}</p>
        <p className="mt-2 text-xs text-gray-600">
          Página {reading.current_page}
          {book.total_pages ? ` / ${book.total_pages}` : ''}
        </p>
      </div>
    </div>
  )
}
