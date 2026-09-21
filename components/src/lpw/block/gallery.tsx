import type React from 'react'
import { useState } from 'react'
import { resolveAssetSrcPure, useLpwRuntime } from '../runtime-provider'
import type {
  LpwBlockSlotProps,
  LpwGalleryImage,
  LpwGalleryProps,
} from '../types'

const GalleryItem: React.FC<{ image: LpwGalleryImage; index: number }> = ({
  image,
  index,
}) => {
  const [hasError, setHasError] = useState(false)
  const runtime = useLpwRuntime()
  const resolvedSrc = resolveAssetSrcPure(image.src, runtime)

  if (hasError) {
    return (
      <figure
        data-testid={`gallery-fallback-${index}`}
        className="flex h-44 w-full flex-col items-center justify-center border border-dashed border-line bg-surface-muted/30 p-3 text-center text-xs text-sea-ink-soft shadow-2xs font-sans"
      >
        <span className="font-serif font-semibold text-sea-ink">
          {image.alt}
        </span>
        <span className="mt-1 font-mono text-[10px] text-sea-ink-soft/70 break-all">
          加载失败: {image.src}
        </span>
      </figure>
    )
  }

  return (
    <figure
      data-testid={`gallery-item-${index}`}
      className="group flex flex-col font-sans"
    >
      <div className="border border-line bg-surface p-1 shadow-2xs transition-all duration-150 group-hover:border-sea-ink/60">
        <img
          src={resolvedSrc}
          alt={image.alt}
          onError={() => setHasError(true)}
          className="w-full max-h-72 sm:h-40 object-contain sm:object-cover bg-surface-muted/20"
        />
      </div>
      {image.caption && (
        <figcaption className="mt-2 text-center text-xs font-serif italic text-sea-ink-soft">
          {image.caption}
        </figcaption>
      )}
    </figure>
  )
}

export const GalleryBlock: React.FC<LpwBlockSlotProps<LpwGalleryProps>> = ({
  props,
}) => {
  return (
    <div
      data-testid="gallery-block"
      className="my-8 grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4 sm:gap-6"
    >
      {props.images.map((img, idx) => (
        <GalleryItem key={`${img.src}-${idx}`} image={img} index={idx} />
      ))}
    </div>
  )
}
