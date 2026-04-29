import { PieChart, Pie, Cell, Tooltip, ResponsiveContainer, Legend } from 'recharts'

interface Props {
  ok: number
  failed: number
  unknown: number
}

const COLORS = {
  OK: '#22c55e',
  Failed: '#ef4444',
  Unknown: '#52525b',
}

export default function StatusPieChart({ ok, failed, unknown }: Props) {
  const data = [
    { name: 'OK', value: ok },
    { name: 'Failed', value: failed },
    { name: 'Unknown', value: unknown },
  ].filter((d) => d.value > 0)

  if (data.length === 0) {
    return (
      <div className="flex items-center justify-center h-48 text-zinc-600 text-sm">
        No data
      </div>
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
          {data.map((entry) => (
            <Cell key={entry.name} fill={COLORS[entry.name as keyof typeof COLORS]} />
          ))}
        </Pie>
        <Tooltip
          contentStyle={{
            background: '#18181b',
            border: '1px solid #27272a',
            borderRadius: '8px',
            color: '#fafafa',
            fontSize: 12,
          }}
          formatter={(value: number, name: string) => [value, name]}
        />
        <Legend
          iconType="circle"
          iconSize={8}
          formatter={(value) => (
            <span style={{ color: '#a1a1aa', fontSize: 12 }}>{value}</span>
          )}
        />
      </PieChart>
    </ResponsiveContainer>
  )
}
