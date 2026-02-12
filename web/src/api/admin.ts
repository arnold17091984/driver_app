import client from './client';
import type { User, AuditLog } from '../types/api';

export async function listUsers(): Promise<User[]> {
  const { data } = await client.get<User[]>('/admin/users');
  return data;
}

export async function updateUserRole(id: string, role: string) {
  await client.put(`/admin/users/${id}/role`, { role });
}

export async function updateUserPriority(id: string, priorityLevel: number) {
  await client.put(`/admin/users/${id}/priority`, { priority_level: priorityLevel });
}

export async function listAuditLogs(params?: {
  actor_id?: string;
  action?: string;
  target_type?: string;
  from?: string;
  to?: string;
  limit?: number;
  offset?: number;
}): Promise<AuditLog[]> {
  const { data } = await client.get<AuditLog[]>('/admin/audit-logs', { params });
  return data;
}

export async function getAuditLog(id: string): Promise<AuditLog> {
  const { data } = await client.get<AuditLog>(`/admin/audit-logs/${id}`);
  return data;
}
