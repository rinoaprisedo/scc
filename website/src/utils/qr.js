// Printed QR gates encode "<website>/<code>" so a phone's native camera opens
// the site directly; the in-app scanner reads the same QR, so it has to
// accept both that URL form and a bare legacy code.
export function extractQrCode(text) {
  const raw = (text || '').trim()
  if (!/^https?:\/\//i.test(raw)) return raw
  try {
    const segments = new URL(raw).pathname.split('/').filter(Boolean)
    return segments.length > 0 ? decodeURIComponent(segments[segments.length - 1]) : raw
  } catch {
    return raw
  }
}

// sessionStorage so a camera-scanned code survives the login detour
// (/:code → /dashboard → / → login → /dashboard) but not a new tab/session.
const PENDING_KEY = 'pending_qr_code'

export const setPendingQrCode = (code) => sessionStorage.setItem(PENDING_KEY, code)

export function takePendingQrCode() {
  const code = sessionStorage.getItem(PENDING_KEY)
  sessionStorage.removeItem(PENDING_KEY)
  return code
}
