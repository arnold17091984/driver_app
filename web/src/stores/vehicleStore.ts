import { create } from 'zustand';
import type { Vehicle } from '../types/api';
import { listVehicles } from '../api/vehicles';

interface VehicleState {
  vehicles: Vehicle[];
  selectedVehicleId: string | null;
  isLoading: boolean;
  fetchVehicles: () => Promise<void>;
  selectVehicle: (id: string | null) => void;
}

export const useVehicleStore = create<VehicleState>((set) => ({
  vehicles: [],
  selectedVehicleId: null,
  isLoading: false,

  fetchVehicles: async () => {
    set({ isLoading: true });
    try {
      const vehicles = await listVehicles();
      set({ vehicles });
    } finally {
      set({ isLoading: false });
    }
  },

  selectVehicle: (id) => set({ selectedVehicleId: id }),
}));
