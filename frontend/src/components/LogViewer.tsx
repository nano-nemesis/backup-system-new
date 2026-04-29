import { useEffect, useRef } from 'react'
import { cn } from '@/lib/utils'

function lineClass(line: string): string {
  const lower = line.toLowerCase()
  if (lower.includes('error')) return 'text-red-600 dark:text-red-400'
  if (lower.includes('success')) return 'text-green-600 dark:text-green-400'
  return 'text-muted'
}

export default function LogViewer({ lines }: { lines: string[] }) {
  const bottomRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [lines])

  return (
    <div className="rounded-xl border border-line bg-surface-2 overflow-hidden">
      <div className="flex items-center gap-1.5 px-4 py-2.5 border-b border-line">
        <div className="w-2.5 h-2.5 rounded-full bg-red-400/60" />
        <div className="w-2.5 h-2.5 rounded-full bg-amber-400/60" />
        <div className="w-2.5 h-2.5 rounded-full bg-green-400/60" />
        <span className="ml-2 text-xs text-muted">node.log ({lines.length} lines)</span>
      </div>
      <div className="overflow-y-auto max-h-96 p-4 scrollbar-thin">
        {lines.length === 0 ? (
          <p className="text-xs text-muted italic">No log entries found.</p>
        ) : (
          <div className="space-y-0.5">
            {lines.map((line, i) => (
              <p
                key={i}
                className={cn('font-mono text-xs leading-5 whitespace-pre-wrap break-all', lineClass(line))}
              >
                {line}
              </p>
            ))}
          </div>
        )}
        <div ref={bottomRef} />
      </div>
    </div>
  )
}
