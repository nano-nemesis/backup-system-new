import { useQuery } from '@tanstack/react-query'
import api from '@/lib/axios'
import type { NodesResponse } from '@/types'

export function useNodes() {
  return useQuery({
    queryKey: ['nodes'],
    queryFn: () => api.get<NodesResponse>('/nodes').then((r) => r.data),
    refetchInterval: 30_000,
    refetchIntervalInBackground: false,
  })
}
