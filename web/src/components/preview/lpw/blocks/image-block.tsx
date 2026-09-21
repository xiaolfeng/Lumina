import type React from 'react'
import { useState } from 'react'
import type { LpwBlockSlotProps, LpwImageProps } from '../types'

export const ImageBlock: React.FC<LpwBlockSlotProps<LpwImageProps>> = ({
  props,
}) => {
  const [hasError, setHasError] = useState(false)

  if (hasError) {
    return (
      <figure
        data-testid="image-fallback"
        className="my-6 border border-dashed border-line bg-surface-muted/30 p-6 text-center text-xs text-sea-ink-soft shadow-2xs font-sans"
      >
        <div className="font-serif font-semibold text-sm text-sea-ink">
          {props.alt}
        </div>
        <div className="mt-1 font-mono text-[11px] text-sea-ink-soft/70">
          图片加载失败: {props.src}
        </div>
      </figure>
    )
  }

  return (
    <figure className="my-8 flex flex-col items-center font-sans">
      <div className="border border-line bg-surface p-1.5 shadow-2xs">
        <img
          src={props.src}
          alt={props.alt}
          style={props.width ? { width: props.width } : undefined}
          onError={() => setHasError(true)}
          className="max-w-full object-contain bg-surface-muted/20"
        />
      </div>
      {props.caption && (
        <figcaption className="mt-2.5 text-center text-xs font-serif italic text-sea-ink-soft tracking-wide">
          {props.caption}
        </figcaption>
      )}
    </figure>
  )
}
