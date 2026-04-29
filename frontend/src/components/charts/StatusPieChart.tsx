import { useMemo } from 'react'
import { PieChart, Pie, Cell, Tooltip, ResponsiveContainer, Legend } from 'recharts'
import { useTheme } from '@/context/ThemeContext'

interface Props {
  ok: number
  failed: number
  unknown: number
}

const SLICES = [
  { name: 'OK',      color: '#22c55e' },
  { name: 'Failed',  color: '#ef4444' },
  { name: 'Unknown', color: '#94a3b8' },
]

export default function StatusPieChart({ ok, failed, unknown }: Props) {
  const { theme } = useTheme()

  const data = useMemo(
    () => [
      { name: 'OK',      value: ok },
      { name: 'Failed',  value: failed },
      { name: 'Unknown', value: unknown },
    ].filter(d => d.value > 0),
    [ok, failed, unknown],
  )

  const tooltipStyle = {
    background: theme === 'dark' ? '#18181b' : '#ffffff',
    border: `1px solid ${theme === 'dark' ? '#27272a' : '#e5e7eb'}`,
    borderRadius: '8px',
    color: theme === 'dark' ? '#fafafa' : '#111827',
    fontSize: 12,
  }

  if (data.length === 0) {
    return (
      <div className="flex items-center justify-center h-48 text-muted text-sm">No data</div>
    )
  }

  return (
    <ResponsiveContainer width="100%" height={200}>
      <PieChart>
        <Pie
          data={data}
          cx="50%"
          cy="50%"
          innerRadius={55}
          outerRadius={80}
          paddingAngle={3}
          dataKey="value"
          strokeWidth={0}
        >
          {data.map(entry => {
            const slice = SLICES.find(s => s.name === entry.name)
            return <Cell key={entry.name} fill={slice?.color ?? '#94a3b8'} />
          })}
        </Pie>
        <Tooltip contentStyle={tooltipStyle} formatter={(v: number, n: string) => [v, n]} />
        <Legend
          iconType="circle"
          iconSize={8}
          formatter={value => (
            <span style={{ color: theme === 'dark' ? '#a1a1aa' : '#6b7280', fontSize: 12 }}>
              {value}
            </span>
          )}
        />
      </PieChart>
    </ResponsiveContainer>
  )
}
