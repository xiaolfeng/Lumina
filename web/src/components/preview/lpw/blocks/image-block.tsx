import React, { useState } from 'react'
import type { LpwBlockSlotProps, LpwImageProps } from '../types'

export const ImageBlock: React.FC<LpwBlockSlotProps<LpwImageProps>> = ({
  props,
}) => {
  const [hasError, setHasError] = useState(false)

  if (hasError) {
    return (
      <figure
        data-testid="image-fallback"
        className="my-4 border border-dashed border-line bg-surface-muted/20 p-4 text-center text-xs text-sea-ink-soft"
      >
        <div className="font-semibold text-sea-ink">{props.alt}</div>
        <div className="mt-1 font-mono text-[11px] text-sea-ink-soft/70">
          图片加载失败: {props.src}
        </div>
      </figure>
    )
  }

  return (
    <figure className="my-4 flex flex-col items-center">
      <img
        src={props.src}
        alt={props.alt}
        style={props.width ? { width: props.width } : undefined}
        onError={() => setHasError(true)}
        className="max-w-full object-contain"
      />
      {props.caption && (
        <figcaption className="mt-1.5 text-center text-xs text-sea-ink-soft">
          {props.caption}
        </figcaption>
      )}
    </figure>
  )
}
