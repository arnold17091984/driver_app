import { create } from 'zustand';
import client from '../services/apiClient';
import type { Dispatch } from '../types';

interface TripState {
  currentTrip: Dispatch | null;
  isLoading: boolean;
  fetchCurrentTrip: () => Promise<void>;
  acceptTrip: (tripId: string) => Promise<void>;
  startEnRoute: (tripId: string) => Promise<void>;
  markArrived: (tripId: string) => Promise<void>;
  completeTrip: (tripId: string) => Promise<void>;
}

export const useTripStore = create<TripState>((set) => ({
  currentTrip: null,
  isLoading: false,

  fetchCurrentTrip: async () => {
    set({ isLoading: true });
    try {
      const { data } = await client.get<Dispatch>('/driver/trips/current');
      set({ currentTrip: data });
    } catch {
      set({ currentTrip: null });
    } finally {
      set({ isLoading: false });
    }
  },

  acceptTrip: async (tripId: string) => {
    await client.post(`/driver/trips/${tripId}/accept`);
    const { data } = await client.get<Dispatch>('/driver/trips/current');
    set({ currentTrip: data });
  },

  startEnRoute: async (tripId: string) => {
    await client.post(`/driver/trips/${tripId}/en-route`);
    const { data } = await client.get<Dispatch>('/driver/trips/current');
    set({ currentTrip: data });
  },

  markArrived: async (tripId: string) => {
    await client.post(`/driver/trips/${tripId}/arrived`);
    const { data } = await client.get<Dispatch>('/driver/trips/current');
    set({ currentTrip: data });
  },

  completeTrip: async (tripId: string) => {
    await client.post(`/driver/trips/${tripId}/complete`);
    set({ currentTrip: null });
  },
}));
