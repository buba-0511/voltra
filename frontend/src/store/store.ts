import { configureStore } from '@reduxjs/toolkit'
import { apiSlice } from '../api/apiSlice'
import { setUnauthorizedHandler } from '../api/authEvents'
import authReducer, { loggedOut } from './authSlice'

export const store = configureStore({
  reducer: {
    auth: authReducer,
    [apiSlice.reducerPath]: apiSlice.reducer,
  },
  middleware: (getDefaultMiddleware) => getDefaultMiddleware().concat(apiSlice.middleware),
})

setUnauthorizedHandler(() => store.dispatch(loggedOut()))

export type RootState = ReturnType<typeof store.getState>
export type AppDispatch = typeof store.dispatch
