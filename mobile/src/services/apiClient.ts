import axios from 'axios';
import { Platform } from 'react-native';

// iOS Simulator uses localhost, Android emulator uses 10.0.2.2
// For real devices, set to your Mac's local IP address
const DEFAULT_HOST = Platform.OS === 'android' ? '10.0.2.2' : 'localhost';
let apiBase = `http://${DEFAULT_HOST}:8080`;

// Override for real device testing
export function setApiBase(host: string) {
  apiBase = `http://${host}:8080`;
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
