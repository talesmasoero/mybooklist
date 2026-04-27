import { Link } from 'react-router-dom'
import { Button } from '@/components/ui/Button'

export function Landing() {
  return (
    <div className="flex min-h-screen flex-col bg-white">
      <nav className="mx-auto flex w-full max-w-7xl items-center justify-between px-4 py-5">
        <span className="text-xl font-bold text-[#162447]">MyBookList</span>
        <div className="flex items-center gap-3">
          <Link to="/login">
            <Button variant="ghost" className="px-4 py-2 text-sm">Entrar</Button>
          </Link>
          <Link to="/register">
            <Button variant="primary" className="px-4 py-2 text-sm">Cadastrar</Button>
          </Link>
        </div>
      </nav>

      <main className="mx-auto flex w-full max-w-7xl flex-1 flex-col items-center justify-center px-4 py-20 text-center">
        <h1 className="text-5xl font-bold leading-tight tracking-tight text-[#162447] sm:text-6xl">
          Sua jornada de leitura,<br />do seu jeito.
        </h1>
        <p className="mt-6 max-w-xl text-lg text-gray-500">
          Registre sessões, anote reflexões no momento certo e construa o hábito da leitura com metas e progresso visível.
        </p>
        <div className="mt-10 flex flex-col gap-4 sm:flex-row">
          <Link to="/register">
            <Button variant="primary" className="px-8 py-4 text-base">
              Começar agora — é grátis
            </Button>
          </Link>
          <Link to="/login">
            <Button variant="secondary" className="px-8 py-4 text-base">
              Já tenho conta
            </Button>
          </Link>
        </div>
      </main>

      <footer className="py-8 text-center text-sm text-gray-400">
        &copy; {new Date().getFullYear()} MyBookList
      </footer>
    </div>
  )
}
