import { useState, type FormEvent } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import api from '@/lib/api'
import { saveSession, type SessionUser } from '@/lib/auth'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'

interface RegisterResponse {
  access_token: string
  refresh_token: string
  user: SessionUser
}

export function Register() {
  const navigate = useNavigate()
  const [name, setName] = useState('')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [consented, setConsented] = useState(false)
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  async function handleSubmit(e: FormEvent) {
    e.preventDefault()
    if (!consented) {
      setError('Você precisa aceitar os termos de uso para criar uma conta.')
      return
    }
    setError('')
    setLoading(true)
    try {
      const { data } = await api.post<RegisterResponse>('/api/v1/auth/register', { name, email, password })
      saveSession(data.access_token, data.refresh_token, data.user)
      navigate('/home', { replace: true })
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Erro ao criar conta.')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="flex min-h-screen items-center justify-center bg-gray-50 px-4">
      <div className="w-full max-w-sm rounded-2xl bg-white p-8 shadow-sm ring-1 ring-gray-100">
        <div className="mb-8 text-center">
          <Link to="/" className="text-xl font-bold text-[#162447]">MyBookList</Link>
          <p className="mt-2 text-sm text-gray-500">Crie sua conta</p>
        </div>

        {error && (
          <div className="mb-4 rounded-lg bg-red-50 px-4 py-3 text-sm text-red-600">
            {error}
          </div>
        )}

        <form onSubmit={handleSubmit} className="flex flex-col gap-4">
          <Input
            label="Nome"
            type="text"
            placeholder="Seu nome"
            value={name}
            onChange={(e) => setName(e.target.value)}
            required
          />
          <Input
            label="E-mail"
            type="email"
            placeholder="voce@exemplo.com"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            required
          />
          <Input
            label="Senha"
            type="password"
            placeholder="Mínimo 8 caracteres"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
          />

          <label className="flex items-start gap-3 text-sm text-gray-600">
            <input
              type="checkbox"
              checked={consented}
              onChange={(e) => setConsented(e.target.checked)}
              className="mt-0.5 h-4 w-4 rounded border-gray-300 accent-[#162447]"
            />
            <span>
              Li e aceito os{' '}
              <span className="font-medium text-[#162447]">termos de uso</span>
              {' '}e a{' '}
              <span className="font-medium text-[#162447]">política de privacidade</span>
            </span>
          </label>

          <Button type="submit" loading={loading} className="mt-2 w-full">
            Criar conta
          </Button>
        </form>

        <p className="mt-6 text-center text-sm text-gray-500">
          Já tem conta?{' '}
          <Link to="/login" className="font-medium text-[#162447] hover:underline">
            Entrar
          </Link>
        </p>
      </div>
    </div>
  )
}
