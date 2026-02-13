import {create} from 'zustand';
import {client} from '../services/apiClient';
import type {Dispatch, VehicleLocation, BookingResponse} from '../types';

interface RideState {
  currentRide: Dispatch | null;
  rideHistory: Dispatch[];
  driverLocation: VehicleLocation | null;
  driverLocationStale: boolean;
  isRequesting: boolean;
  isLoadingHistory: boolean;
  pollingInterval: ReturnType<typeof setInterval> | null;
  _pollFailureCount: number;

  requestRide: (
    pickupAddress: string,
    pickupLat: number,
    pickupLng: number,
    dropoffAddress?: string,
    dropoffLat?: number,
    dropoffLng?: number,
    passengerName?: string,
  ) => Promise<void>;
  cancelRide: (rideId: string) => Promise<void>;
  fetchCurrentRide: () => Promise<void>;
  fetchHistory: () => Promise<void>;
  rateRide: (rideId: string, rating: number, comment?: string) => Promise<void>;
  pollDriverLocation: (rideId: string) => void;
  stopPolling: () => void;
  clearCurrentRide: () => void;
}

export const useRideStore = create<RideState>((set, get) => ({
  currentRide: null,
  rideHistory: [],
  driverLocation: null,
  driverLocationStale: false,
  isRequesting: false,
  isLoadingHistory: false,
  pollingInterval: null,
  _pollFailureCount: 0,

  requestRide: async (
    pickupAddress,
    pickupLat,
    pickupLng,
    dropoffAddress,
    dropoffLat,
    dropoffLng,
    passengerName,
  ) => {
    set({isRequesting: true});
    try {
      const res = await client.post<BookingResponse>('/passenger/rides', {
        pickup_address: pickupAddress,
        pickup_lat: pickupLat,
        pickup_lng: pickupLng,
        dropoff_address: dropoffAddress || '',
        dropoff_lat: dropoffLat,
        dropoff_lng: dropoffLng,
        passenger_name: passengerName || '',
        passenger_count: 1,
      });
      if (res.data.dispatch) {
        set({currentRide: res.data.dispatch});
      }
    } finally {
      set({isRequesting: false});
    }
  },

  cancelRide: async (rideId) => {
    await client.post(`/passenger/rides/${rideId}/cancel`);
    set({currentRide: null, driverLocation: null});
    get().stopPolling();
  },

  fetchCurrentRide: async () => {
    try {
      const res = await client.get('/passenger/rides/current');
      set({currentRide: res.data || null});
    } catch {
      set({currentRide: null});
    }
  },

  fetchHistory: async () => {
    set({isLoadingHistory: true});
    try {
      const res = await client.get('/passenger/rides/history?limit=50');
      set({rideHistory: res.data || []});
    } catch {
      set({rideHistory: []});
    } finally {
      set({isLoadingHistory: false});
    }
  },

  rateRide: async (rideId, rating, comment) => {
    await client.post(`/passenger/rides/${rideId}/rate`, {rating, comment});
  },

  pollDriverLocation: (rideId) => {
    // Stop any existing polling
    get().stopPolling();

    // Reset failure counter when starting new polling
    set({_pollFailureCount: 0, driverLocationStale: false});

    const fetchLocation = async () => {
      try {
        const res = await client.get(
          `/passenger/rides/${rideId}/driver-location`,
        );
        if (res.data) {
          set({driverLocation: res.data, _pollFailureCount: 0, driverLocationStale: false});
        }
      } catch {
        const count = get()._pollFailureCount + 1;
        set({
          _pollFailureCount: count,
          driverLocationStale: count >= 3,
        });
      }
    };

    // Fetch immediately, then every 10 seconds
    fetchLocation();
    const interval = setInterval(fetchLocation, 10000);
    set({pollingInterval: interval});
  },

  stopPolling: () => {
    const {pollingInterval} = get();
    if (pollingInterval) {
      clearInterval(pollingInterval);
      set({pollingInterval: null});
    }
  },

  clearCurrentRide: () => {
    get().stopPolling();
    set({currentRide: null, driverLocation: null, driverLocationStale: false, _pollFailureCount: 0});
  },
}));
