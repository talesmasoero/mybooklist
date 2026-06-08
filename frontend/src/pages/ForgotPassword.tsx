import { useState } from 'react'
import { Link } from 'react-router-dom'
import {
  forgotPassword, verifySecurityAnswers, resetPassword,
  type SecurityQuestion,
} from '@/lib/api'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'

type Step = 'email' | 'questions' | 'reset' | 'done'

export function ForgotPassword() {
  const [step, setStep] = useState<Step>('email')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  // email step
  const [email, setEmail] = useState('')
  const [genericMsg, setGenericMsg] = useState('')

  // questions step
  const [userId, setUserId] = useState('')
  const [questions, setQuestions] = useState<SecurityQuestion[]>([])
  const [answers, setAnswers] = useState<Record<number, string>>({})

  // reset step
  const [resetToken, setResetToken] = useState('')
  const [newPassword, setNewPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')

  async function handleEmailSubmit(e: React.FormEvent) {
    e.preventDefault()
    setError('')
    setGenericMsg('')
    setLoading(true)
    try {
      const result = await forgotPassword(email)
      if (result.questions.length >= 2) {
        setUserId(result.user_id)
        setQuestions(result.questions)
        setAnswers(Object.fromEntries(result.questions.map((q) => [q.id, ''])))
        setStep('questions')
      } else {
        setGenericMsg(
          'Se este e-mail existir e tiver perguntas de segurança configuradas, elas aparecerão aqui. Verifique seu e-mail cadastrado.'
        )
      }
    } catch {
      setGenericMsg(
        'Se este e-mail existir e tiver perguntas de segurança configuradas, elas aparecerão aqui. Verifique seu e-mail cadastrado.'
      )
    } finally {
      setLoading(false)
    }
  }

  async function handleVerifySubmit(e: React.FormEvent) {
    e.preventDefault()
    setError('')
    setLoading(true)
    try {
      const payload = Object.entries(answers)
        .filter(([, ans]) => ans.trim() !== '')
        .map(([id, ans]) => ({ question_id: Number(id), answer: ans.trim() }))

      const result = await verifySecurityAnswers(userId, payload)
      setResetToken(result.reset_token)
      setStep('reset')
    } catch {
      setError('Respostas incorretas. Tente novamente.')
    } finally {
      setLoading(false)
    }
  }

  async function handleResetSubmit(e: React.FormEvent) {
    e.preventDefault()
    setError('')

    if (newPassword.length < 8) {
      setError('A nova senha deve ter pelo menos 8 caracteres.')
      return
    }
    if (newPassword !== confirmPassword) {
      setError('A nova senha e a confirmação não coincidem.')
      return
    }

    setLoading(true)
    try {
      await resetPassword(resetToken, newPassword)
      setStep('done')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Erro ao redefinir senha. O link pode ter expirado.')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="flex min-h-screen items-center justify-center bg-gray-50 px-4">
      <div className="w-full max-w-sm rounded-2xl bg-white p-8 shadow-sm ring-1 ring-gray-100">
        <div className="mb-8 text-center">
          <Link to="/" className="text-xl font-bold text-[#162447]">MyBookList</Link>
          <p className="mt-2 text-sm text-gray-500">Recuperar senha</p>
        </div>

        {/* ── Etapa: e-mail ─────────────────────────────────────────── */}
        {step === 'email' && (
          <>
            <p className="mb-4 text-sm text-gray-600">
              Informe seu e-mail para buscar suas perguntas de segurança.
            </p>
            {genericMsg && (
              <div className="mb-4 rounded-lg bg-blue-50 px-4 py-3 text-sm text-blue-700">
                {genericMsg}
              </div>
            )}
            {error && (
              <div className="mb-4 rounded-lg bg-red-50 px-4 py-3 text-sm text-red-600">
                {error}
              </div>
            )}
            <form onSubmit={handleEmailSubmit} className="flex flex-col gap-4">
              <Input
                label="E-mail"
                type="email"
                placeholder="voce@exemplo.com"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                required
              />
              <Button type="submit" loading={loading} className="w-full">
                Continuar
              </Button>
            </form>
          </>
        )}

        {/* ── Etapa: perguntas ──────────────────────────────────────── */}
        {step === 'questions' && (
          <>
            <p className="mb-4 text-sm text-gray-600">
              Responda pelo menos 2 perguntas corretamente para recuperar o acesso.
            </p>
            {error && (
              <div className="mb-4 rounded-lg bg-red-50 px-4 py-3 text-sm text-red-600">
                {error}
              </div>
            )}
            <form onSubmit={handleVerifySubmit} className="flex flex-col gap-4">
              {questions.map((q) => (
                <div key={q.id}>
                  <label className="block text-xs font-medium text-gray-600 mb-1">{q.text}</label>
                  <input
                    type="text"
                    placeholder="Sua resposta"
                    value={answers[q.id] ?? ''}
                    onChange={(e) => setAnswers((prev) => ({ ...prev, [q.id]: e.target.value }))}
                    className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm text-gray-700 placeholder-gray-400 focus:border-[#162447] focus:outline-none focus:ring-2 focus:ring-[#162447]/20"
                  />
                </div>
              ))}
              <Button type="submit" loading={loading} className="w-full">
                Verificar
              </Button>
            </form>
          </>
        )}

        {/* ── Etapa: nova senha ─────────────────────────────────────── */}
        {step === 'reset' && (
          <>
            <p className="mb-4 text-sm text-gray-600">
              Defina sua nova senha. Ela deve ter pelo menos 8 caracteres.
            </p>
            {error && (
              <div className="mb-4 rounded-lg bg-red-50 px-4 py-3 text-sm text-red-600">
                {error}
              </div>
            )}
            <form onSubmit={handleResetSubmit} className="flex flex-col gap-4">
              <Input
                label="Nova senha"
                type="password"
                placeholder="Mínimo 8 caracteres"
                value={newPassword}
                onChange={(e) => setNewPassword(e.target.value)}
                required
              />
              <Input
                label="Confirmar nova senha"
                type="password"
                placeholder="Repita a nova senha"
                value={confirmPassword}
                onChange={(e) => setConfirmPassword(e.target.value)}
                required
              />
              <Button type="submit" loading={loading} className="w-full">
                Definir nova senha
              </Button>
            </form>
          </>
        )}

        {/* ── Etapa: concluído ──────────────────────────────────────── */}
        {step === 'done' && (
          <div className="text-center space-y-4">
            <p className="text-green-600 font-medium">Senha alterada com sucesso!</p>
            <p className="text-sm text-gray-500">Você já pode entrar com sua nova senha.</p>
            <Link
              to="/login"
              className="inline-block mt-2 font-medium text-[#162447] hover:underline"
            >
              Ir para o login
            </Link>
          </div>
        )}

        {/* Botão "Voltar para login" persistente */}
        {step !== 'done' && (
          <p className="mt-6 text-center text-sm text-gray-500">
            <Link to="/login" className="font-medium text-[#162447] hover:underline">
              Voltar para o login
            </Link>
          </p>
        )}
      </div>
    </div>
  )
}
