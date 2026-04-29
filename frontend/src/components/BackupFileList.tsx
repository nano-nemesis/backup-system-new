import { Download, FileArchive } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { formatDateTime } from '@/lib/utils'
import type { BackupFile } from '@/types'

interface Props {
  files: BackupFile[]
}

export default function BackupFileList({ files }: Props) {
  if (files.length === 0) {
    return (
      <div className="rounded-xl border border-zinc-800 px-4 py-8 text-center">
        <FileArchive className="w-8 h-8 text-zinc-700 mx-auto mb-2" />
        <p className="text-sm text-zinc-500">No backup files found</p>
      </div>
    )
  }

  return (
    <div className="rounded-xl border border-zinc-800 divide-y divide-zinc-800 overflow-hidden">
      {files.map((f) => (
        <div key={f.name} className="flex items-center justify-between px-4 py-3 hover:bg-zinc-800/30 transition-colors">
          <div className="min-w-0">
            <p className="text-sm font-medium text-zinc-200 truncate">{f.name}</p>
            <p className="text-xs text-zinc-500 mt-0.5">
              {f.size_str} &middot; {formatDateTime(f.mod_time)}
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
