import type { BaseQueryFn } from '@reduxjs/toolkit/query'
import type { AxiosError, AxiosRequestConfig } from 'axios'
import { axiosInstance } from './axiosInstance'

type AxiosBaseQueryArgs = {
  url: string
  method?: AxiosRequestConfig['method']
  data?: AxiosRequestConfig['data']
  params?: AxiosRequestConfig['params']
}

export const axiosBaseQuery = (): BaseQueryFn<AxiosBaseQueryArgs, unknown, string> =>
  async ({ url, method = 'GET', data, params }) => {
    try {
      const result = await axiosInstance({ url, method, data, params })
      return { data: result.data }
    } catch (err) {
      const error = err as AxiosError<{ error?: string }>
      return {
        error: error.response?.data?.error ?? error.message ?? 'Request failed',
      }
    }
  }
