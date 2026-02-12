import client from './client';
import type { Dispatch, ETASnapshot, VehicleETA } from '../types/api';

export async function listDispatches(status?: string): Promise<Dispatch[]> {
  const params = status ? { status } : {};
  const { data } = await client.get<Dispatch[]>('/dispatches', { params });
  return data;
}

export async function getDispatch(id: string): Promise<Dispatch> {
  const { data } = await client.get<Dispatch>(`/dispatches/${id}`);
  return data;
}

export async function createDispatch(req: {
  purpose: string;
  pickup_address: string;
  pickup_lat?: number;
  pickup_lng?: number;
  dropoff_address?: string;
  dropoff_lat?: number;
  dropoff_lng?: number;
  passenger_name?: string;
  passenger_count?: number;
  notes?: string;
}): Promise<Dispatch> {
  const { data } = await client.post<Dispatch>('/dispatches', req);
  return data;
}

export async function assignDispatch(id: string, vehicleId: string) {
  await client.post(`/dispatches/${id}/assign`, { vehicle_id: vehicleId });
}

export async function quickBoard(vehicleId: string, passengerName: string, purpose?: string, passengerCount?: number): Promise<Dispatch> {
  const { data } = await client.post<Dispatch>('/dispatches/quick-board', {
    vehicle_id: vehicleId,
    passenger_name: passengerName,
    purpose: purpose || '',
    passenger_count: passengerCount || 1,
  });
  return data;
}

export async function alightDispatch(id: string) {
  await client.post(`/dispatches/${id}/alight`);
}

export async function cancelDispatch(id: string, reason: string) {
  await client.post(`/dispatches/${id}/cancel`, { reason });
}

export async function getETASnapshots(id: string): Promise<ETASnapshot[]> {
  const { data } = await client.get<ETASnapshot[]>(`/dispatches/${id}/eta`);
  return data;
}

// Driver endpoints
export async function driverBoard(passengerName: string, purpose?: string, passengerCount?: number, estimatedMinutes?: number): Promise<Dispatch> {
  const { data } = await client.post<Dispatch>('/driver/board', {
    passenger_name: passengerName,
    purpose: purpose || '',
    passenger_count: passengerCount || 1,
    estimated_minutes: estimatedMinutes || 0,
  });
  return data;
}

export async function driverAlight(dispatchId: string) {
  await client.post(`/driver/trips/${dispatchId}/alight`);
}

export async function driverCurrentTrip(): Promise<Dispatch | null> {
  const { data } = await client.get<Dispatch | null>('/driver/trips/current');
  return data;
}

export async function calculateETAs(pickupLat: number, pickupLng: number): Promise<VehicleETA[]> {
  const { data } = await client.post<VehicleETA[]>('/dispatches/calculate-eta', {
    pickup_lat: pickupLat,
    pickup_lng: pickupLng,
  });
  return data;
}
