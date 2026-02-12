export type Role = 'admin' | 'dispatcher' | 'viewer' | 'driver';

export type VehicleStatus =
  | 'available'
  | 'waiting'
  | 'driver_absent'
  | 'reserved'
  | 'in_trip'
  | 'maintenance'
  | 'stale_location';

export type DispatchStatus =
  | 'pending'
  | 'assigned'
  | 'accepted'
  | 'en_route'
  | 'arrived'
  | 'completed'
  | 'cancelled';

export type ReservationStatus =
  | 'confirmed'
  | 'pending_conflict'
  | 'pending_driver'
  | 'driver_declined'
  | 'cancelled'
  | 'completed';

export type ConflictStatus =
  | 'pending'
  | 'resolved_reassign'
  | 'resolved_changed'
  | 'resolved_cancelled'
  | 'force_assigned';

export interface User {
  id: string;
  employee_id: string;
  name: string;
  role: Role;
  priority_level: number;
  is_active: boolean;
}

export interface LoginResponse {
  access_token: string;
  refresh_token: string;
  user: User;
}

export interface Vehicle {
  id: string;
  name: string;
  license_plate: string;
  driver_id: string;
  driver_name: string;
  is_maintenance: boolean;
  is_clocked_in: boolean;
  photo_url?: string;
  status: VehicleStatus;
  latitude?: number;
  longitude?: number;
  heading?: number;
  speed?: number;
  location_at?: string;
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
  estimated_end_at?: string;
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

export interface ETASnapshot {
  id: string;
  dispatch_id: string;
  vehicle_id: string;
  vehicle_name: string;
  duration_sec: number;
  distance_m: number;
  calculated_at: string;
}

export interface Reservation {
  id: string;
  vehicle_id: string;
  requester_id: string;
  start_time: string;
  end_time: string;
  purpose: string;
  destinations?: string[];
  notes?: string;
  passenger_name?: string;
  pickup_address?: string;
  pickup_lat?: number;
  pickup_lng?: number;
  declined_by_driver_ids?: string[];
  priority_level: number;
  status: ReservationStatus;
  cancel_reason?: string;
  cancelled_by?: string;
  vehicle_name?: string;
  requester_name?: string;
  created_at: string;
  updated_at: string;
}

export type BookingMode = 'specific' | 'any';

export interface UnifiedBookingRequest {
  mode: BookingMode;
  vehicle_id?: string;
  is_now: boolean;
  start_time?: string;
  end_time?: string;
  pickup_address: string;
  pickup_lat?: number;
  pickup_lng?: number;
  purpose: string;
  destinations?: string[];
  passenger_name?: string;
  notes?: string;
}

export interface UnifiedBookingResponse {
  type: 'dispatch' | 'reservation';
  dispatch?: Dispatch;
  reservation?: Reservation;
}

export interface ReservationConflict {
  id: string;
  winning_reservation_id: string;
  losing_reservation_id: string;
  status: ConflictStatus;
  resolved_by?: string;
  resolution_reason?: string;
  resolved_at?: string;
  created_at: string;
}

export interface ConflictDetail {
  conflict: ReservationConflict;
  winning_reservation: Reservation;
  losing_reservation: Reservation;
}

export interface AuditLog {
  id: string;
  actor_id: string;
  actor_name: string;
  action: string;
  target_type: string;
  target_id: string;
  before_state?: unknown;
  after_state?: unknown;
  reason?: string;
  created_at: string;
}

export interface VehicleETA {
  vehicle_id: string;
  vehicle_name: string;
  driver_name: string;
  plate: string;
  status: string;
  latitude: number;
  longitude: number;
  distance_m: number;
  duration_sec: number;
  is_available: boolean;
}

export interface ApiError {
  error: {
    code: string;
    message: string;
  };
}
