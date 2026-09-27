import { createSlice, type PayloadAction } from '@reduxjs/toolkit'
import { clearSession, hasSession, setSession } from '../api/sessionFlag'
import type { User } from '../api/types'

interface AuthState {
  user: User | null
  hasSession: boolean
}

const initialState: AuthState = {
  user: null,
  hasSession: hasSession(),
}

const authSlice = createSlice({
  name: 'auth',
  initialState,
  reducers: {
    credentialsSet(state, action: PayloadAction<{ user: User }>) {
      setSession()
      state.user = action.payload.user
      state.hasSession = true
    },
    userLoaded(state, action: PayloadAction<User>) {
      state.user = action.payload
    },
    loggedOut(state) {
      clearSession()
      state.user = null
      state.hasSession = false
    },
  },
})

export const { credentialsSet, userLoaded, loggedOut } = authSlice.actions
export default authSlice.reducer
