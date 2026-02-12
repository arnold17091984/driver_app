import {Platform, PermissionsAndroid} from 'react-native';

interface Coords {
  latitude: number;
  longitude: number;
}

export async function requestLocationPermission(): Promise<boolean> {
  if (Platform.OS === 'ios') {
    // iOS permissions are handled via Info.plist + native prompt
    return true;
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
    // Use the Geolocation API from react-native
    const {Geolocation} =
      require('react-native') as typeof import('react-native') & {
        Geolocation: any;
      };

    // Fallback: use navigator.geolocation if available
    const geo = Geolocation || (global as any).navigator?.geolocation;
    if (!geo) {
      // Simulator fallback: Tokyo Station
      resolve({latitude: 35.6812, longitude: 139.7671});
      return;
    }

    geo.getCurrentPosition(
      (position: any) => {
        resolve({
          latitude: position.coords.latitude,
          longitude: position.coords.longitude,
        });
      },
      (error: any) => {
        // Fallback to Tokyo Station on error
        console.warn('Geolocation error, using fallback:', error);
        resolve({latitude: 35.6812, longitude: 139.7671});
      },
      {enableHighAccuracy: true, timeout: 10000, maximumAge: 5000},
    );
  });
}
