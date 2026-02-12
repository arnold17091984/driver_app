import { create } from 'zustand';
import EncryptedStorage from 'react-native-encrypted-storage';
import client, { setTokens, clearTokens } from '../services/apiClient';
import { setupNotifications } from '../services/notificationService';
import type { User, LoginResponse } from '../types';

interface AuthState {
  user: User | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  login: (employeeId: string, password: string) => Promise<void>;
  logout: () => void;
  restoreSession: () => Promise<void>;
  _restoreFromTokens: (accessToken: string, refreshToken: string) => Promise<void>;
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

      // Store tokens securely
      await EncryptedStorage.setItem('access_token', data.access_token);
      await EncryptedStorage.setItem('refresh_token', data.refresh_token);

      setupNotifications();
    } finally {
      set({ isLoading: false });
    }
  },

  logout: () => {
    clearTokens();
    set({ user: null, isAuthenticated: false });
    EncryptedStorage.removeItem('access_token').catch(() => {});
    EncryptedStorage.removeItem('refresh_token').catch(() => {});
  },

  // Called on app start to restore session from encrypted storage
  restoreSession: async () => {
    set({ isLoading: true });
    try {
      const accessToken = await EncryptedStorage.getItem('access_token');
      const refreshToken = await EncryptedStorage.getItem('refresh_token');
      if (!accessToken || !refreshToken) {
        set({ isLoading: false });
        return;
      }
      setTokens(accessToken, refreshToken);
      const { data } = await client.get<User>('/auth/me');
      set({ user: data, isAuthenticated: true });
      setupNotifications();
    } catch {
      clearTokens();
      await EncryptedStorage.removeItem('access_token').catch(() => {});
      await EncryptedStorage.removeItem('refresh_token').catch(() => {});
      set({ user: null, isAuthenticated: false });
    } finally {
      set({ isLoading: false });
    }
  },

  // For manual token restore (e.g., deep linking)
  _restoreFromTokens: async (accessToken: string, refreshToken: string) => {
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
