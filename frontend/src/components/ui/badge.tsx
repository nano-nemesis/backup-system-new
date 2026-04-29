import * as React from 'react'
import { cva, type VariantProps } from 'class-variance-authority'
import { cn } from '@/lib/utils'

const badgeVariants = cva(
  'inline-flex items-center gap-1.5 rounded-full border px-2.5 py-0.5 text-xs font-semibold transition-colors',
  {
    variants: {
      variant: {
        default: 'border-zinc-700 bg-zinc-800 text-zinc-300',
        success: 'border-green-800 bg-green-950 text-green-400',
        error: 'border-red-800 bg-red-950 text-red-400',
        warning: 'border-amber-800 bg-amber-950 text-amber-400',
        unknown: 'border-zinc-700 bg-zinc-800 text-zinc-500',
        admin: 'border-indigo-800 bg-indigo-950 text-indigo-400',
        viewer: 'border-zinc-700 bg-zinc-800 text-zinc-400',
      },
    },
    defaultVariants: { variant: 'default' },
  },
)

export interface BadgeProps
  extends React.HTMLAttributes<HTMLDivElement>,
    VariantProps<typeof badgeVariants> {}

function Badge({ className, variant, ...props }: BadgeProps) {
  return <div className={cn(badgeVariants({ variant }), className)} {...props} />
}

export { Badge, badgeVariants }
