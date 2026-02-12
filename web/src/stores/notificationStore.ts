import { create } from 'zustand';

interface NotificationState {
  soundEnabled: boolean;
  toggleSound: () => void;
}

export const useNotificationStore = create<NotificationState>((set) => ({
  soundEnabled: localStorage.getItem('notification_sound') !== 'off',

  toggleSound: () =>
    set((state) => {
      const newVal = !state.soundEnabled;
      localStorage.setItem('notification_sound', newVal ? 'on' : 'off');
      return { soundEnabled: newVal };
    }),
}));
