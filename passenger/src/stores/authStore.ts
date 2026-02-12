import {create} from 'zustand';
import EncryptedStorage from 'react-native-encrypted-storage';
import {client, setTokens, clearTokens} from '../services/apiClient';
import type {User, LoginResponse} from '../types';

interface AuthState {
  user: User | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  login: (phoneNumber: string, password: string) => Promise<void>;
  register: (
    phoneNumber: string,
    password: string,
    name: string,
  ) => Promise<void>;
  logout: () => Promise<void>;
  restoreSession: () => Promise<void>;
}

export const useAuthStore = create<AuthState>((set) => ({
  user: null,
  isAuthenticated: false,
  isLoading: true,

  login: async (phoneNumber, password) => {
    const res = await client.post<LoginResponse>('/auth/passenger/login', {
      phone_number: phoneNumber,
      password,
    });
    const data = res.data;
    setTokens(data.access_token, data.refresh_token);
    await EncryptedStorage.setItem('access_token', data.access_token);
    await EncryptedStorage.setItem('refresh_token', data.refresh_token);
    set({user: data.user, isAuthenticated: true});
  },

  register: async (phoneNumber, password, name) => {
    const res = await client.post<LoginResponse>(
      '/auth/passenger/register',
      {
        phone_number: phoneNumber,
        password,
        name,
      },
    );
    const data = res.data;
    setTokens(data.access_token, data.refresh_token);
    await EncryptedStorage.setItem('access_token', data.access_token);
    await EncryptedStorage.setItem('refresh_token', data.refresh_token);
    set({user: data.user, isAuthenticated: true});
  },

  logout: async () => {
    clearTokens();
    await EncryptedStorage.removeItem('access_token');
    await EncryptedStorage.removeItem('refresh_token');
    set({user: null, isAuthenticated: false});
  },

  restoreSession: async () => {
    try {
      const accessToken = await EncryptedStorage.getItem('access_token');
      const refreshToken = await EncryptedStorage.getItem('refresh_token');
      if (!accessToken || !refreshToken) {
        set({isLoading: false});
        return;
      }
      setTokens(accessToken, refreshToken);
      const res = await client.get('/auth/me');
      set({user: res.data, isAuthenticated: true, isLoading: false});
    } catch {
      clearTokens();
      await EncryptedStorage.removeItem('access_token');
      await EncryptedStorage.removeItem('refresh_token');
      set({isLoading: false});
    }
  },
}));
