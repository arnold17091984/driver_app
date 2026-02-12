import { useRef, useCallback } from 'react';
import { playNotificationSound } from '../utils/notificationSound';
import { useNotificationStore } from '../stores/notificationStore';

type SoundType = 'info' | 'dispatch' | 'urgent' | 'success';

/**
 * Detects NEW items in polled data by comparing IDs across renders.
 * First call records existing IDs without playing sounds.
 * Subsequent calls play a sound when new IDs appear.
 */
export function useNotificationSound<T extends { id: string }>(soundType: SoundType = 'info') {
  const prevIdsRef = useRef<Set<string> | null>(null);

  const checkForNew = useCallback(
    (items: T[]) => {
      const currentIds = new Set(items.map((item) => item.id));

      if (prevIdsRef.current === null) {
        prevIdsRef.current = currentIds;
        return;
      }

      const { soundEnabled } = useNotificationStore.getState();
      if (!soundEnabled) {
        prevIdsRef.current = currentIds;
        return;
      }

      let hasNew = false;
      for (const id of currentIds) {
        if (!prevIdsRef.current.has(id)) {
          hasNew = true;
          break;
        }
      }

      if (hasNew) {
        playNotificationSound(soundType);
      }

      prevIdsRef.current = currentIds;
    },
    [soundType],
  );

  return { checkForNew };
}
