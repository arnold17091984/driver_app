import { create } from 'zustand';
import { createBooking } from '../api/bookings';
import type { VehicleETA, UnifiedBookingResponse } from '../types/api';

export type BookingStep = 'idle' | 'destination' | 'pickup-map' | 'vehicle-select' | 'status';

export interface BookingLocation {
  name: string;
  address: string;
  lat: number;
  lng: number;
}

interface BookingState {
  // Navigation
  step: BookingStep;
  setStep: (step: BookingStep) => void;

  // Mode
  isNow: boolean;
  setIsNow: (v: boolean) => void;
  scheduledStart: string | null;
  scheduledEnd: string | null;
  setScheduledTime: (start: string | null, end: string | null) => void;

  // Locations
  origin: BookingLocation | null;
  destination: BookingLocation | null;
  setOrigin: (loc: BookingLocation | null) => void;
  setDestination: (loc: BookingLocation | null) => void;

  // Pickup details
  pickupDetails: string;
  setPickupDetails: (v: string) => void;

  // Purpose / category
  purpose: string;
  setPurpose: (v: string) => void;
  purposeCategory: string | null;
  setPurposeCategory: (cat: string | null) => void;

  // Passenger info
  passengerName: string;
  setPassengerName: (v: string) => void;
  notes: string;
  setNotes: (v: string) => void;
  destinations: string[];
  setDestinations: (d: string[]) => void;

  // Vehicle selection
  availableVehicles: VehicleETA[];
  setAvailableVehicles: (v: VehicleETA[]) => void;
  selectedVehicleId: string | null;
  setSelectedVehicleId: (id: string | null) => void;

  // Submission
  isSubmitting: boolean;
  submitError: string;
  result: UnifiedBookingResponse | null;
  submit: () => Promise<void>;
  reset: () => void;
}

const INITIAL_STATE = {
  step: 'idle' as BookingStep,
  isNow: true,
  scheduledStart: null as string | null,
  scheduledEnd: null as string | null,
  origin: null as BookingLocation | null,
  destination: null as BookingLocation | null,
  pickupDetails: '',
  purpose: '',
  purposeCategory: null as string | null,
  passengerName: '',
  notes: '',
  destinations: [''] as string[],
  availableVehicles: [] as VehicleETA[],
  selectedVehicleId: null as string | null,
  isSubmitting: false,
  submitError: '',
  result: null as UnifiedBookingResponse | null,
};

export const useBookingStore = create<BookingState>((set, get) => ({
  ...INITIAL_STATE,

  setStep: (step) => set({ step }),
  setIsNow: (isNow) => set({ isNow }),
  setScheduledTime: (start, end) => set({ scheduledStart: start, scheduledEnd: end }),
  setOrigin: (origin) => set({ origin }),
  setDestination: (destination) => set({ destination }),
  setPickupDetails: (pickupDetails) => set({ pickupDetails }),
  setPurpose: (purpose) => set({ purpose }),
  setPurposeCategory: (purposeCategory) => set({ purposeCategory }),
  setPassengerName: (passengerName) => set({ passengerName }),
  setNotes: (notes) => set({ notes }),
  setDestinations: (destinations) => set({ destinations }),
  setAvailableVehicles: (availableVehicles) => set({ availableVehicles }),
  setSelectedVehicleId: (selectedVehicleId) => set({ selectedVehicleId }),

  submit: async () => {
    const s = get();
    set({ isSubmitting: true, submitError: '', step: 'status' });
    try {
      const pickupAddress = s.pickupDetails
        ? `${s.origin?.address || ''} (${s.pickupDetails})`
        : s.origin?.address || '';

      const resp = await createBooking({
        mode: s.selectedVehicleId ? 'specific' : 'any',
        vehicle_id: s.selectedVehicleId || undefined,
        is_now: s.isNow,
        start_time: !s.isNow && s.scheduledStart ? s.scheduledStart : undefined,
        end_time: !s.isNow && s.scheduledEnd ? s.scheduledEnd : undefined,
        pickup_address: pickupAddress || s.origin?.name || 'Pickup',
        pickup_lat: s.origin?.lat,
        pickup_lng: s.origin?.lng,
        purpose: s.purpose || s.purposeCategory || s.destination?.name || 'Transport',
        destinations: s.destinations.filter(d => d.trim()) || undefined,
        passenger_name: s.passengerName || undefined,
        notes: s.notes || undefined,
      });
      set({ result: resp, isSubmitting: false });
    } catch (err: unknown) {
      const msg = (err as { response?: { data?: { error?: { message?: string } } } })
        ?.response?.data?.error?.message || 'Failed to create booking';
      set({ submitError: msg, isSubmitting: false });
    }
  },

  reset: () => set({ ...INITIAL_STATE }),
}));
