// The JWT itself lives in an HttpOnly cookie now (set by the backend on
// login), so client-side JS never has access to it - that's the whole
// point, it can't be stolen by an injected script. This flag is not a
// credential: it only remembers "did the user log in on this browser"
// so the UI can skip straight to /login instead of flashing the app
// shell before the first /auth/me call comes back. The cookie is what
// actually authenticates every request; if this flag is ever wrong
// (tampered with, stale), the next API call 401s and clears it.
const FLAG_KEY = 'energy_platform_session'

export function hasSession(): boolean {
  try {
    return localStorage.getItem(FLAG_KEY) === '1'
  } catch {
    return false
  }
}

export function setSession() {
  try {
    localStorage.setItem(FLAG_KEY, '1')
  } catch {
    // ignore - private browsing / blocked storage, session just won't persist
  }
}

export function clearSession() {
  try {
    localStorage.removeItem(FLAG_KEY)
  } catch {
    // ignore
  }
}
