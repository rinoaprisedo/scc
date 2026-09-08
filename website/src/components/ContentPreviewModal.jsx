import { X } from 'lucide-react'
import { fileURL } from '../utils/url'
import Carousel from './ui/Carousel'

function isPdf(path) {
  return !!path && path.toLowerCase().endsWith('.pdf')
}

// event_information_file/about_malaysia_file store a JSON array of paths
// (multi-image, rendered as a Carousel slider) — agenda_file/dress_code_file
// still store a single path/PDF. A value that isn't valid JSON is a
// pre-multi-image legacy single path, not a parse failure, mirroring the
// backend's own settings.parseImagePaths.
function parsePaths(raw) {
  if (!raw) return []
  try {
    const parsed = JSON.parse(raw)
    return Array.isArray(parsed) ? parsed : [raw]
  } catch {
    return [raw]
  }
}

function ContentPreviewModal({ title, path, onClose }) {
  const paths = parsePaths(path)
  const pdf = paths.length === 1 && isPdf(paths[0])

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-navy-dark/80 p-4 backdrop-blur-sm">
      <div className="relative flex h-[90vh] w-[90vw] items-center justify-center">
        <button
          type="button"
          onClick={onClose}
          aria-label="Close"
          className="absolute right-0 top-0 z-10 rounded-full bg-black/50 p-2 text-white transition-colors hover:bg-black/70"
        >
          <X size={20} />
        </button>

        {paths.length === 0 ? (
          <p className="px-6 text-center text-sm text-white/80">Konten belum tersedia. Silakan cek kembali nanti.</p>
        ) : pdf ? (
          <iframe src={fileURL(paths[0])} title={title} className="h-full w-full rounded-lg" />
        ) : (
          <Carousel
            images={paths.map((p) => ({ src: fileURL(p), alt: title }))}
            className="h-full w-full"
            imgClassName="h-full w-full object-contain"
          />
        )}
      </div>
    </div>
  )
}

export default ContentPreviewModal
