import React, {useEffect} from 'react';
import {useAuthStore} from './stores/authStore';
import {requestLocationPermission} from './services/locationService';
import RootNavigator from './navigation/RootNavigator';

export default function App() {
  const restoreSession = useAuthStore(s => s.restoreSession);

  useEffect(() => {
    restoreSession();
    requestLocationPermission();
  }, [restoreSession]);

  return <RootNavigator />;
}
