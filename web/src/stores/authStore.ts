import { create } from 'zustand';
import axios from 'axios';
import type { User, Role } from '../types/api';
import { login as apiLogin, getMe } from '../api/auth';

const API_BASE = import.meta.env.VITE_API_BASE || '';

interface AuthState {
  user: User | null;
  // access_token is stored in memory only (never localStorage) to prevent XSS token theft
  accessToken: string | null;
  isAuthenticated: boolean;
  // isLoading starts true so ProtectedRoute waits for loadUser() to complete on refresh
  isLoading: boolean;
  login: (employeeId: string, password: string) => Promise<void>;
  logout: () => void;
  loadUser: () => Promise<void>;
  setAccessToken: (token: string | null) => void;
  hasRole: (...roles: Role[]) => boolean;
  canDispatch: () => boolean;
  isAdmin: () => boolean;
}

export const useAuthStore = create<AuthState>((set, get) => ({
  user: null,
  // Access token lives in Zustand memory only — never written to localStorage
  accessToken: null,
  isAuthenticated: false,
  // Start true: App must await loadUser() before rendering protected routes
  isLoading: true,

  setAccessToken: (token: string | null) => {
    set({ accessToken: token });
  },

  login: async (employeeId: string, password: string) => {
    set({ isLoading: true });
    try {
      const resp = await apiLogin(employeeId, password);
      // Keep access_token in memory only — eliminated from localStorage to prevent XSS access
      set({ accessToken: resp.access_token, user: resp.user, isAuthenticated: true });
      // TODO: Migrate refresh_token to httpOnly cookie set by the server to eliminate
      // all client-side token storage. Currently stored in localStorage as an interim measure.
      localStorage.setItem('refresh_token', resp.refresh_token);
    } finally {
      set({ isLoading: false });
    }
  },

  logout: () => {
    // TODO: When refresh_token moves to httpOnly cookie, call a server-side /auth/logout
    // endpoint here to invalidate/clear the cookie instead of removing from localStorage.
    localStorage.removeItem('refresh_token');
    set({ user: null, accessToken: null, isAuthenticated: false });
  },

  loadUser: async () => {
    // On page refresh the access_token (memory) is gone; attempt refresh from stored token
    const refreshToken = localStorage.getItem('refresh_token');

    if (!refreshToken) {
      // No refresh token available — user must log in again
      set({ isLoading: false });
      return;
    }

    set({ isLoading: true });
    try {
      // Exchange the refresh token for a new access token
      const { data } = await axios.post(`${API_BASE}/api/v1/auth/refresh`, {
        refresh_token: refreshToken,
      });
      const newAccessToken: string = data.access_token;

      // Store the new access token in memory only
      set({ accessToken: newAccessToken });

      // Update stored refresh token if the server rotated it
      if (data.refresh_token) {
        // TODO: Remove this localStorage write once server sets httpOnly cookie
        localStorage.setItem('refresh_token', data.refresh_token);
      }

      // Fetch the user profile using the new access token
      // getMe() will use the token via the axios interceptor reading from this store
      const user = await getMe();
      set({ user, isAuthenticated: true });
    } catch {
      // Refresh failed (expired/revoked) — clear all auth state
      set({ user: null, accessToken: null, isAuthenticated: false });
      localStorage.removeItem('refresh_token');
    } finally {
      set({ isLoading: false });
    }
  },

  hasRole: (...roles: Role[]) => {
    const { user } = get();
    return user ? roles.includes(user.role) : false;
  },

  canDispatch: () => {
    const { user } = get();
    return user ? ['admin', 'dispatcher'].includes(user.role) : false;
  },

  isAdmin: () => {
    const { user } = get();
    return user?.role === 'admin';
  },
}));
