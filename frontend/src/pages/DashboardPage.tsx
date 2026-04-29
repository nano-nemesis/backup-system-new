import { RefreshCw, ServerCrash, CheckCircle2, HelpCircle, Server } from 'lucide-react'
import { toast } from 'sonner'
import { useEffect, useRef } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import StatCard from '@/components/StatCard'
import NodeTable from '@/components/NodeTable'
import StatusPieChart from '@/components/charts/StatusPieChart'
import NodeBarChart from '@/components/charts/NodeBarChart'
import { useNodes } from '@/hooks/useNodes'
import { useQueryClient } from '@tanstack/react-query'

export default function DashboardPage() {
  const { data, isLoading, isError, dataUpdatedAt } = useNodes()
  const queryClient = useQueryClient()
  const prevErrorNodes = useRef<Set<string>>(new Set())

  // Toast when nodes flip to ERROR
  useEffect(() => {
    if (!data) return
    const currentErrors = new Set(
      data.nodes.filter((n) => n.last_status === 'ERROR').map((n) => n.name),
    )
    currentErrors.forEach((name) => {
      if (!prevErrorNodes.current.has(name)) {
        toast.error(`Backup failed: ${name}`, { duration: 6000 })
      }
    })
    prevErrorNodes.current = currentErrors
  }, [data])

  function handleRefresh() {
    queryClient.invalidateQueries({ queryKey: ['nodes'] })
  }

  const stats = data?.stats ?? { total: 0, ok: 0, failed: 0, unknown: 0 }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-zinc-100">Dashboard</h1>
          <p className="text-sm text-zinc-500 mt-0.5">
            {isLoading
              ? 'Loading…'
              : `${stats.total} nodes monitored · auto-refresh every 30s`}
          </p>
        </div>
        <div className="flex items-center gap-3">
          {dataUpdatedAt > 0 && (
            <span className="text-xs text-zinc-600">
              Updated {new Date(dataUpdatedAt).toLocaleTimeString()}
            </span>
          )}
          <Button variant="outline" size="sm" onClick={handleRefresh} disabled={isLoading}>
            <RefreshCw className={`w-3.5 h-3.5 ${isLoading ? 'animate-spin' : ''}`} />
            Refresh
          </Button>
        </div>
      </div>

      {/* Error banner */}
      {isError && (
        <div className="rounded-xl border border-red-900 bg-red-950/40 px-4 py-3 text-sm text-red-400">
          Failed to load node data. The backend may be unreachable.
        </div>
      )}

      {/* Stat cards */}
      <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
        <StatCard
          label="Total Nodes"
          value={stats.total}
          colorClass="text-zinc-100"
          icon={<Server className="w-4 h-4" />}
        />
        <StatCard
          label="Healthy"
          value={stats.ok}
          colorClass="text-green-400"
          icon={<CheckCircle2 className="w-4 h-4" />}
        />
        <StatCard
          label="Failed"
          value={stats.failed}
          colorClass="text-red-400"
          icon={<ServerCrash className="w-4 h-4" />}
        />
        <StatCard
          label="Unknown"
          value={stats.unknown}
          colorClass="text-zinc-400"
          icon={<HelpCircle className="w-4 h-4" />}
        />
      </div>

      {/* Charts */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
        <Card>
          <CardHeader>
            <CardTitle className="text-sm">Health Breakdown</CardTitle>
          </CardHeader>
          <CardContent className="pt-0">
            {data ? (
              <StatusPieChart ok={stats.ok} failed={stats.failed} unknown={stats.unknown} />
            ) : (
              <div className="h-[200px] animate-pulse bg-zinc-800 rounded-lg" />
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="text-sm">Backup Files per Node</CardTitle>
          </CardHeader>
          <CardContent className="pt-0">
            {data ? (
              <NodeBarChart nodes={data.nodes} />
            ) : (
              <div className="h-[220px] animate-pulse bg-zinc-800 rounded-lg" />
            )}
          </CardContent>
        </Card>
      </div>

      {/* Node table */}
      <div>
        <h2 className="text-sm font-semibold text-zinc-400 uppercase tracking-wider mb-3">
          All Nodes
        </h2>
        {isLoading ? (
          <div className="space-y-2">
            {Array.from({ length: 5 }).map((_, i) => (
              <div key={i} className="h-12 animate-pulse bg-zinc-800/60 rounded-lg" />
            ))}
          </div>
        ) : data ? (
          <NodeTable nodes={data.nodes} />
        ) : null}
      </div>
    </div>
  )
}
