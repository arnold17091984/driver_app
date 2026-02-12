import client from './client';
import type { UnifiedBookingRequest, UnifiedBookingResponse, Reservation } from '../types/api';

export async function createBooking(req: UnifiedBookingRequest): Promise<UnifiedBookingResponse> {
  const { data } = await client.post<UnifiedBookingResponse>('/bookings', req);
  return data;
}

export async function getVehicleTimeline(vehicleId: string, date: string): Promise<Reservation[]> {
  const { data } = await client.get<Reservation[]>(`/vehicles/${vehicleId}/timeline`, {
    params: { date },
  });
  return data;
}

export async function driverPendingReservations(): Promise<Reservation[]> {
  const { data } = await client.get<Reservation[]>('/driver/reservations/pending');
  return data;
}

export async function driverAcceptReservation(id: string): Promise<void> {
  await client.post(`/driver/reservations/${id}/accept`);
}

export async function driverDeclineReservation(id: string, reason: string): Promise<void> {
  await client.post(`/driver/reservations/${id}/decline`, { reason });
}
