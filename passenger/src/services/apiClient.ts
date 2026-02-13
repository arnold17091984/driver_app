import axios from 'axios';
import {Platform} from 'react-native';

const DEFAULT_HOST = __DEV__
  ? Platform.OS === 'android'
    ? '10.0.2.2'
    : 'localhost'
  : 'localhost';

let apiBase = `http://${DEFAULT_HOST}:8080/api/v1`;

export function setApiBase(host: string) {
  apiBase = `http://${host}:8080/api/v1`;
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
        original.headers.Authorization = `Bearer ${newToken}`;
        return client(original);
      } catch {
        clearTokens();
      }
    }
    return Promise.reject(error);
  },
);
