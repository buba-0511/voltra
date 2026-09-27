import axios from 'axios'
import { notifyUnauthorized } from './authEvents'
import { clearToken, getToken } from './tokenStorage'

const API_URL = import.meta.env.VITE_API_URL ?? 'http://localhost:8080'

export const axiosInstance = axios.create({
  baseURL: API_URL,
})

// Request interceptor: attach the bearer token to every call, so no
// endpoint has to remember to do it itself.
axiosInstance.interceptors.request.use((config) => {
  const token = getToken()
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

// Response interceptor: any 401 means the token is missing/expired -
// clear it globally so ProtectedRoute's redirect-to-/login kicks in on
// the next render, instead of every caller handling this itself.
axiosInstance.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      clearToken()
      notifyUnauthorized()
    }
    return Promise.reject(error)
  },
)
