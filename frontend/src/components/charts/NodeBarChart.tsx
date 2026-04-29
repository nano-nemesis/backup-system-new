import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  Tooltip,
  ResponsiveContainer,
  Cell,
} from 'recharts'
import type { NodeSummary } from '@/types'

interface Props {
  nodes: NodeSummary[]
}

export default function NodeBarChart({ nodes }: Props) {
  const data = nodes
    .filter((n) => n.backup_files.length > 0)
    .map((n) => ({
      name: n.name.length > 18 ? n.name.slice(0, 16) + '…' : n.name,
      files: n.backup_files.length,
      status: n.last_status,
    }))
    .sort((a, b) => b.files - a.files)
    .slice(0, 15)

  if (data.length === 0) {
    return (
      <div className="flex items-center justify-center h-48 text-zinc-600 text-sm">
        No backup files yet
      </div>
    )
  }

  const colorOf = (status: string) => {
    if (status === 'SUCCESS') return '#6366f1'
    if (status === 'ERROR') return '#ef4444'
    return '#52525b'
  }

  return (
    <ResponsiveContainer width="100%" height={220}>
      <BarChart data={data} layout="vertical" margin={{ left: 8, right: 16, top: 4, bottom: 4 }}>
        <XAxis
          type="number"
          allowDecimals={false}
          tick={{ fill: '#52525b', fontSize: 11 }}
          tickLine={false}
          axisLine={false}
        />
        <YAxis
          type="category"
          dataKey="name"
          width={120}
          tick={{ fill: '#a1a1aa', fontSize: 11 }}
          tickLine={false}
          axisLine={false}
        />
        <Tooltip
          cursor={{ fill: '#27272a' }}
          contentStyle={{
            background: '#18181b',
            border: '1px solid #27272a',
            borderRadius: '8px',
            color: '#fafafa',
            fontSize: 12,
          }}
          formatter={(value: number) => [value, 'files']}
        />
        <Bar dataKey="files" radius={[0, 4, 4, 0]} maxBarSize={20}>
          {data.map((entry, i) => (
            <Cell key={i} fill={colorOf(entry.status)} />
          ))}
        </Bar>
      </BarChart>
    </ResponsiveContainer>
  )
}
