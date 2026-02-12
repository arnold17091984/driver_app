export const ROLES = {
  ADMIN: 'admin',
  DISPATCHER: 'dispatcher',
  VIEWER: 'viewer',
  DRIVER: 'driver',
} as const;

export const VEHICLE_STATUSES = {
  AVAILABLE: 'available',
  DRIVER_ABSENT: 'driver_absent',
  RESERVED: 'reserved',
  IN_TRIP: 'in_trip',
  MAINTENANCE: 'maintenance',
  STALE_LOCATION: 'stale_location',
} as const;

export const DISPATCH_STATUSES = {
  PENDING: 'pending',
  ASSIGNED: 'assigned',
  ACCEPTED: 'accepted',
  EN_ROUTE: 'en_route',
  ARRIVED: 'arrived',
  COMPLETED: 'completed',
  CANCELLED: 'cancelled',
} as const;

export const RESERVATION_STATUSES = {
  CONFIRMED: 'confirmed',
  PENDING_CONFLICT: 'pending_conflict',
  CANCELLED: 'cancelled',
  COMPLETED: 'completed',
} as const;

export const PERMISSIONS = {
  P1_VIEW_VEHICLES: 'P1',
  P2_CREATE_RESERVATION: 'P2',
  P3_EDIT_RESERVATION: 'P3',
  P4_CREATE_DISPATCH: 'P4',
  P5_ASSIGN_DISPATCH: 'P5',
  P6_CANCEL_DISPATCH: 'P6',
  P7_RESOLVE_CONFLICT: 'P7',
  P8_FORCE_ASSIGN: 'P8',
  P9_MANAGE_ROLES: 'P9',
  P10_TOGGLE_MAINTENANCE: 'P10',
  P11_VIEW_AUDIT: 'P11',
} as const;

export const ROLE_PERMISSIONS: Record<string, string[]> = {
  admin: Object.values(PERMISSIONS),
  dispatcher: ['P1', 'P2', 'P3', 'P4', 'P5', 'P6', 'P7', 'P10'],
  viewer: ['P1'],
  driver: ['P1'],
};

export const POLLING_INTERVALS = {
  VEHICLE_POSITIONS: 10000, // 10 seconds
  DISPATCHES: 15000,        // 15 seconds
  CONFLICTS: 30000,         // 30 seconds
} as const;
