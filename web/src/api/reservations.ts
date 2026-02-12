import client from './client';
import type { Reservation, ReservationConflict, ConflictDetail } from '../types/api';

export async function listReservations(params?: {
  vehicle_id?: string;
  from?: string;
  to?: string;
  status?: string;
}): Promise<Reservation[]> {
  const { data } = await client.get<Reservation[]>('/reservations', { params });
  return data;
}

export async function getReservation(id: string): Promise<Reservation> {
  const { data } = await client.get<Reservation>(`/reservations/${id}`);
  return data;
}

export async function createReservation(req: {
  vehicle_id: string;
  start_time: string;
  end_time: string;
  purpose: string;
  destinations?: string[];
  notes?: string;
}): Promise<Reservation> {
  const { data } = await client.post<Reservation>('/reservations', req);
  return data;
}

export async function updateReservation(id: string, req: Partial<{
  vehicle_id: string;
  start_time: string;
  end_time: string;
  purpose: string;
  destinations: string[];
  notes: string;
}>): Promise<Reservation> {
  const { data } = await client.put<Reservation>(`/reservations/${id}`, req);
  return data;
}

export async function cancelReservation(id: string, reason: string) {
  await client.post(`/reservations/${id}/cancel`, { reason });
}

export async function listConflicts(): Promise<ReservationConflict[]> {
  const { data } = await client.get<ReservationConflict[]>('/conflicts');
  return data;
}

export async function getConflict(id: string): Promise<ConflictDetail> {
  const { data } = await client.get<ConflictDetail>(`/conflicts/${id}`);
  return data;
}

export async function resolveConflictReassign(id: string, newVehicleId: string, reason: string) {
  await client.post(`/conflicts/${id}/reassign`, { new_vehicle_id: newVehicleId, reason });
}

export async function resolveConflictChangeTime(id: string, newStartTime: string, newEndTime: string, reason: string) {
  await client.post(`/conflicts/${id}/change-time`, { new_start_time: newStartTime, new_end_time: newEndTime, reason });
}

export async function resolveConflictCancel(id: string, reason: string) {
  await client.post(`/conflicts/${id}/cancel`, { reason });
}

export async function forceAssign(id: string, reason: string) {
  await client.post(`/conflicts/${id}/force-assign`, { reason });
}
