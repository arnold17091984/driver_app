export interface User {
  id: string;
  employee_id: string;
  name: string;
  role: string;
  phone_number?: string;
}

export interface LoginResponse {
  access_token: string;
  refresh_token: string;
  user: User;
}

export interface Location {
  lat: number;
  lng: number;
  address: string;
}

export interface Dispatch {
  id: string;
  vehicle_id?: string;
  requester_id: string;
  dispatcher_id?: string;
  purpose: string;
  passenger_name?: string;
  passenger_count: number;
  notes?: string;
  pickup_address: string;
  pickup_lat?: number;
  pickup_lng?: number;
  dropoff_address?: string;
  dropoff_lat?: number;
  dropoff_lng?: number;
  status: DispatchStatus;
  estimated_duration_sec?: number;
  estimated_distance_m?: number;
  assigned_at?: string;
  accepted_at?: string;
  en_route_at?: string;
  arrived_at?: string;
  completed_at?: string;
  cancelled_at?: string;
  cancel_reason?: string;
  created_at: string;
  updated_at: string;
}

export type DispatchStatus =
  | 'pending'
  | 'assigned'
  | 'accepted'
  | 'en_route'
  | 'arrived'
  | 'completed'
  | 'cancelled';

export interface VehicleLocation {
  vehicle_id: string;
  latitude: number;
  longitude: number;
  speed?: number;
  heading?: number;
  recorded_at: string;
}

export interface BookingResponse {
  type: 'dispatch' | 'reservation';
  dispatch?: Dispatch;
  reservation?: unknown;
}
