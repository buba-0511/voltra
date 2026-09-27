import { createSlice, type PayloadAction } from '@reduxjs/toolkit'
import { clearToken, getToken, setToken } from '../api/tokenStorage'
import type { User } from '../api/types'

interface AuthState {
  user: User | null
  hasToken: boolean
}

const initialState: AuthState = {
  user: null,
  hasToken: !!getToken(),
}

const authSlice = createSlice({
  name: 'auth',
  initialState,
  reducers: {
    credentialsSet(state, action: PayloadAction<{ token: string; user: User }>) {
      setToken(action.payload.token)
      state.user = action.payload.user
      state.hasToken = true
    },
    userLoaded(state, action: PayloadAction<User>) {
      state.user = action.payload
    },
    loggedOut(state) {
      clearToken()
      state.user = null
      state.hasToken = false
    },
  },
})

export const { credentialsSet, userLoaded, loggedOut } = authSlice.actions
export default authSlice.reducer
