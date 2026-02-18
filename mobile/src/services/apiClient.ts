import axios from 'axios';
import { Platform } from 'react-native';
import EncryptedStorage from 'react-native-encrypted-storage';

// In development, connect to local backend over HTTP (emulator/simulator).
// In production, always use HTTPS. Never allow cleartext traffic in release builds.
const DEFAULT_HOST = Platform.OS === 'android' ? '10.0.2.2' : 'localhost';
const DEFAULT_SCHEME = __DEV__ ? 'http' : 'https';
let apiBase = `${DEFAULT_SCHEME}://${DEFAULT_HOST}:8080`;

// Override for real device testing (dev only).
export function setApiBase(host: string) {
  const scheme = __DEV__ ? 'http' : 'https';
  apiBase = `${scheme}://${host}:8080`;
  client.defaults.baseURL = `${apiBase}/api/v1`;
}
export function getApiBase() {
  return apiBase;
}

const client = axios.create({
  baseURL: `${apiBase}/api/v1`,
  headers: { 'Content-Type': 'application/json' },
  timeout: 15000,
});

let accessToken: string | null = null;
let refreshToken: string | null = null;

export function setTokens(access: string, refresh: string) {
  accessToken = access;
  refreshToken = refresh;
}

export function clearTokens() {
  accessToken = null;
  refreshToken = null;
}

export function getAccessToken() {
  return accessToken;
}

client.interceptors.request.use((config) => {
  if (accessToken) {
    config.headers.Authorization = `Bearer ${accessToken}`;
  }
  return config;
});

client.interceptors.response.use(
  (response) => response,
  async (error) => {
    const originalRequest = error.config;

    if (error.response?.status === 401 && !originalRequest._retry && refreshToken) {
      originalRequest._retry = true;
      try {
        const { data } = await axios.post(`${apiBase}/api/v1/auth/refresh`, {
          refresh_token: refreshToken,
        });
        accessToken = data.access_token;
        // Persist the new access token to encrypted storage so it survives
        // app restarts and is not stored in plain-text (M22).
        await EncryptedStorage.setItem('access_token', accessToken as string);
        originalRequest.headers.Authorization = `Bearer ${accessToken}`;
        return client(originalRequest);
      } catch {
        clearTokens();
      }
    }

    return Promise.reject(error);
  }
);

export default client;
