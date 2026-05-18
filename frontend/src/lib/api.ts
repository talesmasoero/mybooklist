import axios from 'axios'
import { getSession } from '@/lib/auth'

const api = axios.create({
  baseURL: import.meta.env.VITE_API_URL || 'http://localhost:8080',
  headers: { 'Content-Type': 'application/json' },
})

api.interceptors.request.use((config) => {
  const session = getSession()
  if (session?.token) {
    config.headers.Authorization = `Bearer ${session.token}`
  }
  return config
})

api.interceptors.response.use(
  (response) => response,
  (error) => {
    const message =
      error.response?.data?.error?.message || 'Ocorreu um erro inesperado.'
    return Promise.reject(new Error(message))
  },
)

export default api

export interface User {
  id: string
  email: string
  name: string
  created_at: string
}

export type ReadingStatus = 'want_to_read' | 'reading' | 'read' | 'abandoned'
export type AddStatus = Extract<ReadingStatus, 'want_to_read' | 'reading'>
export type BookSource = 'google_books' | 'manual'

export interface BookSearchResult {
  google_books_id: string
  title: string
  authors: string[]
  genres: string[]
  isbn?: string
  synopsis?: string
  cover_url?: string
  total_pages?: number
}

export interface Book {
  id: string
  google_books_id?: string
  title: string
  authors: string[]
  genres: string[]
  isbn?: string
  synopsis?: string
  cover_url?: string
  total_pages?: number
  source: BookSource
  created_at: string
}

export interface Reading {
  id: string
  user_id: string
  book_id: string
  status: ReadingStatus
  current_page: number
  added_at: string
  completed_at?: string
  updated_at: string
  book?: Book
}

export interface BookDataPayload {
  title: string
  authors: string[]
  genres?: string[]
  isbn?: string
  synopsis?: string
  cover_url?: string
  total_pages?: number
}

export interface AddToLibraryPayload {
  source: BookSource
  google_books_id?: string
  book_data: BookDataPayload
  status: AddStatus
}

export async function searchBooks(query: string, max = 10): Promise<BookSearchResult[]> {
  const { data } = await api.get<BookSearchResult[]>('/api/v1/books/search', {
    params: { q: query, max },
  })
  return data ?? []
}

export async function addToLibrary(payload: AddToLibraryPayload): Promise<Reading> {
  const { data } = await api.post<Reading>('/api/v1/library', payload)
  return data
}

export async function listLibrary(status?: ReadingStatus): Promise<Reading[]> {
  const { data } = await api.get<Reading[]>('/api/v1/library', {
    params: status ? { status } : undefined,
  })
  return data ?? []
}

export async function updateReadingStatus(readingId: string, status: ReadingStatus): Promise<Reading> {
  const { data } = await api.patch<Reading>(`/api/v1/library/${readingId}/status`, { status })
  return data
}

export interface Session {
  id: string
  reading_id: string
  start_page: number
  end_page: number
  duration_seconds?: number
  session_date: string
  created_at: string
}

export interface CreateSessionPayload {
  start_page: number
  end_page: number
  duration_seconds?: number
  session_date?: string
}

export async function createSession(readingId: string, payload: CreateSessionPayload): Promise<Session> {
  const { data } = await api.post<Session>(`/api/v1/readings/${readingId}/sessions`, payload)
  return data
}

export async function listSessions(readingId: string): Promise<Session[]> {
  const { data } = await api.get<Session[]>(`/api/v1/readings/${readingId}/sessions`)
  return data ?? []
}

export async function updateSession(sessionId: string, payload: Partial<CreateSessionPayload>): Promise<Session> {
  const { data } = await api.patch<Session>(`/api/v1/sessions/${sessionId}`, payload)
  return data
}

export async function deleteSession(sessionId: string): Promise<void> {
  await api.delete(`/api/v1/sessions/${sessionId}`)
}

export async function getProfile(): Promise<User> {
  const { data } = await api.get<User>('/api/v1/me')
  return data
}

export async function updateName(name: string): Promise<User> {
  const { data } = await api.patch<User>('/api/v1/me', { name })
  return data
}

export async function updatePassword(currentPassword: string, newPassword: string): Promise<void> {
  await api.patch('/api/v1/me/password', { current_password: currentPassword, new_password: newPassword })
}

export async function deleteAccount(currentPassword: string): Promise<void> {
  await api.delete('/api/v1/me', { data: { current_password: currentPassword } })
}
