import { NavLink } from 'react-router-dom'
import {
  LayoutDashboard,
  ShieldCheck,
  LogOut,
  Server,
} from 'lucide-react'
import { cn } from '@/lib/utils'
import { useAuth } from '@/hooks/useAuth'
import { toast } from 'sonner'

const navItems = [
  { to: '/', label: 'Dashboard', icon: LayoutDashboard, exact: true },
  { to: '/admin', label: 'Users', icon: ShieldCheck, exact: false, adminOnly: true },
]

export default function Sidebar() {
  const { user, logout } = useAuth()

  async function handleLogout() {
    try {
      await logout()
    } catch {
      toast.error('Failed to log out')
    }
  }

  return (
    <aside className="flex flex-col w-56 shrink-0 border-r border-zinc-800 bg-zinc-950 h-screen sticky top-0">
      {/* Logo */}
      <div className="flex items-center gap-2.5 px-4 h-14 border-b border-zinc-800">
        <div className="w-7 h-7 rounded-lg bg-indigo-600 flex items-center justify-center">
          <Server className="w-4 h-4 text-white" />
        </div>
        <span className="font-semibold text-zinc-100 text-sm tracking-tight">BackupSystem</span>
      </div>

      {/* Nav */}
      <nav className="flex-1 px-2 py-3 space-y-0.5 overflow-y-auto">
        {navItems
          .filter((item) => !item.adminOnly || user?.role === 'admin')
          .map((item) => (
            <NavLink
              key={item.to}
              to={item.to}
              end={item.exact}
              className={({ isActive }) =>
                cn(
                  'flex items-center gap-2.5 px-3 py-2 rounded-md text-sm transition-colors',
                  isActive
                    ? 'bg-indigo-600/10 text-indigo-400 font-medium'
                    : 'text-zinc-400 hover:bg-zinc-800 hover:text-zinc-100',
                )
              }
            >
              <item.icon className="w-4 h-4 shrink-0" />
              {item.label}
            </NavLink>
          ))}
      </nav>

      {/* User section */}
      <div className="border-t border-zinc-800 p-3">
        <div className="flex items-center gap-2.5 mb-2 px-2">
          <div className="w-7 h-7 rounded-full bg-indigo-900 flex items-center justify-center shrink-0">
            <span className="text-xs font-semibold text-indigo-300">
              {user?.username?.[0]?.toUpperCase() ?? '?'}
            </span>
          </div>
          <div className="min-w-0">
            <p className="text-sm font-medium text-zinc-200 truncate">{user?.username}</p>
            <p className="text-xs text-zinc-500 capitalize">{user?.role}</p>
          </div>
        </div>
        <button
          onClick={handleLogout}
          className="flex items-center gap-2 w-full px-3 py-2 rounded-md text-sm text-zinc-400 hover:bg-zinc-800 hover:text-zinc-100 transition-colors"
        >
          <LogOut className="w-4 h-4 shrink-0" />
          Sign out
        </button>
      </div>
    </aside>
  )
}
