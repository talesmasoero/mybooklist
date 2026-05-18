import { useState, useEffect, useRef } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { getSession, saveSession, clearSession } from '@/lib/auth'
import { getProfile, updateName, updatePassword, deleteAccount } from '@/lib/api'
import type { User } from '@/lib/api'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'

function useSuccessMessage() {
  const [message, setMessage] = useState<string | null>(null)
  const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  function show(msg: string) {
    if (timerRef.current) clearTimeout(timerRef.current)
    setMessage(msg)
    timerRef.current = setTimeout(() => setMessage(null), 3000)
  }

  return { message, show }
}

export function Profile() {
  const navigate = useNavigate()
  const session = getSession()

  const [user, setUser] = useState<User | null>(null)
  const [loadError, setLoadError] = useState<string | null>(null)

  // Name section
  const [name, setName] = useState('')
  const [nameError, setNameError] = useState<string | null>(null)
  const [nameSaving, setNameSaving] = useState(false)
  const nameSuccess = useSuccessMessage()

  // Password section
  const [currentPwd, setCurrentPwd] = useState('')
  const [newPwd, setNewPwd] = useState('')
  const [confirmPwd, setConfirmPwd] = useState('')
  const [pwdError, setPwdError] = useState<string | null>(null)
  const [pwdSaving, setPwdSaving] = useState(false)
  const pwdSuccess = useSuccessMessage()

  // Delete section
  const [deleteModalOpen, setDeleteModalOpen] = useState(false)
  const [deletePwd, setDeletePwd] = useState('')
  const [deleteChecked, setDeleteChecked] = useState(false)
  const [deleteError, setDeleteError] = useState<string | null>(null)
  const [deleteLoading, setDeleteLoading] = useState(false)

  useEffect(() => {
    getProfile()
      .then((u) => {
        setUser(u)
        setName(u.name)
      })
      .catch((err: Error) => setLoadError(err.message))
  }, [])

  function handleLogout() {
    clearSession()
    navigate('/')
  }

  async function handleNameSave(e: React.FormEvent) {
    e.preventDefault()
    setNameError(null)
    setNameSaving(true)
    try {
      const updated = await updateName(name)
      setUser(updated)
      setName(updated.name)
      if (session) {
        saveSession(session.token, session.refreshToken, {
          id: updated.id,
          email: updated.email,
          name: updated.name,
        })
      }
      nameSuccess.show('Nome atualizado com sucesso.')
    } catch (err: unknown) {
      setNameError(err instanceof Error ? err.message : 'Erro ao atualizar nome.')
    } finally {
      setNameSaving(false)
    }
  }

  async function handlePasswordChange(e: React.FormEvent) {
    e.preventDefault()
    setPwdError(null)

    if (newPwd.length < 8) {
      setPwdError('A nova senha deve ter pelo menos 8 caracteres.')
      return
    }
    if (newPwd !== confirmPwd) {
      setPwdError('A nova senha e a confirmação não coincidem.')
      return
    }

    setPwdSaving(true)
    try {
      await updatePassword(currentPwd, newPwd)
      setCurrentPwd('')
      setNewPwd('')
      setConfirmPwd('')
      pwdSuccess.show('Senha alterada com sucesso.')
    } catch (err: unknown) {
      setPwdError(err instanceof Error ? err.message : 'Erro ao alterar senha.')
    } finally {
      setPwdSaving(false)
    }
  }

  function openDeleteModal() {
    setDeletePwd('')
    setDeleteChecked(false)
    setDeleteError(null)
    setDeleteModalOpen(true)
  }

  async function handleDeleteAccount() {
    setDeleteError(null)
    setDeleteLoading(true)
    try {
      await deleteAccount(deletePwd)
      clearSession()
      navigate('/')
    } catch (err: unknown) {
      setDeleteError(err instanceof Error ? err.message : 'Erro ao excluir conta.')
      setDeleteLoading(false)
    }
  }

  const userName = session?.user.name ?? user?.name ?? 'Leitor'

  if (loadError) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <p className="text-red-500">{loadError}</p>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-gray-50">
      <header className="bg-white shadow-sm">
        <div className="mx-auto flex max-w-7xl items-center justify-between px-4 py-4">
          <Link to="/home" className="text-lg font-bold text-[#162447] hover:opacity-80 transition-opacity">
            MyBookList
          </Link>
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

      <main className="mx-auto max-w-2xl px-4 py-10 space-y-8">
        <h1 className="text-2xl font-bold text-[#162447]">Meu perfil</h1>

        {/* Informações pessoais */}
        <section className="bg-white rounded-xl shadow-sm p-6">
          <h2 className="text-lg font-semibold text-gray-800 mb-5">Informações pessoais</h2>
          <form onSubmit={handleNameSave} className="space-y-4">
            <Input
              label="Nome"
              value={name}
              onChange={(e) => setName(e.target.value)}
              error={nameError ?? undefined}
            />
            <Input
              label="Email"
              value={user?.email ?? ''}
              readOnly
              className="bg-gray-50 cursor-not-allowed"
            />
            {nameSuccess.message && (
              <p className="text-sm text-green-600">{nameSuccess.message}</p>
            )}
            <div className="flex justify-end">
              <Button type="submit" loading={nameSaving}>
                Salvar
              </Button>
            </div>
          </form>
        </section>

        {/* Alterar senha */}
        <section className="bg-white rounded-xl shadow-sm p-6">
          <h2 className="text-lg font-semibold text-gray-800 mb-5">Alterar senha</h2>
          <form onSubmit={handlePasswordChange} className="space-y-4">
            <Input
              label="Senha atual"
              type="password"
              value={currentPwd}
              onChange={(e) => setCurrentPwd(e.target.value)}
            />
            <Input
              label="Nova senha"
              type="password"
              value={newPwd}
              onChange={(e) => setNewPwd(e.target.value)}
            />
            <Input
              label="Confirmar nova senha"
              type="password"
              value={confirmPwd}
              onChange={(e) => setConfirmPwd(e.target.value)}
              error={pwdError ?? undefined}
            />
            {pwdSuccess.message && (
              <p className="text-sm text-green-600">{pwdSuccess.message}</p>
            )}
            <div className="flex justify-end">
              <Button type="submit" loading={pwdSaving}>
                Alterar senha
              </Button>
            </div>
          </form>
        </section>

        {/* Zona de perigo */}
        <section className="bg-white rounded-xl shadow-sm p-6 border border-red-200">
          <h2 className="text-lg font-semibold text-red-600 mb-2">Zona de perigo</h2>
          <p className="text-sm text-gray-600 mb-5">
            Excluir sua conta é uma ação permanente e irreversível. Todos os seus dados pessoais
            serão removidos conforme a LGPD (Lei 13.709/2018).
          </p>
          <Button
            type="button"
            onClick={openDeleteModal}
            className="bg-red-600 hover:bg-red-700 focus:ring-red-500 text-white border-0"
            variant="primary"
          >
            Excluir minha conta
          </Button>
        </section>
      </main>

      {/* Modal de confirmação de exclusão */}
      {deleteModalOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 px-4">
          <div className="bg-white rounded-xl shadow-xl w-full max-w-md p-6 space-y-5">
            <h3 className="text-lg font-semibold text-red-600">Excluir conta</h3>
            <p className="text-sm text-gray-600">
              Esta ação não pode ser desfeita. Para confirmar, informe sua senha atual.
            </p>
            <Input
              label="Senha atual"
              type="password"
              value={deletePwd}
              onChange={(e) => setDeletePwd(e.target.value)}
              error={deleteError ?? undefined}
            />
            <label className="flex items-start gap-3 cursor-pointer">
              <input
                type="checkbox"
                checked={deleteChecked}
                onChange={(e) => setDeleteChecked(e.target.checked)}
                className="mt-0.5 h-4 w-4 rounded border-gray-300 text-red-600 focus:ring-red-500"
              />
              <span className="text-sm text-gray-700">
                Confirmo que entendo que esta ação é irreversível.
              </span>
            </label>
            <div className="flex justify-end gap-3 pt-2">
              <Button
                type="button"
                variant="secondary"
                onClick={() => setDeleteModalOpen(false)}
                disabled={deleteLoading}
              >
                Cancelar
              </Button>
              <button
                type="button"
                onClick={handleDeleteAccount}
                disabled={!deleteChecked || !deletePwd || deleteLoading}
                className="inline-flex items-center justify-center rounded-lg px-6 py-3 font-medium transition-colors bg-red-600 text-white hover:bg-red-700 focus:outline-none focus:ring-2 focus:ring-red-500 focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {deleteLoading ? 'Excluindo…' : 'Excluir permanentemente'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
