import { useState, useMemo } from 'react'
import { Link } from 'react-router-dom'
import { Search, ChevronRight } from 'lucide-react'
import {
  Table,
  TableHeader,
  TableBody,
  TableRow,
  TableHead,
  TableCell,
} from '@/components/ui/table'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import StatusBadge from '@/components/StatusBadge'
import { formatRelativeTime, formatDateTime } from '@/lib/utils'
import type { NodeSummary } from '@/types'

interface Props {
  nodes: NodeSummary[]
}

type Filter = 'all' | 'mikrotik' | 'database' | 'SUCCESS' | 'ERROR' | 'UNKNOWN'

const FILTER_TABS: { key: Filter; label: string }[] = [
  { key: 'all', label: 'All' },
  { key: 'mikrotik', label: 'Mikrotik' },
  { key: 'database', label: 'Database' },
  { key: 'SUCCESS', label: 'OK' },
  { key: 'ERROR', label: 'Error' },
  { key: 'UNKNOWN', label: 'Unknown' },
]

export default function NodeTable({ nodes }: Props) {
  const [search, setSearch] = useState('')
  const [filter, setFilter] = useState<Filter>('all')

  const visible = useMemo(() => {
    return nodes.filter((n) => {
      const matchSearch = n.name.toLowerCase().includes(search.toLowerCase())
      const matchFilter =
        filter === 'all' ||
        n.type === filter ||
        n.last_status === filter
      return matchSearch && matchFilter
    })
  }, [nodes, search, filter])

  return (
    <div className="space-y-3">
      {/* Toolbar */}
      <div className="flex items-center gap-3 flex-wrap">
        <div className="relative flex-1 min-w-48 max-w-xs">
          <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 h-3.5 w-3.5 text-zinc-500" />
          <Input
            placeholder="Search nodes…"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="pl-8 h-8 text-xs"
          />
        </div>
        <div className="flex items-center gap-1">
          {FILTER_TABS.map((tab) => (
            <button
              key={tab.key}
              onClick={() => setFilter(tab.key)}
              className={`px-3 py-1.5 rounded-md text-xs font-medium transition-colors ${
                filter === tab.key
                  ? 'bg-indigo-600 text-white'
                  : 'text-zinc-400 hover:bg-zinc-800 hover:text-zinc-100'
              }`}
            >
              {tab.label}
            </button>
          ))}
        </div>
      </div>

      {/* Table */}
      <div className="rounded-xl border border-zinc-800 overflow-hidden">
        <Table>
          <TableHeader>
            <TableRow className="hover:bg-transparent">
              <TableHead>Node</TableHead>
              <TableHead>Type</TableHead>
              <TableHead>Host</TableHead>
              <TableHead>Last Backup</TableHead>
              <TableHead>Files</TableHead>
              <TableHead>Status</TableHead>
              <TableHead className="w-8" />
            </TableRow>
          </TableHeader>
          <TableBody>
            {visible.length === 0 && (
              <TableRow>
                <TableCell colSpan={7} className="text-center text-zinc-500 py-10">
                  No nodes match the current filter.
                </TableCell>
              </TableRow>
            )}
            {visible.map((node) => (
              <TableRow key={node.name}>
                <TableCell>
                  <Link
                    to={`/nodes/${encodeURIComponent(node.name)}`}
                    className="font-medium text-zinc-100 hover:text-indigo-400 transition-colors"
                  >
                    {node.name}
                  </Link>
                </TableCell>
                <TableCell>
                  <Badge variant="default" className="text-xs">
                    {node.type}
                  </Badge>
                </TableCell>
                <TableCell>
                  <span className="font-mono text-xs text-zinc-400">{node.host}</span>
                </TableCell>
                <TableCell>
                  <span className="text-xs text-zinc-400">
                    {node.last_time ? (
                      <span title={formatDateTime(node.last_time)}>
                        {formatRelativeTime(node.last_time)}
                      </span>
                    ) : (
                      '—'
                    )}
                  </span>
                </TableCell>
                <TableCell>
                  <span className="text-xs text-zinc-500">{node.backup_files.length}</span>
                </TableCell>
                <TableCell>
                  <StatusBadge status={node.last_status} />
                </TableCell>
                <TableCell>
                  <Link
                    to={`/nodes/${encodeURIComponent(node.name)}`}
                    className="text-zinc-600 hover:text-zinc-300 transition-colors"
                  >
                    <ChevronRight className="w-4 h-4" />
                  </Link>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </div>
      <p className="text-xs text-zinc-600 px-1">
        Showing {visible.length} of {nodes.length} nodes
      </p>
    </div>
  )
}
