import React, { useEffect } from 'react';
import { RootNavigator } from './navigation/RootNavigator';
import { configureBackgroundLocation } from './services/locationService';
import { setupNotifications } from './services/notificationService';

export default function App() {
  useEffect(() => {
    configureBackgroundLocation();
    setupNotifications();
  }, []);

  return <RootNavigator />;
}
