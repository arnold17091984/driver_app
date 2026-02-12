import { create } from 'zustand';
import type { User, Role } from '../types/api';
import { login as apiLogin, getMe } from '../api/auth';

interface AuthState {
  user: User | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  login: (employeeId: string, password: string) => Promise<void>;
  logout: () => void;
  loadUser: () => Promise<void>;
  hasRole: (...roles: Role[]) => boolean;
  canDispatch: () => boolean;
  isAdmin: () => boolean;
}

export const useAuthStore = create<AuthState>((set, get) => ({
  user: null,
  isAuthenticated: !!localStorage.getItem('access_token'),
  isLoading: false,

  login: async (employeeId: string, password: string) => {
    set({ isLoading: true });
    try {
      const resp = await apiLogin(employeeId, password);
      localStorage.setItem('access_token', resp.access_token);
      localStorage.setItem('refresh_token', resp.refresh_token);
      set({ user: resp.user, isAuthenticated: true });
    } finally {
      set({ isLoading: false });
    }
  },

  logout: () => {
    localStorage.removeItem('access_token');
    localStorage.removeItem('refresh_token');
    set({ user: null, isAuthenticated: false });
  },

  loadUser: async () => {
    if (!localStorage.getItem('access_token')) return;
    set({ isLoading: true });
    try {
      const user = await getMe();
      set({ user, isAuthenticated: true });
    } catch {
      set({ user: null, isAuthenticated: false });
      localStorage.removeItem('access_token');
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
