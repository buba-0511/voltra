import { createApi } from '@reduxjs/toolkit/query/react'
import { axiosBaseQuery } from './axiosBaseQuery'
import type {
  AnalysisRun,
  Anomaly,
  DailyPoint,
  DashboardSummary,
  LoginResponse,
  MeterDailyRow,
  MeterDetail,
  MeterSummary,
  ReadingsPage,
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
    logout: builder.mutation<void, void>({
      query: () => ({ url: '/auth/logout', method: 'POST' }),
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
    getMeterReadings: builder.query<ReadingsPage, { meterId: string; cursor?: string; limit?: number }>({
      query: ({ meterId, cursor, limit }) => ({
        url: `/meters/${meterId}/readings`,
        params: { cursor, limit },
      }),
    }),
    getMeterDaily: builder.query<DailyPoint[], string>({
      query: (meterId) => ({ url: `/meters/${meterId}/daily` }),
      providesTags: (_result, _error, meterId) => [{ type: 'Meter', id: meterId }],
    }),
    getAllMetersDaily: builder.query<MeterDailyRow[], void>({
      query: () => ({ url: '/meters/daily' }),
      providesTags: ['Meter'],
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
      // Longer timeout than the default: this reads and scores every
      // meter server-side, so it legitimately takes longer as the
      // tenant's meter count grows.
      query: () => ({ url: '/ai/analyze', method: 'POST', timeout: 120_000 }),
      // A fresh analysis run changes meter status/variation, the
      // anomalies list, and the dashboard KPIs all at once.
      invalidatesTags: ['Meter', 'Anomaly', 'Dashboard', 'AnalysisRun'],
    }),
    getAnalysisRun: builder.query<AnalysisRun, number>({
      query: (id) => ({ url: `/ai/analysis/${id}` }),
      providesTags: (_result, _error, id) => [{ type: 'AnalysisRun', id }],
    }),
    updateAnomalyStatus: builder.mutation<Anomaly, { id: number; status: string }>({
      query: ({ id, status }) => ({ url: `/anomalies/${id}`, method: 'PATCH', data: { status } }),
      invalidatesTags: (_result, _error, { id }) => [{ type: 'Anomaly', id }, 'Anomaly'],
    }),
  }),
})

export const {
  useLoginMutation,
  useLogoutMutation,
  useMeQuery,
  useListMetersQuery,
  useGetMeterQuery,
  useGetMeterReadingsQuery,
  useGetMeterDailyQuery,
  useGetAllMetersDailyQuery,
  useDashboardSummaryQuery,
  useListAnomaliesQuery,
  useGetAnomalyQuery,
  useRunAnalysisMutation,
  useGetAnalysisRunQuery,
  useUpdateAnomalyStatusMutation,
} = apiSlice
