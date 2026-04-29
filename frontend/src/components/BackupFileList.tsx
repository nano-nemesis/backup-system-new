import { Download, FileArchive } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { formatDateTime } from '@/lib/utils'
import type { BackupFile } from '@/types'

export default function BackupFileList({ files }: { files: BackupFile[] }) {
  if (files.length === 0) {
    return (
      <div className="rounded-xl border border-line px-4 py-8 text-center">
        <FileArchive className="w-8 h-8 text-muted/40 mx-auto mb-2" />
        <p className="text-sm text-muted">No backup files found</p>
      </div>
    )
  }

  return (
    <div className="rounded-xl border border-line divide-y divide-line overflow-hidden">
      {files.map(f => (
        <div
          key={f.name}
          className="flex items-center justify-between px-4 py-3 hover:bg-surface-2/60 transition-colors"
        >
          <div className="min-w-0">
            <p className="text-sm font-medium text-fg truncate">{f.name}</p>
            <p className="text-xs text-muted mt-0.5">
              {f.size_str} · {formatDateTime(f.mod_time)}
            </p>
          </div>
          <Button asChild variant="ghost" size="sm" className="ml-4 shrink-0">
            <a href={`/storage/${f.rel_path}`} download>
              <Download className="w-3.5 h-3.5" />
              Download
            </a>
          </Button>
        </div>
      ))}
    </div>
  )
}
