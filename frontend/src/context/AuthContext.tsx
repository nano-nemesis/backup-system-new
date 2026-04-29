import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useRef,
  useState,
  type ReactNode,
} from 'react'
import api from '@/lib/axios'
import { queryClient } from '@/lib/queryClient'
import type { User } from '@/types'

interface AuthState {
  user: User | null
  isLoading: boolean
}

interface AuthContextValue extends AuthState {
  login: (username: string, password: string) => Promise<void>
  logout: () => Promise<void>
  setUser: (user: User | null) => void
}

const AuthContext = createContext<AuthContextValue | null>(null)

export function AuthProvider({ children }: { children: ReactNode }) {
  const [state, setState] = useState<AuthState>({ user: null, isLoading: true })
  const initialized = useRef(false)

  const setUser = useCallback((user: User | null) => {
    setState({ user, isLoading: false })
  }, [])

  const logout = useCallback(async () => {
    try {
      await api.post('/logout')
    } catch {
      // ignore — we always clear state
    }
    queryClient.clear()
    setState({ user: null, isLoading: false })
  }, [])

  // Bootstrap: check current session
  useEffect(() => {
    if (initialized.current) return
    initialized.current = true

    api
      .get<{ user: User }>('/me')
      .then((res) => setState({ user: res.data.user, isLoading: false }))
      .catch(() => setState({ user: null, isLoading: false }))
  }, [])

  // Global 401 handler
  useEffect(() => {
    const handle = () => {
      queryClient.clear()
      setState({ user: null, isLoading: false })
    }
    window.addEventListener('auth:unauthorized', handle)
    return () => window.removeEventListener('auth:unauthorized', handle)
  }, [])

  const login = useCallback(async (username: string, password: string) => {
    const res = await api.post<{ user: User }>('/login', { username, password })
    setState({ user: res.data.user, isLoading: false })
  }, [])

  return (
    <AuthContext.Provider value={{ ...state, login, logout, setUser }}>
      {children}
    </AuthContext.Provider>
  )
}

export function useAuthContext() {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error('useAuthContext must be used inside AuthProvider')
  return ctx
}
