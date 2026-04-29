import { useParams, Link } from 'react-router-dom'
import { ArrowLeft, RefreshCw, MonitorSmartphone, Database } from 'lucide-react'
import { useQueryClient } from '@tanstack/react-query'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import StatusBadge from '@/components/StatusBadge'
import BackupFileList from '@/components/BackupFileList'
import LogViewer from '@/components/LogViewer'
import { formatRelativeTime, formatDateTime } from '@/lib/utils'
import { useNode } from '@/hooks/useNode'

export default function NodeDetailPage() {
  const { name = '' } = useParams<{ name: string }>()
  const { data, isLoading, isError } = useNode(name)
  const queryClient = useQueryClient()

  function handleRefresh() {
    queryClient.invalidateQueries({ queryKey: ['node', name] })
  }

  if (isLoading) {
    return (
      <div className="space-y-6">
        <div className="h-8 w-48 animate-pulse bg-zinc-800 rounded-lg" />
        <div className="h-24 animate-pulse bg-zinc-800 rounded-xl" />
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          <div className="h-64 animate-pulse bg-zinc-800 rounded-xl" />
          <div className="h-64 animate-pulse bg-zinc-800 rounded-xl" />
        </div>
      </div>
    )
  }

  if (isError || !data) {
    return (
      <div className="space-y-4">
        <Link to="/">
          <Button variant="ghost" size="sm">
            <ArrowLeft className="w-4 h-4" />
            Back
          </Button>
        </Link>
        <div className="rounded-xl border border-red-900 bg-red-950/40 px-4 py-3 text-sm text-red-400">
          Node not found or failed to load.
        </div>
      </div>
    )
  }

  const { node, logs } = data
  const TypeIcon = node.type === 'database' ? Database : MonitorSmartphone

  return (
    <div className="space-y-6">
      {/* Back + Refresh */}
      <div className="flex items-center justify-between">
        <Link to="/">
          <Button variant="ghost" size="sm">
            <ArrowLeft className="w-4 h-4" />
            Back
          </Button>
        </Link>
        <Button variant="outline" size="sm" onClick={handleRefresh}>
          <RefreshCw className="w-3.5 h-3.5" />
          Refresh
        </Button>
      </div>

      {/* Node header */}
      <div className="rounded-xl border border-zinc-800 bg-zinc-900 p-5">
        <div className="flex items-start gap-4">
          <div className="w-10 h-10 rounded-lg bg-zinc-800 flex items-center justify-center shrink-0">
            <TypeIcon className="w-5 h-5 text-zinc-400" />
          </div>
          <div className="flex-1 min-w-0">
            <div className="flex items-center gap-3 flex-wrap">
              <h1 className="text-xl font-bold text-zinc-100">{node.name}</h1>
              <StatusBadge status={node.last_status} />
              <Badge variant="default" className="text-xs">
                {node.type}
              </Badge>
            </div>
            <div className="flex items-center gap-4 mt-2 flex-wrap text-xs text-zinc-500">
              <span className="font-mono">{node.host}</span>
              <span>
                Last backup:{' '}
                {node.last_time ? (
                  <span title={formatDateTime(node.last_time)}>
                    {formatDateTime(node.last_time)}{' '}
                    <span className="text-zinc-600">({formatRelativeTime(node.last_time)})</span>
                  </span>
                ) : (
                  'never'
                )}
              </span>
            </div>
            {node.last_msg && (
              <p className="mt-2 text-xs text-zinc-500 font-mono truncate" title={node.last_msg}>
                {node.last_msg}
              </p>
            )}
          </div>
        </div>
      </div>

      {/* Files + Logs grid */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <div>
          <h2 className="text-xs font-semibold text-zinc-500 uppercase tracking-wider mb-3">
            Backup Files ({node.backup_files.length})
          </h2>
          <BackupFileList files={node.backup_files} />
        </div>

        <div>
          <h2 className="text-xs font-semibold text-zinc-500 uppercase tracking-wider mb-3">
            Log (last {logs.length} lines)
          </h2>
          <LogViewer lines={logs} />
        </div>
      </div>
    </div>
  )
}
