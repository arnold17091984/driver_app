import axios from 'axios';
import { useAuthStore } from '../stores/authStore';

const API_BASE = import.meta.env.VITE_API_BASE || '';

const client = axios.create({
  baseURL: `${API_BASE}/api/v1`,
  headers: { 'Content-Type': 'application/json' },
});

// Read the access token from Zustand memory store — never from localStorage
client.interceptors.request.use((config) => {
  const token = useAuthStore.getState().accessToken;
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

client.interceptors.response.use(
  (response) => response,
  async (error) => {
    const originalRequest = error.config;

    if (error.response?.status === 401 && !originalRequest._retry) {
      originalRequest._retry = true;
      // TODO: When refresh_token moves to httpOnly cookie this explicit read can be removed —
      // the browser will send the cookie automatically on the refresh request.
      const refreshToken = localStorage.getItem('refresh_token');

      if (refreshToken) {
        try {
          const { data } = await axios.post(`${API_BASE}/api/v1/auth/refresh`, {
            refresh_token: refreshToken,
          });

          // Store the new access token in Zustand memory only — never in localStorage
          useAuthStore.getState().setAccessToken(data.access_token);

          // Update stored refresh token if the server rotated it
          if (data.refresh_token) {
            // TODO: Remove once server sets httpOnly cookie on refresh
            localStorage.setItem('refresh_token', data.refresh_token);
          }

          originalRequest.headers.Authorization = `Bearer ${data.access_token}`;
          return client(originalRequest);
        } catch {
          // Refresh failed — clear all auth state and redirect to login
          useAuthStore.getState().setAccessToken(null);
          localStorage.removeItem('refresh_token');
          window.location.href = '/login';
        }
      }
    }

    return Promise.reject(error);
  }
);

export default client;
