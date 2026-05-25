import { useCallback, useEffect, useState } from 'react'
import { useParams, useNavigate, Link } from 'react-router-dom'
import { getSession, clearSession } from '@/lib/auth'
import { getReading, listSessions, deleteSession, type Reading, type Session } from '@/lib/api'
import { SessionModal } from '@/components/SessionModal'

const STATUS_LABEL: Record<string, string> = {
  want_to_read: 'Quero ler',
  reading: 'Lendo',
  read: 'Lido',
  abandoned: 'Abandonado',
}

function formatDate(iso: string): string {
  const [year, month, day] = iso.slice(0, 10).split('-')
  return `${day}/${month}/${year}`
}

export function ReadingDetail() {
  const { readingId } = useParams<{ readingId: string }>()
  const navigate = useNavigate()
  const session = getSession()
  const userName = session?.user.name ?? 'Leitor'

  const [reading, setReading] = useState<Reading | null>(null)
  const [sessions, setSessions] = useState<Session[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [sessionModalOpen, setSessionModalOpen] = useState(false)
  const [editingSession, setEditingSession] = useState<Session | null>(null)

  const load = useCallback(async () => {
    if (!readingId) return
    setLoading(true)
    setError(null)
    try {
      const [readingData, sessionsData] = await Promise.all([
        getReading(readingId),
        listSessions(readingId),
      ])
      setReading(readingData)
      setSessions(sessionsData)
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Erro ao carregar.'
      setError(message)
    } finally {
      setLoading(false)
    }
  }, [readingId])

  useEffect(() => {
    void load()
  }, [load])

  function handleLogout() {
    clearSession()
    navigate('/', { replace: true })
  }

  async function handleDeleteSession(sessionId: string) {
    if (!window.confirm('Tem certeza que deseja excluir esta sessão?')) return
    try {
      await deleteSession(sessionId)
      void load()
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Erro ao excluir sessão.'
      window.alert(message)
    }
  }

  function handleSessionSaved() {
    setSessionModalOpen(false)
    setEditingSession(null)
    void load()
  }

  if (loading) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-gray-50 text-sm text-gray-400">
        Carregando…
      </div>
    )
  }

  if (error || !reading) {
    return (
      <div className="flex min-h-screen flex-col items-center justify-center gap-4 bg-gray-50 px-4 text-center">
        <p className="text-sm text-gray-500">{error ?? 'Leitura não encontrada.'}</p>
        <button
          onClick={() => navigate('/home')}
          className="rounded-lg border border-[#162447] px-4 py-2 text-sm font-medium text-[#162447] transition-colors hover:bg-[#162447]/5"
        >
          Voltar para a biblioteca
        </button>
      </div>
    )
  }

  const book = reading.book!

  return (
    <div className="min-h-screen bg-gray-50">
      <header className="bg-white shadow-sm">
        <div className="mx-auto flex max-w-7xl items-center justify-between px-4 py-4">
          <Link to="/home" className="text-lg font-bold text-[#162447]">
            MyBookList
          </Link>
          <div className="flex items-center gap-4 text-sm text-gray-600">
            <Link to="/profile" className="hover:underline">
              Olá, <strong className="text-[#162447]">{userName}</strong>
            </Link>
            <button
              onClick={handleLogout}
              className="text-gray-400 transition-colors hover:text-red-500"
            >
              Sair
            </button>
          </div>
        </div>
      </header>

      <main className="mx-auto max-w-7xl px-4 py-10">
        <div className="mb-6">
          <button
            onClick={() => navigate('/home')}
            className="text-sm text-gray-500 transition-colors hover:text-[#162447]"
          >
            ← Voltar
          </button>
        </div>

        <section className="mb-8">
          <h2 className="mb-4 text-xl font-semibold text-gray-800">Informações do livro</h2>
          <div className="flex gap-5 rounded-2xl border border-gray-200 bg-white p-5">
            <div className="h-28 w-20 flex-shrink-0 overflow-hidden rounded bg-gray-100">
              {book.cover_url && (
                <img src={book.cover_url} alt="" className="h-full w-full object-cover" />
              )}
            </div>
            <div className="flex flex-1 flex-col">
              <p className="text-base font-semibold text-[#162447]">{book.title}</p>
              <p className="text-sm text-gray-500">{book.authors.join(', ')}</p>
              <p className="mt-1 text-xs text-gray-600">
                Página {reading.current_page}
                {book.total_pages ? ` / ${book.total_pages}` : ''}
              </p>
              <span className="mt-2 inline-block self-start rounded bg-gray-100 px-2 py-0.5 text-xs text-gray-600">
                {STATUS_LABEL[reading.status]}
              </span>
              <button
                onClick={() => setSessionModalOpen(true)}
                className="mt-3 self-start rounded-lg border border-[#162447] px-3 py-1.5 text-xs font-medium text-[#162447] transition-colors hover:bg-[#162447]/5"
              >
                Registrar nova sessão
              </button>
            </div>
          </div>
        </section>

        <section>
          <h2 className="mb-4 text-xl font-semibold text-gray-800">Histórico de sessões</h2>
          {sessions.length === 0 ? (
            <div className="rounded-2xl border border-dashed border-gray-200 bg-white px-6 py-12 text-center text-gray-400">
              <p className="text-sm">Nenhuma sessão registrada ainda.</p>
              <p className="mt-1 text-xs text-gray-300">
                Clique em "Registrar nova sessão" para começar.
              </p>
            </div>
          ) : (
            <div className="flex flex-col gap-2">
              {sessions.map((s) => (
                <div
                  key={s.id}
                  className="flex items-center justify-between rounded-xl border border-gray-200 bg-white px-4 py-3"
                >
                  <div className="flex flex-wrap items-center gap-4 text-sm text-gray-700">
                    <span className="text-xs text-gray-400">{formatDate(s.session_date)}</span>
                    <span className="font-medium">
                      {s.start_page} → {s.end_page}
                    </span>
                    {s.duration_seconds != null && (
                      <span className="text-xs text-gray-400">
                        {Math.round(s.duration_seconds / 60)} min
                      </span>
                    )}
                  </div>
                  <div className="flex items-center gap-1">
                    <button
                      onClick={() => setEditingSession(s)}
                      className="rounded px-2 py-1 text-xs text-gray-500 transition-colors hover:bg-gray-100 hover:text-gray-700"
                      aria-label="Editar sessão"
                    >
                      Editar
                    </button>
                    <button
                      onClick={() => void handleDeleteSession(s.id)}
                      className="rounded px-2 py-1 text-xs text-red-400 transition-colors hover:bg-red-50 hover:text-red-600"
                      aria-label="Excluir sessão"
                    >
                      Excluir
                    </button>
                  </div>
                </div>
              ))}
            </div>
          )}
        </section>
      </main>

      {(sessionModalOpen || editingSession !== null) && (
        <SessionModal
          readingId={reading.id}
          currentPage={reading.current_page}
          totalPages={book.total_pages}
          initialSession={editingSession ?? undefined}
          onCreated={handleSessionSaved}
          onClose={() => {
            setSessionModalOpen(false)
            setEditingSession(null)
          }}
        />
      )}
    </div>
  )
}
