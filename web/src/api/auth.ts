import client from './client';
import type { LoginResponse, User } from '../types/api';

export async function login(employeeId: string, password: string): Promise<LoginResponse> {
  const { data } = await client.post<LoginResponse>('/auth/login', {
    employee_id: employeeId,
    password,
  });
  return data;
}

export async function refreshToken(token: string) {
  const { data } = await client.post('/auth/refresh', { refresh_token: token });
  return data;
}

export async function getMe(): Promise<User> {
  const { data } = await client.get<User>('/auth/me');
  return data;
}
