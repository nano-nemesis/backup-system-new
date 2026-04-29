import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { UserPlus, Trash2, KeyRound, ShieldCheck, Loader2 } from 'lucide-react'
import { isAxiosError } from 'axios'
import { toast } from 'sonner'
import api from '@/lib/axios'
import { useAuth } from '@/hooks/useAuth'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogClose } from '@/components/ui/dialog'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Table, TableHeader, TableBody, TableRow, TableHead, TableCell } from '@/components/ui/table'
import { formatDateTime } from '@/lib/utils'
import type { AdminUser, AdminUsersResponse, Role } from '@/types'

// ─── API helpers ─────────────────────────────────────────────────────────────

const fetchUsers  = () => api.get<AdminUsersResponse>('/admin/users').then(r => r.data)
const createUser  = (body: { username: string; password: string; role: Role }) => api.post('/admin/users', body)
const deleteUser  = (id: string) => api.delete(`/admin/users/${id}`)
const updateRole  = (id: string, role: Role) => api.patch(`/admin/users/${id}/role`, { role })
const updatePassword = (id: string, password: string) => api.patch(`/admin/users/${id}/password`, { password })

function apiError(err: unknown, fallback: string): string {
  return isAxiosError(err) ? (err.response?.data?.error ?? fallback) : fallback
}

// ─── Create user dialog ──────────────────────────────────────────────────────

function CreateUserDialog({ open, onClose }: { open: boolean; onClose: () => void }) {
  const qc = useQueryClient()
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [role, setRole] = useState<Role>('viewer')

  const mut = useMutation({
    mutationFn: () => createUser({ username, password, role }),
    onSuccess: () => {
      toast.success('User created')
      qc.invalidateQueries({ queryKey: ['admin-users'] })
      setUsername(''); setPassword(''); setRole('viewer')
      onClose()
    },
    onError: err => toast.error(apiError(err, 'Failed to create user')),
  })

  return (
    <Dialog open={open} onOpenChange={v => !v && onClose()}>
      <DialogContent>
        <DialogHeader><DialogTitle>Create User</DialogTitle></DialogHeader>
        <div className="space-y-4">
          <div className="space-y-1.5">
            <Label>Username</Label>
            <Input value={username} onChange={e => setUsername(e.target.value)} placeholder="username" disabled={mut.isPending} />
          </div>
          <div className="space-y-1.5">
            <Label>Password</Label>
            <Input type="password" value={password} onChange={e => setPassword(e.target.value)} placeholder="Min. 8 characters" disabled={mut.isPending} />
          </div>
          <div className="space-y-1.5">
            <Label>Role</Label>
            <Select value={role} onValueChange={v => setRole(v as Role)} disabled={mut.isPending}>
              <SelectTrigger><SelectValue /></SelectTrigger>
              <SelectContent>
                <SelectItem value="viewer">Viewer</SelectItem>
                <SelectItem value="admin">Admin</SelectItem>
              </SelectContent>
            </Select>
          </div>
          <div className="flex gap-2 pt-1">
            <Button className="flex-1" onClick={() => mut.mutate()} disabled={!username || password.length < 8 || mut.isPending}>
              {mut.isPending ? <Loader2 className="animate-spin" /> : <UserPlus />}
              Create
            </Button>
            <DialogClose asChild>
              <Button variant="outline" className="flex-1">Cancel</Button>
            </DialogClose>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  )
}

// ─── Change password dialog ──────────────────────────────────────────────────

function ChangePasswordDialog({ user, onClose }: { user: AdminUser | null; onClose: () => void }) {
  const qc = useQueryClient()
  const [password, setPassword] = useState('')

  const mut = useMutation({
    mutationFn: () => updatePassword(user!.id, password),
    onSuccess: () => {
      toast.success('Password updated')
      qc.invalidateQueries({ queryKey: ['admin-users'] })
      setPassword('')
      onClose()
    },
    onError: err => toast.error(apiError(err, 'Failed to update password')),
  })

  return (
    <Dialog open={!!user} onOpenChange={v => !v && onClose()}>
      <DialogContent>
        <DialogHeader><DialogTitle>Change password — {user?.username}</DialogTitle></DialogHeader>
        <div className="space-y-4">
          <div className="space-y-1.5">
            <Label>New Password</Label>
            <Input type="password" value={password} onChange={e => setPassword(e.target.value)} placeholder="Min. 8 characters" disabled={mut.isPending} />
          </div>
          <div className="flex gap-2 pt-1">
            <Button className="flex-1" onClick={() => mut.mutate()} disabled={password.length < 8 || mut.isPending}>
              {mut.isPending ? <Loader2 className="animate-spin" /> : <KeyRound />}
              Update
            </Button>
            <DialogClose asChild>
              <Button variant="outline" className="flex-1">Cancel</Button>
            </DialogClose>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  )
}

// ─── Main page ────────────────────────────────────────────────────────────────

export default function AdminPage() {
  const { user: me } = useAuth()
  const qc = useQueryClient()
  const [showCreate, setShowCreate] = useState(false)
  const [pwUser, setPwUser] = useState<AdminUser | null>(null)

  const { data, isLoading } = useQuery({ queryKey: ['admin-users'], queryFn: fetchUsers })

  const deleteMut = useMutation({
    mutationFn: (id: string) => deleteUser(id),
    onSuccess: () => { toast.success('User deleted'); qc.invalidateQueries({ queryKey: ['admin-users'] }) },
    onError: err => toast.error(apiError(err, 'Failed to delete user')),
  })

  const roleMut = useMutation({
    mutationFn: ({ id, role }: { id: string; role: Role }) => updateRole(id, role),
    onSuccess: () => { toast.success('Role updated'); qc.invalidateQueries({ queryKey: ['admin-users'] }) },
    onError: err => toast.error(apiError(err, 'Failed to update role')),
  })

  function confirmDelete(u: AdminUser) {
    if (!confirm(`Delete "${u.username}"? This cannot be undone.`)) return
    deleteMut.mutate(u.id)
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-fg">Users</h1>
          <p className="text-sm text-muted mt-0.5">
            {data ? `${data.users.length} account${data.users.length !== 1 ? 's' : ''}` : 'Loading…'}
          </p>
        </div>
        <Button onClick={() => setShowCreate(true)}>
          <UserPlus />Add User
        </Button>
      </div>

      <div className="rounded-xl border border-line overflow-hidden">
        <Table>
          <TableHeader>
            <TableRow className="hover:bg-transparent">
              <TableHead>Username</TableHead>
              <TableHead>Role</TableHead>
              <TableHead className="hidden sm:table-cell">Created</TableHead>
              <TableHead className="text-right">Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {isLoading && (
              <TableRow>
                <TableCell colSpan={4} className="text-center py-10">
                  <Loader2 className="w-5 h-5 animate-spin text-muted mx-auto" />
                </TableCell>
              </TableRow>
            )}
            {data?.users.map(u => (
              <TableRow key={u.id}>
                <TableCell>
                  <div className="flex items-center gap-2">
                    <div className="w-7 h-7 rounded-full bg-surface-2 flex items-center justify-center shrink-0">
                      <span className="text-xs font-semibold text-muted">{u.username[0].toUpperCase()}</span>
                    </div>
                    <span className="font-medium text-fg">{u.username}</span>
                    {u.id === me?.id && <Badge variant="default" className="text-xs">You</Badge>}
                  </div>
                </TableCell>
                <TableCell>
                  <Select
                    value={u.role}
                    disabled={u.id === me?.id || roleMut.isPending}
                    onValueChange={v => roleMut.mutate({ id: u.id, role: v as Role })}
                  >
                    <SelectTrigger className="w-28 h-7 text-xs"><SelectValue /></SelectTrigger>
                    <SelectContent>
                      <SelectItem value="viewer">Viewer</SelectItem>
                      <SelectItem value="admin">Admin</SelectItem>
                    </SelectContent>
                  </Select>
                </TableCell>
                <TableCell className="hidden sm:table-cell">
                  <span className="text-xs text-muted">{formatDateTime(u.created_at)}</span>
                </TableCell>
                <TableCell>
                  <div className="flex items-center justify-end gap-1.5">
                    <Button variant="ghost" size="sm" onClick={() => setPwUser(u)} title="Change password">
                      <KeyRound className="w-3.5 h-3.5" />
                    </Button>
                    <Button
                      variant="ghost"
                      size="sm"
                      disabled={u.id === me?.id || deleteMut.isPending}
                      onClick={() => confirmDelete(u)}
                      className="text-red-500 hover:text-red-500 hover:bg-red-50 dark:hover:bg-red-950/30"
                      title="Delete user"
                    >
                      <Trash2 className="w-3.5 h-3.5" />
                    </Button>
                  </div>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </div>

      <div className="flex items-center gap-2 text-xs text-muted">
        <ShieldCheck className="w-3.5 h-3.5" />
        Admins can manage users. Viewers can only view the dashboard.
      </div>

      <CreateUserDialog open={showCreate} onClose={() => setShowCreate(false)} />
      <ChangePasswordDialog user={pwUser} onClose={() => setPwUser(null)} />
    </div>
  )
}
