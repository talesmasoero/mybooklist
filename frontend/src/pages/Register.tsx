import { useState, useEffect, type FormEvent } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import api from '@/lib/api'
import { listSecurityQuestions, type SecurityQuestion } from '@/lib/api'
import { saveSession, type SessionUser } from '@/lib/auth'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { TermsModal } from '@/components/TermsModal'

interface RegisterResponse {
  access_token: string
  refresh_token: string
  user: SessionUser
}

interface AnswerBlock {
  questionId: number
  answer: string
}

const emptyBlock = (): AnswerBlock => ({ questionId: 0, answer: '' })

export function Register() {
  const navigate = useNavigate()
  const [name, setName] = useState('')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [consented, setConsented] = useState(false)
  const [showTerms, setShowTerms] = useState(false)
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  const [questions, setQuestions] = useState<SecurityQuestion[]>([])
  const [blocks, setBlocks] = useState<AnswerBlock[]>([emptyBlock(), emptyBlock(), emptyBlock()])

  useEffect(() => {
    listSecurityQuestions().then(setQuestions).catch(() => {})
  }, [])

  function updateBlock(index: number, field: keyof AnswerBlock, value: string | number) {
    setBlocks((prev) => prev.map((b, i) => (i === index ? { ...b, [field]: value } : b)))
  }

  function getAvailableQuestions(blockIndex: number): SecurityQuestion[] {
    const selectedIds = blocks
      .map((b, i) => (i !== blockIndex ? b.questionId : 0))
      .filter((id) => id !== 0)
    return questions.filter((q) => !selectedIds.includes(q.id))
  }

  function buildSecurityAnswers() {
    return blocks
      .filter((b) => b.questionId !== 0 && b.answer.trim().length >= 2)
      .map((b) => ({ question_id: b.questionId, answer: b.answer.trim() }))
  }

  function validateAnswers(): string | null {
    const block1 = blocks[0]
    const block2 = blocks[1]
    const block3 = blocks[2]

    if (block1.questionId === 0 || block1.answer.trim().length < 2) {
      return 'Selecione a pergunta 1 e escreva uma resposta com pelo menos 2 caracteres.'
    }
    if (block2.questionId === 0 || block2.answer.trim().length < 2) {
      return 'Selecione a pergunta 2 e escreva uma resposta com pelo menos 2 caracteres.'
    }
    if (block3.questionId !== 0 || block3.answer.trim() !== '') {
      if (block3.questionId === 0 || block3.answer.trim().length < 2) {
        return 'Se preenchida, a pergunta 3 precisa ter uma pergunta selecionada e resposta com pelo menos 2 caracteres.'
      }
    }

    const selectedIds = blocks.filter((b) => b.questionId !== 0).map((b) => b.questionId)
    if (new Set(selectedIds).size !== selectedIds.length) {
      return 'As perguntas de segurança devem ser diferentes entre si.'
    }
    return null
  }

  async function handleSubmit(e: FormEvent) {
    e.preventDefault()
    if (!consented) {
      setError('Você precisa aceitar os termos de uso para criar uma conta.')
      return
    }

    const answerError = validateAnswers()
    if (answerError) {
      setError(answerError)
      return
    }

    setError('')
    setLoading(true)
    try {
      const { data } = await api.post<RegisterResponse>('/api/v1/auth/register', {
        name,
        email,
        password,
        security_answers: buildSecurityAnswers(),
      })
      saveSession(data.access_token, data.refresh_token, data.user)
      navigate('/home', { replace: true })
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Erro ao criar conta.')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="flex min-h-screen items-center justify-center bg-gray-50 px-4 py-8">
      <div className="w-full max-w-md rounded-2xl bg-white p-8 shadow-sm ring-1 ring-gray-100">
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

          {/* Seção de perguntas de segurança */}
          <div className="rounded-lg border border-gray-200 p-4 space-y-4">
            <div>
              <p className="text-sm font-medium text-gray-800">Perguntas de segurança</p>
              <p className="mt-1 text-xs text-gray-500">
                Cadastre 2 ou 3 perguntas para recuperar sua conta caso esqueça a senha.
              </p>
            </div>

            {[0, 1, 2].map((i) => (
              <div key={i} className="space-y-2">
                <p className="text-xs font-medium text-gray-600">
                  Pergunta {i + 1}{i < 2 ? ' *' : ' (opcional)'}
                </p>
                <select
                  value={blocks[i].questionId}
                  onChange={(e) => updateBlock(i, 'questionId', Number(e.target.value))}
                  className="w-full rounded-lg border border-gray-300 bg-white px-3 py-2 text-sm text-gray-700 focus:border-[#162447] focus:outline-none focus:ring-2 focus:ring-[#162447]/20"
                >
                  <option value={0}>Selecione uma pergunta</option>
                  {getAvailableQuestions(i).map((q) => (
                    <option key={q.id} value={q.id}>{q.text}</option>
                  ))}
                  {blocks[i].questionId !== 0 && !getAvailableQuestions(i).find(q => q.id === blocks[i].questionId) && (
                    <option value={blocks[i].questionId}>
                      {questions.find(q => q.id === blocks[i].questionId)?.text}
                    </option>
                  )}
                </select>
                <input
                  type="text"
                  placeholder="Sua resposta"
                  value={blocks[i].answer}
                  onChange={(e) => updateBlock(i, 'answer', e.target.value)}
                  className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm text-gray-700 placeholder-gray-400 focus:border-[#162447] focus:outline-none focus:ring-2 focus:ring-[#162447]/20"
                />
              </div>
            ))}
          </div>

          <label className="flex items-start gap-3 text-sm text-gray-600">
            <input
              type="checkbox"
              checked={consented}
              onChange={(e) => setConsented(e.target.checked)}
              className="mt-0.5 h-4 w-4 rounded border-gray-300 accent-[#162447]"
            />
            <span>
              Li e aceito os{' '}
              <button
                type="button"
                className="font-medium text-[#162447] underline"
                onClick={() => setShowTerms(true)}
              >
                termos de uso
              </button>
              {' '}e a{' '}
              <button
                type="button"
                className="font-medium text-[#162447] underline"
                onClick={() => setShowTerms(true)}
              >
                política de privacidade
              </button>
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

      <TermsModal
        open={showTerms}
        onClose={() => setShowTerms(false)}
        onAccept={() => { setConsented(true); setShowTerms(false) }}
      />
    </div>
  )
}
