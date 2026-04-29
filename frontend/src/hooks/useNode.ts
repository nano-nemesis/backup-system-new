import { useQuery } from '@tanstack/react-query'
import api from '@/lib/axios'
import type { NodeDetailResponse } from '@/types'

export function useNode(name: string) {
  return useQuery({
    queryKey: ['node', name],
    queryFn: () =>
      api.get<NodeDetailResponse>(`/nodes/${encodeURIComponent(name)}`).then((r) => r.data),
    refetchInterval: 15_000,
    enabled: !!name,
  })
}
