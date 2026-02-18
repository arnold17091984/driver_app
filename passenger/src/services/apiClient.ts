import axios from 'axios';
import {Platform} from 'react-native';
import EncryptedStorage from 'react-native-encrypted-storage';

// In development, connect to local backend over HTTP (emulator/simulator).
// In production, always use HTTPS. Never allow cleartext traffic in release builds.
const DEFAULT_HOST = __DEV__
  ? Platform.OS === 'android'
    ? '10.0.2.2'
    : 'localhost'
  : 'localhost';

const DEFAULT_SCHEME = __DEV__ ? 'http' : 'https';
let apiBase = `${DEFAULT_SCHEME}://${DEFAULT_HOST}:8080/api/v1`;

export function setApiBase(host: string) {
  const scheme = __DEV__ ? 'http' : 'https';
  apiBase = `${scheme}://${host}:8080/api/v1`;
  client.defaults.baseURL = apiBase;
}

export function getApiBase() {
  return apiBase;
}

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

export const client = axios.create({
  baseURL: apiBase,
  timeout: 15000,
  headers: {'Content-Type': 'application/json'},
});

// Attach token to every request
client.interceptors.request.use(config => {
  if (accessToken) {
    config.headers.Authorization = `Bearer ${accessToken}`;
  }
  return config;
});

// Auto-refresh on 401
client.interceptors.response.use(
  response => response,
  async error => {
    const original = error.config;
    if (
      error.response?.status === 401 &&
      !original._retry &&
      refreshToken
    ) {
      original._retry = true;
      try {
        const res = await axios.post(`${apiBase}/auth/refresh`, {
          refresh_token: refreshToken,
        });
        const newToken = res.data.access_token;
        accessToken = newToken;
        // Persist the new access token to encrypted storage so it survives
        // app restarts and is not stored in plain-text (M22).
        await EncryptedStorage.setItem('access_token', newToken as string);
        original.headers.Authorization = `Bearer ${newToken}`;
        return client(original);
      } catch {
        clearTokens();
      }
    }
    return Promise.reject(error);
  },
);
