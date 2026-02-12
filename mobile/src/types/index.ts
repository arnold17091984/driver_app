export type DispatchStatus =
  | 'pending'
  | 'assigned'
  | 'accepted'
  | 'en_route'
  | 'arrived'
  | 'completed'
  | 'cancelled';

export interface User {
  id: string;
  employee_id: string;
  name: string;
  role: string;
  priority_level: number;
}

export interface LoginResponse {
  access_token: string;
  refresh_token: string;
  user: User;
}

export interface AttendanceStatus {
  clocked_in: boolean;
  attendance: {
    id: string;
    driver_id: string;
    clock_in_at: string;
    clock_out_at?: string;
  } | null;
}

export interface Dispatch {
  id: string;
  vehicle_id?: string;
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
  created_at: string;
}

export interface LocationPoint {
  latitude: number;
  longitude: number;
  heading?: number;
  speed?: number;
  accuracy?: number;
  recorded_at: string;
}

export type RootStackParamList = {
  Login: undefined;
  Home: undefined;
  TripDetail: { tripId: string };
  TripActive: { tripId: string };
};
