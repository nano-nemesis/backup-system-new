import type { ReactNode } from 'react'
import Sidebar from './Sidebar'

interface Props {
  children: ReactNode
}

export default function AppShell({ children }: Props) {
  return (
    <div className="flex min-h-screen bg-zinc-950">
      <Sidebar />
      <main className="flex-1 min-w-0 overflow-y-auto">
        <div className="max-w-7xl mx-auto px-6 py-7">{children}</div>
      </main>
    </div>
  )
}
