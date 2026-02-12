import { create } from 'zustand';
import client, { setTokens, clearTokens } from '../services/apiClient';
import { setupNotifications } from '../services/notificationService';
import type { User, LoginResponse } from '../types';

interface AuthState {
  user: User | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  login: (employeeId: string, password: string) => Promise<void>;
  logout: () => void;
  restoreSession: (accessToken: string, refreshToken: string) => Promise<void>;
}

export const useAuthStore = create<AuthState>((set) => ({
  user: null,
  isAuthenticated: false,
  isLoading: false,

  login: async (employeeId: string, password: string) => {
    set({ isLoading: true });
    try {
      const { data } = await client.post<LoginResponse>('/auth/login', {
        employee_id: employeeId,
        password,
      });
      setTokens(data.access_token, data.refresh_token);
      set({ user: data.user, isAuthenticated: true });
      setupNotifications();

      // Store tokens securely
      // await EncryptedStorage.setItem('access_token', data.access_token);
      // await EncryptedStorage.setItem('refresh_token', data.refresh_token);
    } finally {
      set({ isLoading: false });
    }
  },

  logout: () => {
    clearTokens();
    set({ user: null, isAuthenticated: false });
    // EncryptedStorage.removeItem('access_token');
    // EncryptedStorage.removeItem('refresh_token');
  },

  restoreSession: async (accessToken: string, refreshToken: string) => {
    setTokens(accessToken, refreshToken);
    set({ isLoading: true });
    try {
      const { data } = await client.get<User>('/auth/me');
      set({ user: data, isAuthenticated: true });
      setupNotifications();
    } catch {
      clearTokens();
      set({ user: null, isAuthenticated: false });
    } finally {
      set({ isLoading: false });
    }
  },
}));
