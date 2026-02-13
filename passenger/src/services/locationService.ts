import {Platform, PermissionsAndroid} from 'react-native';

interface Coords {
  latitude: number;
  longitude: number;
}

// Manila default coordinates for development fallback
const MANILA_COORDS: Coords = {latitude: 14.5547, longitude: 121.0244};

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
      if (__DEV__) {
        // Development fallback: Manila
        resolve(MANILA_COORDS);
      } else {
        reject(new Error('Geolocation not available'));
      }
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
        console.warn('Geolocation error:', error);
        if (__DEV__) {
          // Development fallback: Manila
          resolve(MANILA_COORDS);
        } else {
          reject(new Error(`Geolocation error: ${error?.message || 'unknown'}`));
        }
      },
      {enableHighAccuracy: true, timeout: 10000, maximumAge: 5000},
    );
  });
}
