import { useEffect, useState } from 'react'
import { ChevronLeft, ChevronRight } from 'lucide-react'

// Single-image galleries render statically (no dots/autoplay/arrow chrome
// for a carousel of one) — this stays a real carousel the moment a second
// image is added to `images`, so callers can start with one slide today
// without a later rewrite.
function Carousel({ images, className = '', imgClassName = '', intervalMs = 5000 }) {
  const [index, setIndex] = useState(0)
  const multiple = images.length > 1

  useEffect(() => {
    if (!multiple) return undefined
    const id = setInterval(() => setIndex((i) => (i + 1) % images.length), intervalMs)
    return () => clearInterval(id)
  }, [multiple, images.length, intervalMs])

  if (images.length === 0) return null

  const goPrev = () => setIndex((i) => (i - 1 + images.length) % images.length)
  const goNext = () => setIndex((i) => (i + 1) % images.length)

  return (
    <div className={`relative overflow-hidden rounded-2xl ${className}`}>
      {images.map((img, i) => (
        <img
          key={img.src}
          src={img.src}
          alt={img.alt || ''}
          className={`${
            i === 0 ? 'relative' : 'absolute inset-0'
          } w-full transition-opacity duration-700 ${
            i === index ? 'opacity-100' : 'pointer-events-none opacity-0'
          } ${imgClassName}`}
        />
      ))}

      {multiple && (
        <>
          <button
            type="button"
            aria-label="Previous slide"
            onClick={goPrev}
            className="absolute left-3 top-1/2 -translate-y-1/2 rounded-full bg-black/40 p-1.5 text-white transition-colors hover:bg-black/60"
          >
            <ChevronLeft size={20} />
          </button>
          <button
            type="button"
            aria-label="Next slide"
            onClick={goNext}
            className="absolute right-3 top-1/2 -translate-y-1/2 rounded-full bg-black/40 p-1.5 text-white transition-colors hover:bg-black/60"
          >
            <ChevronRight size={20} />
          </button>

          <div className="absolute inset-x-0 bottom-3 flex items-center justify-center gap-1.5">
            {images.map((img, i) => (
              <button
                key={img.src}
                type="button"
                aria-label={`Slide ${i + 1}`}
                onClick={() => setIndex(i)}
                className={`h-1.5 rounded-full transition-all ${i === index ? 'w-5 bg-gold' : 'w-1.5 bg-white/40'}`}
              />
            ))}
          </div>
        </>
      )}
    </div>
  )
}

export default Carousel
