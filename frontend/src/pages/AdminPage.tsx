import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { UserPlus, Trash2, KeyRound, ShieldCheck, Loader2 } from 'lucide-react'
import { toast } from 'sonner'
import axios from 'axios'
import api from '@/lib/axios'
import { useAuth } from '@/hooks/useAuth'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogClose,
} from '@/components/ui/dialog'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  Table,
  TableHeader,
  TableBody,
  TableRow,
  TableHead,
  TableCell,
} from '@/components/ui/table'
import { formatDateTime } from '@/lib/utils'
import type { AdminUser, AdminUsersResponse, Role } from '@/types'

// ─── API calls ────────────────────────────────────────────────────────────────

const fetchUsers = () => api.get<AdminUsersResponse>('/admin/users').then((r) => r.data)
const createUser = (body: { username: string; password: string; role: Role }) =>
  api.post('/admin/users', body)
const deleteUser = (id: string) => api.delete(`/admin/users/${id}`)
const updateRole = (id: string, role: Role) => api.patch(`/admin/users/${id}/role`, { role })
const updatePassword = (id: string, password: string) =>
  api.patch(`/admin/users/${id}/password`, { password })

// ─── Sub-modals ───────────────────────────────────────────────────────────────

function CreateUserDialog({ open, onClose }: { open: boolean; onClose: () => void }) {
  const qc = useQueryClient()
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [role, setRole] = useState<Role>('viewer')

  const mutation = useMutation({
    mutationFn: () => createUser({ username, password, role }),
    onSuccess: () => {
      toast.success('User created')
      qc.invalidateQueries({ queryKey: ['admin-users'] })
      onClose()
      setUsername('')
      setPassword('')
      setRole('viewer')
    },
    onError: (err) => {
      if (axios.isAxiosError(err)) toast.error(err.response?.data?.error ?? 'Failed to create user')
    },
  })

  return (
    <Dialog open={open} onOpenChange={(v) => !v && onClose()}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Create User</DialogTitle>
        </DialogHeader>
        <div className="space-y-4">
          <div className="space-y-1.5">
            <Label>Username</Label>
            <Input value={username} onChange={(e) => setUsername(e.target.value)} placeholder="username" />
          </div>
          <div className="space-y-1.5">
            <Label>Password</Label>
            <Input type="password" value={password} onChange={(e) => setPassword(e.target.value)} placeholder="Min. 8 characters" />
          </div>
          <div className="space-y-1.5">
            <Label>Role</Label>
            <Select value={role} onValueChange={(v) => setRole(v as Role)}>
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="viewer">Viewer</SelectItem>
                <SelectItem value="admin">Admin</SelectItem>
              </SelectContent>
            </Select>
          </div>
          <div className="flex gap-2 pt-1">
            <Button
              className="flex-1"
              onClick={() => mutation.mutate()}
              disabled={!username || !password || mutation.isPending}
            >
              {mutation.isPending ? <Loader2 className="animate-spin" /> : <UserPlus />}
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

function ChangePasswordDialog({
  user,
  onClose,
}: {
  user: AdminUser | null
  onClose: () => void
}) {
  const qc = useQueryClient()
  const [password, setPassword] = useState('')

  const mutation = useMutation({
    mutationFn: () => updatePassword(user!.id, password),
    onSuccess: () => {
      toast.success('Password updated')
      qc.invalidateQueries({ queryKey: ['admin-users'] })
      onClose()
      setPassword('')
    },
    onError: (err) => {
      if (axios.isAxiosError(err)) toast.error(err.response?.data?.error ?? 'Failed to update password')
    },
  })

  return (
    <Dialog open={!!user} onOpenChange={(v) => !v && onClose()}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Change password — {user?.username}</DialogTitle>
        </DialogHeader>
        <div className="space-y-4">
          <div className="space-y-1.5">
            <Label>New Password</Label>
            <Input
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              placeholder="Min. 8 characters"
            />
          </div>
          <div className="flex gap-2 pt-1">
            <Button
              className="flex-1"
              onClick={() => mutation.mutate()}
              disabled={password.length < 8 || mutation.isPending}
            >
              {mutation.isPending ? <Loader2 className="animate-spin" /> : <KeyRound />}
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
  const { user: currentUser } = useAuth()
  const qc = useQueryClient()
  const [showCreate, setShowCreate] = useState(false)
  const [pwUser, setPwUser] = useState<AdminUser | null>(null)

  const { data, isLoading } = useQuery({
    queryKey: ['admin-users'],
    queryFn: fetchUsers,
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => deleteUser(id),
    onSuccess: () => {
      toast.success('User deleted')
      qc.invalidateQueries({ queryKey: ['admin-users'] })
    },
    onError: (err) => {
      if (axios.isAxiosError(err)) toast.error(err.response?.data?.error ?? 'Failed to delete user')
    },
  })

  const roleMutation = useMutation({
    mutationFn: ({ id, role }: { id: string; role: Role }) => updateRole(id, role),
    onSuccess: () => {
      toast.success('Role updated')
      qc.invalidateQueries({ queryKey: ['admin-users'] })
    },
    onError: (err) => {
      if (axios.isAxiosError(err)) toast.error(err.response?.data?.error ?? 'Failed to update role')
    },
  })

  function confirmDelete(u: AdminUser) {
    if (!confirm(`Delete user "${u.username}"? This cannot be undone.`)) return
    deleteMutation.mutate(u.id)
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-zinc-100">Users</h1>
          <p className="text-sm text-zinc-500 mt-0.5">
            {data ? `${data.users.length} account${data.users.length !== 1 ? 's' : ''}` : 'Loading…'}
          </p>
        </div>
        <Button onClick={() => setShowCreate(true)}>
          <UserPlus />
          Add User
        </Button>
      </div>

      {/* Table */}
      <div className="rounded-xl border border-zinc-800 overflow-hidden">
        <Table>
          <TableHeader>
            <TableRow className="hover:bg-transparent">
              <TableHead>Username</TableHead>
              <TableHead>Role</TableHead>
              <TableHead>Created</TableHead>
              <TableHead className="text-right">Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {isLoading && (
              <TableRow>
                <TableCell colSpan={4} className="text-center py-10">
                  <Loader2 className="w-5 h-5 animate-spin text-zinc-600 mx-auto" />
                </TableCell>
              </TableRow>
            )}
            {data?.users.map((u) => (
              <TableRow key={u.id}>
                <TableCell>
                  <div className="flex items-center gap-2">
                    <div className="w-7 h-7 rounded-full bg-zinc-800 flex items-center justify-center">
                      <span className="text-xs font-semibold text-zinc-400">
                        {u.username[0].toUpperCase()}
                      </span>
                    </div>
                    <span className="font-medium text-zinc-200">{u.username}</span>
                    {u.id === currentUser?.id && (
                      <Badge variant="default" className="text-xs">You</Badge>
                    )}
                  </div>
                </TableCell>
                <TableCell>
                  <Select
                    value={u.role}
                    disabled={u.id === currentUser?.id || roleMutation.isPending}
                    onValueChange={(v) => roleMutation.mutate({ id: u.id, role: v as Role })}
                  >
                    <SelectTrigger className="w-28 h-7 text-xs">
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="viewer">Viewer</SelectItem>
                      <SelectItem value="admin">Admin</SelectItem>
                    </SelectContent>
                  </Select>
                </TableCell>
                <TableCell>
                  <span className="text-xs text-zinc-500">{formatDateTime(u.created_at)}</span>
                </TableCell>
                <TableCell>
                  <div className="flex items-center justify-end gap-1.5">
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => setPwUser(u)}
                      title="Change password"
                    >
                      <KeyRound className="w-3.5 h-3.5" />
                    </Button>
                    <Button
                      variant="ghost"
                      size="sm"
                      disabled={u.id === currentUser?.id || deleteMutation.isPending}
                      onClick={() => confirmDelete(u)}
                      className="text-red-500 hover:text-red-400 hover:bg-red-950/30"
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

      {/* Admin badge legend */}
      <div className="flex items-center gap-2 text-xs text-zinc-600">
        <ShieldCheck className="w-3.5 h-3.5" />
        Admins can manage users. Viewers can only view the dashboard.
      </div>

      {/* Dialogs */}
      <CreateUserDialog open={showCreate} onClose={() => setShowCreate(false)} />
      <ChangePasswordDialog user={pwUser} onClose={() => setPwUser(null)} />
    </div>
  )
}
