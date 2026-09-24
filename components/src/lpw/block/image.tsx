import type React from 'react'
import { useEffect, useState } from 'react'
import { resolveAssetSrcPure, useLpwRuntime } from '../runtime-provider'
import type { LpwBlockSlotProps, LpwImageProps } from '../types'

export const ImageBlock: React.FC<LpwBlockSlotProps<LpwImageProps>> = ({
  props,
  location,
}) => {
  const [hasError, setHasError] = useState(false)
  const runtime = useLpwRuntime()
  const resolvedSrc = resolveAssetSrcPure(props.src, runtime)

  useEffect(() => {
    setHasError(false)
  }, [props.src])

  const isNested = Boolean(location?.idPath && location.idPath.length > 0)
  const targetWidth = props.width ? props.width : isNested ? '100%' : '80%'

  if (hasError) {
    return (
      <figure
        data-testid="image-fallback"
        className="max-w-full min-w-0 border border-dashed border-line bg-surface-muted/30 p-6 text-center text-xs text-sea-ink-soft font-sans"
      >
        <div className="font-serif font-semibold text-sm text-sea-ink">
          {props.alt}
        </div>
        <div className="mt-1 font-mono text-[11px] text-sea-ink-soft/70 break-all">
          图片加载失败: {props.src}
        </div>
      </figure>
    )
  }

  return (
    <figure
      data-testid="image-block"
      className="flex max-w-full min-w-0 flex-col items-center font-sans"
      style={{ width: targetWidth, maxWidth: '100%', margin: '0 auto' }}
    >
      <div className="w-full max-w-full min-w-0 border border-line bg-surface p-1.5 flex justify-center">
        <img
          src={resolvedSrc}
          alt={props.alt}
          style={props.width ? { width: props.width } : undefined}
          onError={() => setHasError(true)}
          className="w-full max-w-full object-contain bg-surface-muted/20"
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
