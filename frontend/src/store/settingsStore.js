import { create } from 'zustand'
import { getPublicSettings } from '../api/settings'
import { fileURL } from '../utils/url'
import { applyPrimaryColor } from '../utils/color'

const DEFAULT_PRIMARY = '#c2622e'

const useSettingsStore = create((set) => ({
  appName: 'BaseAdmin',
  logo: null,
  favicon: null,
  primaryColor: DEFAULT_PRIMARY,
  // Branding only — uses the unauthenticated /settings/public endpoint so it
  // resolves correctly on the login page too, not just after signing in.
  fetchSettings: async () => {
    try {
      const res = await getPublicSettings()
      // GET /settings/public returns a { key: value } map directly, not an array.
      const map = res.data || {}
      const primaryColor = map.primary_color || DEFAULT_PRIMARY
      applyPrimaryColor(primaryColor)
      set({
        appName: map.app_name || 'BaseAdmin',
        logo: fileURL(map.app_logo),
        favicon: fileURL(map.app_favicon),
        primaryColor,
      })
    } catch {
      // settings API not reachable — keep defaults
    }
  },
  setPrimaryColor: (hex) => {
    applyPrimaryColor(hex)
    set({ primaryColor: hex })
  },
}))

export default useSettingsStore
