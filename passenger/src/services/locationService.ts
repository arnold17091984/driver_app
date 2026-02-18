import {Platform, PermissionsAndroid} from 'react-native';
import Geolocation from '@react-native-community/geolocation';

interface Coords {
  latitude: number;
  longitude: number;
}

export async function requestLocationPermission(): Promise<boolean> {
  if (Platform.OS === 'ios') {
    return new Promise((resolve) => {
      Geolocation.requestAuthorization(
        () => resolve(true),
        () => resolve(false),
      );
    });
  }

  const granted = await PermissionsAndroid.request(
    PermissionsAndroid.PERMISSIONS.ACCESS_FINE_LOCATION,
    {
      title: 'Location Permission',
      message: 'This app needs access to your location for pickup.',
      buttonPositive: 'OK',
    },
  );
  return granted === PermissionsAndroid.RESULTS.GRANTED;
}

export function getCurrentPosition(): Promise<Coords> {
  return new Promise((resolve, reject) => {
    Geolocation.getCurrentPosition(
      (position) => {
        resolve({
          latitude: position.coords.latitude,
          longitude: position.coords.longitude,
        });
      },
      (error) => {
        console.warn('Geolocation error:', error);
        reject(new Error(`Geolocation error: ${error?.message || 'unknown'}`));
      },
      {enableHighAccuracy: true, timeout: 10000, maximumAge: 5000},
    );
  });
}
