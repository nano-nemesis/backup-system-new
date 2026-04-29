import type { ReactNode } from 'react'
import { Card, CardContent } from '@/components/ui/card'
import { cn } from '@/lib/utils'

interface Props {
  label: string
  value: number
  colorClass: string
  icon: ReactNode
}

export default function StatCard({ label, value, colorClass, icon }: Props) {
  return (
    <Card className="relative overflow-hidden">
      <CardContent className="p-5">
        <div className="flex items-center justify-between">
          <div>
            <p className="text-xs font-medium text-zinc-500 uppercase tracking-wider">{label}</p>
            <p className={cn('text-3xl font-bold mt-1 tabular-nums', colorClass)}>{value}</p>
          </div>
          <div className={cn('p-2.5 rounded-lg bg-zinc-800', colorClass)}>{icon}</div>
        </div>
      </CardContent>
    </Card>
  )
}
