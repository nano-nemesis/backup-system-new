import { useState, useMemo } from 'react'
import { Link } from 'react-router-dom'
import { Search, ChevronRight } from 'lucide-react'
import { Table, TableHeader, TableBody, TableRow, TableHead, TableCell } from '@/components/ui/table'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import StatusBadge from '@/components/StatusBadge'
import { formatRelativeTime, formatDateTime } from '@/lib/utils'
import type { NodeSummary } from '@/types'

type Filter = 'all' | 'mikrotik' | 'database' | 'SUCCESS' | 'ERROR' | 'UNKNOWN'

const FILTERS: { key: Filter; label: string }[] = [
  { key: 'all',      label: 'All' },
  { key: 'mikrotik', label: 'Mikrotik' },
  { key: 'database', label: 'Database' },
  { key: 'SUCCESS',  label: 'OK' },
  { key: 'ERROR',    label: 'Error' },
  { key: 'UNKNOWN',  label: 'Unknown' },
]

export default function NodeTable({ nodes }: { nodes: NodeSummary[] }) {
  const [search, setSearch] = useState('')
  const [filter, setFilter] = useState<Filter>('all')

  const visible = useMemo(() =>
    nodes.filter(n => {
      const matchSearch = n.name.toLowerCase().includes(search.toLowerCase())
      const matchFilter = filter === 'all' || n.type === filter || n.last_status === filter
      return matchSearch && matchFilter
    }),
    [nodes, search, filter],
  )

  return (
    <div className="space-y-3">
      {/* Toolbar */}
      <div className="flex flex-col sm:flex-row sm:items-center gap-3">
        <div className="relative sm:w-64">
          <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 h-3.5 w-3.5 text-muted" />
          <Input
            placeholder="Search nodes…"
            value={search}
            onChange={e => setSearch(e.target.value)}
            className="pl-8 h-8 text-xs"
          />
        </div>
        <div className="flex items-center gap-1 flex-wrap">
          {FILTERS.map(tab => (
            <button
              key={tab.key}
              onClick={() => setFilter(tab.key)}
              className={`px-3 py-1.5 rounded-md text-xs font-medium transition-colors ${
                filter === tab.key
                  ? 'bg-indigo-600 text-white'
                  : 'text-muted hover:bg-surface-2 hover:text-fg'
              }`}
            >
              {tab.label}
            </button>
          ))}
        </div>
      </div>

      {/* Table */}
      <div className="rounded-xl border border-line overflow-hidden">
        <Table>
          <TableHeader>
            <TableRow className="hover:bg-transparent">
              <TableHead>Node</TableHead>
              <TableHead className="hidden sm:table-cell">Type</TableHead>
              <TableHead className="hidden md:table-cell">Host</TableHead>
              <TableHead className="hidden lg:table-cell">Last Backup</TableHead>
              <TableHead className="hidden sm:table-cell">Files</TableHead>
              <TableHead>Status</TableHead>
              <TableHead className="w-8" />
            </TableRow>
          </TableHeader>
          <TableBody>
            {visible.length === 0 && (
              <TableRow>
                <TableCell colSpan={7} className="text-center text-muted py-10">
                  No nodes match the current filter.
                </TableCell>
              </TableRow>
            )}
            {visible.map(node => (
              <TableRow key={node.name}>
                <TableCell>
                  <Link
                    to={`/nodes/${encodeURIComponent(node.name)}`}
                    className="font-medium text-fg hover:text-indigo-500 transition-colors"
                  >
                    {node.name}
                  </Link>
                </TableCell>
                <TableCell className="hidden sm:table-cell">
                  <Badge variant="default" className="text-xs">{node.type}</Badge>
                </TableCell>
                <TableCell className="hidden md:table-cell">
                  <span className="font-mono text-xs text-muted">{node.host}</span>
                </TableCell>
                <TableCell className="hidden lg:table-cell">
                  <span className="text-xs text-muted">
                    {node.last_time
                      ? <span title={formatDateTime(node.last_time)}>{formatRelativeTime(node.last_time)}</span>
                      : '—'}
                  </span>
                </TableCell>
                <TableCell className="hidden sm:table-cell">
                  <span className="text-xs text-muted">{node.backup_files.length}</span>
                </TableCell>
                <TableCell>
                  <StatusBadge status={node.last_status} />
                </TableCell>
                <TableCell>
                  <Link
                    to={`/nodes/${encodeURIComponent(node.name)}`}
                    className="text-muted hover:text-fg transition-colors"
                  >
                    <ChevronRight className="w-4 h-4" />
                  </Link>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </div>
      <p className="text-xs text-muted px-1">
        Showing {visible.length} of {nodes.length} nodes
      </p>
    </div>
  )
}
