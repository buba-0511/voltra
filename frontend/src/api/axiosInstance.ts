import axios from 'axios'
import { notifyUnauthorized } from './authEvents'
import { clearSession } from './sessionFlag'

const API_URL = import.meta.env.VITE_API_URL ?? 'http://localhost:8080'

// Default timeout so a hung request fails with a clear error instead of
// spinning forever with no feedback; endpoints that can legitimately take
// longer (e.g. running the AI analysis over many meters) override it
// per-call via axiosBaseQuery's `timeout` arg.
//
// withCredentials: the JWT lives in an HttpOnly cookie the backend sets
// on login - the browser attaches it automatically on every request to
// that origin, nothing here needs to read or attach it manually.
export const axiosInstance = axios.create({
  baseURL: API_URL,
  timeout: 15_000,
  withCredentials: true,
})

// Response interceptor: any 401 means the session cookie is missing/
// expired - clear the local "logged in" flag globally so ProtectedRoute's
// redirect-to-/login kicks in on the next render, instead of every caller
// handling this itself.
axiosInstance.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      clearSession()
      notifyUnauthorized()
    }
    return Promise.reject(error)
  },
)
