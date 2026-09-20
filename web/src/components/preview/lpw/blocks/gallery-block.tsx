import React, { useState } from 'react'
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

  if (hasError) {
    return (
      <figure
        data-testid={`gallery-fallback-${index}`}
        className="flex h-40 w-full flex-col items-center justify-center border border-dashed border-line bg-surface-muted/20 p-2 text-center text-xs text-sea-ink-soft"
      >
        <span className="font-semibold text-sea-ink">{image.alt}</span>
        <span className="mt-1 font-mono text-[10px] text-sea-ink-soft/70">
          加载失败: {image.src}
        </span>
      </figure>
    )
  }

  return (
    <figure data-testid={`gallery-item-${index}`} className="flex flex-col">
      <img
        src={image.src}
        alt={image.alt}
        onError={() => setHasError(true)}
        className="h-40 w-full border border-line bg-surface-muted/30 object-cover"
      />
      {image.caption && (
        <figcaption className="mt-1.5 text-center text-xs text-sea-ink-soft">
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
      className="my-4 grid grid-cols-2 gap-4 lg:grid-cols-3"
    >
      {props.images.map((img, idx) => (
        <GalleryItem key={idx} image={img} index={idx} />
      ))}
    </div>
  )
}
