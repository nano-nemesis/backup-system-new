import { useEffect, useRef } from 'react'
import { cn } from '@/lib/utils'

interface Props {
  lines: string[]
}

function lineClass(line: string): string {
  const lower = line.toLowerCase()
  if (lower.includes('error')) return 'text-red-400'
  if (lower.includes('success')) return 'text-green-400'
  return 'text-zinc-400'
}

export default function LogViewer({ lines }: Props) {
  const bottomRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [lines])

  return (
    <div className="rounded-xl border border-zinc-800 bg-zinc-950 overflow-hidden">
      <div className="flex items-center gap-1.5 px-4 py-2.5 border-b border-zinc-800">
        <div className="w-2.5 h-2.5 rounded-full bg-red-500/60" />
        <div className="w-2.5 h-2.5 rounded-full bg-amber-500/60" />
        <div className="w-2.5 h-2.5 rounded-full bg-green-500/60" />
        <span className="ml-2 text-xs text-zinc-600">node.log (last {lines.length} lines)</span>
      </div>
      <div className="overflow-y-auto max-h-96 p-4 scrollbar-thin">
        {lines.length === 0 ? (
          <p className="text-xs text-zinc-600 italic">No log entries found.</p>
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
