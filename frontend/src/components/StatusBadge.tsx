import { Badge } from '@/components/ui/badge'
import type { NodeStatus } from '@/types'

interface Props {
  status: NodeStatus
}

const DOT = 'inline-block w-1.5 h-1.5 rounded-full'

export default function StatusBadge({ status }: Props) {
  if (status === 'SUCCESS') {
    return (
      <Badge variant="success">
        <span className={`${DOT} bg-green-400 animate-pulse-dot`} />
        OK
      </Badge>
    )
  }
  if (status === 'ERROR') {
    return (
      <Badge variant="error">
        <span className={`${DOT} bg-red-400`} />
        Error
      </Badge>
    )
  }
  return (
    <Badge variant="unknown">
      <span className={`${DOT} bg-zinc-500`} />
      Unknown
    </Badge>
  )
}
