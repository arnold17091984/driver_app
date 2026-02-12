import client from './client';
import type { Vehicle } from '../types/api';

export async function listVehicles(): Promise<Vehicle[]> {
  const { data } = await client.get<Vehicle[]>('/vehicles');
  return data;
}

export async function listAvailableVehicles(): Promise<Vehicle[]> {
  const { data } = await client.get<Vehicle[]>('/vehicles/available');
  return data;
}

export async function createVehicle(req: { name: string; license_plate: string; driver_id: string }) {
  const { data } = await client.post('/vehicles', req);
  return data;
}

export async function updateVehicle(id: string, req: { name: string; license_plate: string; driver_id: string }) {
  await client.put(`/vehicles/${id}`, req);
}

export async function deleteVehicle(id: string) {
  await client.delete(`/vehicles/${id}`);
}

export async function uploadVehiclePhoto(id: string, file: File): Promise<{ photo_url: string }> {
  const formData = new FormData();
  formData.append('photo', file);
  const { data } = await client.post(`/vehicles/${id}/photo`, formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
  });
  return data;
}

export async function toggleMaintenance(id: string, isMaintenance: boolean) {
  await client.patch(`/vehicles/${id}/maintenance`, { is_maintenance: isMaintenance });
}
