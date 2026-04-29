import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { Server, Loader2 } from 'lucide-react'
import { toast } from 'sonner'
import axios from 'axios'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { useAuth } from '@/hooks/useAuth'

export default function LoginPage() {
  const { user, login } = useAuth()
  const navigate = useNavigate()
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    if (user) navigate('/', { replace: true })
  }, [user, navigate])

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    if (!username || !password) return

    setLoading(true)
    try {
      await login(username, password)
      navigate('/', { replace: true })
    } catch (err) {
      if (axios.isAxiosError(err)) {
        const status = err.response?.status
        if (status === 401) toast.error('Invalid username or password')
        else if (status === 429) toast.error('Too many attempts — try again in 5 minutes')
        else toast.error('Login failed. Please try again.')
      } else {
        toast.error('An unexpected error occurred')
      }
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="min-h-screen bg-zinc-950 flex items-center justify-center p-4">
      <div className="w-full max-w-sm">
        {/* Logo */}
        <div className="flex flex-col items-center mb-8">
          <div className="w-12 h-12 rounded-2xl bg-indigo-600 flex items-center justify-center mb-4">
            <Server className="w-6 h-6 text-white" />
          </div>
          <h1 className="text-xl font-bold text-zinc-100">Backup System</h1>
          <p className="text-sm text-zinc-500 mt-1">Sign in to your account</p>
        </div>

        {/* Card */}
        <div className="rounded-xl border border-zinc-800 bg-zinc-900 p-6">
          <form onSubmit={handleSubmit} className="space-y-4">
            <div className="space-y-1.5">
              <Label htmlFor="username">Username</Label>
              <Input
                id="username"
                type="text"
                autoComplete="username"
                autoFocus
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                placeholder="Enter username"
                disabled={loading}
              />
            </div>
            <div className="space-y-1.5">
              <Label htmlFor="password">Password</Label>
              <Input
                id="password"
                type="password"
                autoComplete="current-password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                placeholder="Enter password"
                disabled={loading}
              />
            </div>
            <Button type="submit" className="w-full" disabled={loading || !username || !password}>
              {loading ? (
                <>
                  <Loader2 className="animate-spin" />
                  Signing in…
                </>
              ) : (
                'Sign in'
              )}
            </Button>
          </form>
        </div>
      </div>
    </div>
  )
}
