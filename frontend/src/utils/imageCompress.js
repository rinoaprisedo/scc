// Client-side compression before upload — resizes to fit within maxDimension
// and re-encodes as JPEG, so a multi-MB phone photo doesn't hit the backend's
// 10MB request size limit or eat into StorageMaxSize. Falls back to the
// original file for any type canvas can't decode (e.g. already-tiny images
// where compression would only add risk for no benefit).
export async function compressImage(file, { maxDimension = 1600, quality = 0.75 } = {}) {
  if (!file.type.startsWith('image/')) return file

  const bitmap = await createImageBitmap(file).catch(() => null)
  if (!bitmap) return file

  const scale = Math.min(1, maxDimension / Math.max(bitmap.width, bitmap.height))
  const width = Math.round(bitmap.width * scale)
  const height = Math.round(bitmap.height * scale)

  const canvas = document.createElement('canvas')
  canvas.width = width
  canvas.height = height
  const ctx = canvas.getContext('2d')
  ctx.drawImage(bitmap, 0, 0, width, height)
  bitmap.close?.()

  const blob = await new Promise((resolve) => canvas.toBlob(resolve, 'image/jpeg', quality))
  if (!blob || blob.size >= file.size) return file

  const name = file.name.replace(/\.\w+$/, '') + '.jpg'
  return new File([blob], name, { type: 'image/jpeg' })
}
