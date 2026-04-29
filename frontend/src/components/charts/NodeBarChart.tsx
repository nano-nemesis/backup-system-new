import { useMemo } from 'react'
import { BarChart, Bar, XAxis, YAxis, Tooltip, ResponsiveContainer, Cell } from 'recharts'
import { useTheme } from '@/context/ThemeContext'
import type { NodeSummary } from '@/types'

const MAX_NODES = 15
const MAX_LABEL = 18

function statusColor(status: string): string {
  if (status === 'SUCCESS') return '#6366f1'
  if (status === 'ERROR')   return '#ef4444'
  return '#94a3b8'
}

export default function NodeBarChart({ nodes }: { nodes: NodeSummary[] }) {
  const { theme } = useTheme()

  const data = useMemo(
    () =>
      nodes
        .filter(n => n.backup_files.length > 0)
        .map(n => ({
          name:   n.name.length > MAX_LABEL ? n.name.slice(0, MAX_LABEL - 1) + '…' : n.name,
          files:  n.backup_files.length,
          status: n.last_status,
        }))
        .sort((a, b) => b.files - a.files)
        .slice(0, MAX_NODES),
    [nodes],
  )

  const tickColor  = theme === 'dark' ? '#a1a1aa' : '#6b7280'
  const tooltipStyle = {
    background: theme === 'dark' ? '#18181b' : '#ffffff',
    border: `1px solid ${theme === 'dark' ? '#27272a' : '#e5e7eb'}`,
    borderRadius: '8px',
    color: theme === 'dark' ? '#fafafa' : '#111827',
    fontSize: 12,
  }

  if (data.length === 0) {
    return (
      <div className="flex items-center justify-center h-48 text-muted text-sm">
        No backup files yet
      </div>
    )
  }

  return (
    <ResponsiveContainer width="100%" height={220}>
      <BarChart data={data} layout="vertical" margin={{ left: 8, right: 16, top: 4, bottom: 4 }}>
        <XAxis
          type="number"
          allowDecimals={false}
          tick={{ fill: tickColor, fontSize: 11 }}
          tickLine={false}
          axisLine={false}
        />
        <YAxis
          type="category"
          dataKey="name"
          width={120}
          tick={{ fill: tickColor, fontSize: 11 }}
          tickLine={false}
          axisLine={false}
        />
        <Tooltip
          cursor={{ fill: theme === 'dark' ? '#27272a' : '#f3f4f6' }}
          contentStyle={tooltipStyle}
          formatter={(v: number) => [v, 'files']}
        />
        <Bar dataKey="files" radius={[0, 4, 4, 0]} maxBarSize={20}>
          {data.map((entry, i) => (
            <Cell key={i} fill={statusColor(entry.status)} />
          ))}
        </Bar>
      </BarChart>
    </ResponsiveContainer>
  )
}
