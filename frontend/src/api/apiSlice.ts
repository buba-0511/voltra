import { createApi } from '@reduxjs/toolkit/query/react'
import { axiosBaseQuery } from './axiosBaseQuery'
import type {
  AnalysisRun,
  Anomaly,
  DashboardSummary,
  LoginResponse,
  MeterDetail,
  MeterSummary,
  Reading,
  User,
} from './types'

export const apiSlice = createApi({
  reducerPath: 'api',
  baseQuery: axiosBaseQuery(),
  tagTypes: ['Meter', 'Anomaly', 'Dashboard', 'AnalysisRun'],
  endpoints: (builder) => ({
    login: builder.mutation<LoginResponse, { email: string; password: string }>({
      query: (body) => ({ url: '/auth/login', method: 'POST', data: body }),
    }),
    me: builder.query<User, void>({
      query: () => ({ url: '/auth/me' }),
    }),

    listMeters: builder.query<MeterSummary[], void>({
      query: () => ({ url: '/meters' }),
      providesTags: ['Meter'],
    }),
    getMeter: builder.query<MeterDetail, string>({
      query: (meterId) => ({ url: `/meters/${meterId}` }),
      providesTags: (_result, _error, meterId) => [{ type: 'Meter', id: meterId }],
    }),
    getMeterReadings: builder.query<Reading[], string>({
      query: (meterId) => ({ url: `/meters/${meterId}/readings` }),
    }),

    dashboardSummary: builder.query<DashboardSummary, void>({
      query: () => ({ url: '/dashboard/summary' }),
      providesTags: ['Dashboard'],
    }),

    listAnomalies: builder.query<Anomaly[], void>({
      query: () => ({ url: '/anomalies' }),
      providesTags: ['Anomaly'],
    }),
    getAnomaly: builder.query<Anomaly, number>({
      query: (id) => ({ url: `/anomalies/${id}` }),
      providesTags: (_result, _error, id) => [{ type: 'Anomaly', id }],
    }),
    runAnalysis: builder.mutation<AnalysisRun, void>({
      query: () => ({ url: '/ai/analyze', method: 'POST' }),
      // A fresh analysis run changes meter status/variation, the
      // anomalies list, and the dashboard KPIs all at once.
      invalidatesTags: ['Meter', 'Anomaly', 'Dashboard', 'AnalysisRun'],
    }),
    getAnalysisRun: builder.query<AnalysisRun, number>({
      query: (id) => ({ url: `/ai/analysis/${id}` }),
      providesTags: (_result, _error, id) => [{ type: 'AnalysisRun', id }],
    }),
  }),
})

export const {
  useLoginMutation,
  useMeQuery,
  useListMetersQuery,
  useGetMeterQuery,
  useGetMeterReadingsQuery,
  useDashboardSummaryQuery,
  useListAnomaliesQuery,
  useGetAnomalyQuery,
  useRunAnalysisMutation,
  useGetAnalysisRunQuery,
} = apiSlice
