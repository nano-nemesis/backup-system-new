import { NavLink } from 'react-router-dom'
import { LayoutDashboard, ShieldCheck, LogOut, Server, X, Sun, Moon } from 'lucide-react'
import { cn } from '@/lib/utils'
import { useAuth } from '@/hooks/useAuth'
import { useTheme } from '@/context/ThemeContext'
import { toast } from 'sonner'

interface Props {
  isOpen: boolean
  onClose: () => void
}

const navItems = [
  { to: '/', label: 'Dashboard', icon: LayoutDashboard, exact: true },
  { to: '/admin', label: 'Users', icon: ShieldCheck, exact: false, adminOnly: true },
]

export default function Sidebar({ isOpen, onClose }: Props) {
  const { user, logout } = useAuth()
  const { theme, toggleTheme } = useTheme()

  async function handleLogout() {
    try {
      await logout()
    } catch {
      toast.error('Failed to log out')
    }
  }

  return (
    <aside
      className={cn(
        'fixed inset-y-0 left-0 z-30 flex flex-col w-56 shrink-0',
        'border-r border-line bg-surface',
        'transition-transform duration-200 ease-in-out',
        'md:sticky md:top-0 md:h-screen md:translate-x-0',
        isOpen ? 'translate-x-0' : '-translate-x-full',
      )}
    >
      {/* Logo */}
      <div className="flex items-center justify-between px-4 h-14 border-b border-line shrink-0">
        <div className="flex items-center gap-2.5">
          <div className="w-7 h-7 rounded-lg bg-indigo-600 flex items-center justify-center">
            <Server className="w-4 h-4 text-white" />
          </div>
          <span className="font-semibold text-fg text-sm tracking-tight">BackupSystem</span>
        </div>
        <button
          onClick={onClose}
          className="p-1 rounded-md text-muted hover:text-fg transition-colors md:hidden"
          aria-label="Close menu"
        >
          <X className="w-4 h-4" />
        </button>
      </div>

      {/* Nav */}
      <nav className="flex-1 px-2 py-3 space-y-0.5 overflow-y-auto">
        {navItems
          .filter(item => !item.adminOnly || user?.role === 'admin')
          .map(item => (
            <NavLink
              key={item.to}
              to={item.to}
              end={item.exact}
              onClick={onClose}
              className={({ isActive }) =>
                cn(
                  'flex items-center gap-2.5 px-3 py-2 rounded-md text-sm transition-colors',
                  isActive
                    ? 'bg-indigo-600/10 text-indigo-600 dark:text-indigo-400 font-medium'
                    : 'text-muted hover:bg-surface-2 hover:text-fg',
                )
              }
            >
              <item.icon className="w-4 h-4 shrink-0" />
              {item.label}
            </NavLink>
          ))}
      </nav>

      {/* Footer */}
      <div className="border-t border-line p-3 shrink-0 space-y-1">
        <button
          onClick={toggleTheme}
          className="flex items-center gap-2 w-full px-3 py-2 rounded-md text-sm text-muted hover:bg-surface-2 hover:text-fg transition-colors"
        >
          {theme === 'dark'
            ? <Sun className="w-4 h-4 shrink-0" />
            : <Moon className="w-4 h-4 shrink-0" />}
          {theme === 'dark' ? 'Light mode' : 'Dark mode'}
        </button>

        <div className="flex items-center gap-2.5 px-2 py-2">
          <div className="w-7 h-7 rounded-full bg-indigo-100 dark:bg-indigo-900 flex items-center justify-center shrink-0">
            <span className="text-xs font-semibold text-indigo-600 dark:text-indigo-300">
              {user?.username?.[0]?.toUpperCase() ?? '?'}
            </span>
          </div>
          <div className="min-w-0">
            <p className="text-sm font-medium text-fg truncate">{user?.username}</p>
            <p className="text-xs text-muted capitalize">{user?.role}</p>
          </div>
        </div>

        <button
          onClick={handleLogout}
          className="flex items-center gap-2 w-full px-3 py-2 rounded-md text-sm text-muted hover:bg-surface-2 hover:text-fg transition-colors"
        >
          <LogOut className="w-4 h-4 shrink-0" />
          Sign out
        </button>
      </div>
    </aside>
  )
}
