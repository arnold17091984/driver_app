import React, { useEffect } from 'react';
import { RootNavigator } from './navigation/RootNavigator';
import { useAuthStore } from './stores/authStore';
import { configureBackgroundLocation } from './services/locationService';

export default function App() {
  const restoreSession = useAuthStore((s) => s.restoreSession);

  useEffect(() => {
    // Restore saved session on app start
    restoreSession();
    configureBackgroundLocation();
  }, [restoreSession]);

  return <RootNavigator />;
}
