import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { Server, Loader2 } from 'lucide-react'
import { isAxiosError } from 'axios'
import { toast } from 'sonner'
import api from '@/lib/axios'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'

export default function SetupPage() {
  const navigate = useNavigate()
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [confirm, setConfirm]   = useState('')
  const [loading, setLoading]   = useState(false)
  const [checking, setChecking] = useState(true)

  useEffect(() => {
    api.get<{ needs_setup: boolean }>('/setup/status')
      .then(r => { if (!r.data.needs_setup) navigate('/login', { replace: true }) })
      .catch(() => navigate('/login', { replace: true }))
      .finally(() => setChecking(false))
  }, [navigate])

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    if (password !== confirm) { toast.error('Passwords do not match'); return }
    if (password.length < 8)  { toast.error('Password must be at least 8 characters'); return }
    setLoading(true)
    try {
      await api.post('/setup', { username, password })
      toast.success('Admin account created — please sign in')
      navigate('/login', { replace: true })
    } catch (err) {
      toast.error(isAxiosError(err) ? (err.response?.data?.error ?? 'Setup failed') : 'An unexpected error occurred')
    } finally {
      setLoading(false)
    }
  }

  if (checking) {
    return (
      <div className="min-h-screen bg-bg flex items-center justify-center">
        <Loader2 className="w-6 h-6 animate-spin text-indigo-500" />
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-bg flex items-center justify-center p-4">
      <div className="w-full max-w-sm">
        <div className="flex flex-col items-center mb-8">
          <div className="w-12 h-12 rounded-2xl bg-indigo-600 flex items-center justify-center mb-4">
            <Server className="w-6 h-6 text-white" />
          </div>
          <h1 className="text-xl font-bold text-fg">Initial Setup</h1>
          <p className="text-sm text-muted mt-1">Create your admin account to get started</p>
        </div>

        <div className="rounded-xl border border-line bg-surface p-6">
          <form onSubmit={handleSubmit} className="space-y-4">
            <div className="space-y-1.5">
              <Label htmlFor="username">Username</Label>
              <Input
                id="username"
                autoFocus
                value={username}
                onChange={e => setUsername(e.target.value)}
                placeholder="admin"
                disabled={loading}
              />
            </div>
            <div className="space-y-1.5">
              <Label htmlFor="password">Password</Label>
              <Input
                id="password"
                type="password"
                value={password}
                onChange={e => setPassword(e.target.value)}
                placeholder="Min. 8 characters"
                disabled={loading}
              />
            </div>
            <div className="space-y-1.5">
              <Label htmlFor="confirm">Confirm Password</Label>
              <Input
                id="confirm"
                type="password"
                value={confirm}
                onChange={e => setConfirm(e.target.value)}
                placeholder="Repeat password"
                disabled={loading}
              />
            </div>
            <Button
              type="submit"
              className="w-full"
              disabled={loading || !username || !password || !confirm}
            >
              {loading
                ? <><Loader2 className="animate-spin" />Creating account…</>
                : 'Create admin account'}
            </Button>
          </form>
        </div>
      </div>
    </div>
  )
}
