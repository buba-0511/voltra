import type { BaseQueryFn } from '@reduxjs/toolkit/query'
import type { AxiosError, AxiosRequestConfig } from 'axios'
import { axiosInstance } from './axiosInstance'

type AxiosBaseQueryArgs = {
  url: string
  method?: AxiosRequestConfig['method']
  data?: AxiosRequestConfig['data']
  params?: AxiosRequestConfig['params']
  timeout?: AxiosRequestConfig['timeout']
}

export const axiosBaseQuery = (): BaseQueryFn<AxiosBaseQueryArgs, unknown, string> =>
  async ({ url, method = 'GET', data, params, timeout }) => {
    try {
      const result = await axiosInstance({ url, method, data, params, timeout })
      return { data: result.data }
    } catch (err) {
      const error = err as AxiosError<{ error?: string }>
      if (error.code === 'ECONNABORTED') {
        return { error: 'La operación está tardando más de lo esperado. Probá de nuevo en un momento.' }
      }
      return {
        error: error.response?.data?.error ?? error.message ?? 'Request failed',
      }
    }
  }
