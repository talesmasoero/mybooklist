const SESSION_KEY = 'mbl_session'

export interface SessionUser {
  id: string
  email: string
  name: string
}

interface Session {
  token: string
  refreshToken: string
  user: SessionUser
}

export function saveSession(token: string, refreshToken: string, user: SessionUser): void {
  localStorage.setItem(SESSION_KEY, JSON.stringify({ token, refreshToken, user }))
}

export function getSession(): Session | null {
  const raw = localStorage.getItem(SESSION_KEY)
  if (!raw) return null
  try {
    return JSON.parse(raw) as Session
  } catch {
    return null
  }
}

export function clearSession(): void {
  localStorage.removeItem(SESSION_KEY)
}

export function isAuthenticated(): boolean {
  return getSession() !== null
}
