import { useNavigate } from 'react-router-dom'
import { getSession, clearSession } from '@/lib/auth'
import { Button } from '@/components/ui/Button'

export function Home() {
  const navigate = useNavigate()
  const session = getSession()
  const userName = session?.user.name ?? 'Leitor'

  function handleLogout() {
    clearSession()
    navigate('/', { replace: true })
  }

  return (
    <div className="min-h-screen bg-gray-50">
      <header className="bg-white shadow-sm">
        <div className="mx-auto flex max-w-7xl items-center justify-between px-4 py-4">
          <span className="text-lg font-bold text-[#162447]">MyBookList</span>
          <div className="flex items-center gap-4 text-sm text-gray-600">
            <span>Olá, <strong className="text-[#162447]">{userName}</strong></span>
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
          <h2 className="mb-4 text-xl font-semibold text-gray-800">Continuar lendo</h2>
          <div className="rounded-2xl border border-dashed border-gray-200 bg-white px-6 py-12 text-center text-gray-400">
            <p className="text-sm">Você ainda não está lendo nenhum livro.</p>
            <p className="mt-1 text-xs text-gray-300">Em breve: adicionar livro à biblioteca</p>
            <Button variant="secondary" className="mt-6 px-5 py-2.5 text-sm" disabled>
              + Adicionar livro
            </Button>
          </div>
        </section>

        <section>
          <h2 className="mb-4 text-xl font-semibold text-gray-800">Meta anual</h2>
          <div className="rounded-2xl border border-dashed border-gray-200 bg-white px-6 py-10 text-center text-gray-400">
            <p className="text-sm">Defina sua meta de leitura para {new Date().getFullYear()}.</p>
            <p className="mt-1 text-xs text-gray-300">Em breve</p>
          </div>
        </section>
      </main>
    </div>
  )
}
