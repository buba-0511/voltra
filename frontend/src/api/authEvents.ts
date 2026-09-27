// Small pub/sub so axiosInstance's 401 handler can notify the Redux store
// without importing it directly (store -> apiSlice -> axiosInstance would
// otherwise become circular). store.ts registers the handler once, after
// it's created.
let onUnauthorized: (() => void) | null = null

export function setUnauthorizedHandler(fn: () => void) {
  onUnauthorized = fn
}

export function notifyUnauthorized() {
  onUnauthorized?.()
}
