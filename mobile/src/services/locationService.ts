/**
 * Foreground Location Service
 *
 * Uses React Native's built-in Geolocation API for location tracking.
 * Sends location updates to the backend every 15 seconds while tracking is active.
 */

import { Platform, PermissionsAndroid } from 'react-native';
import Geolocation from '@react-native-community/geolocation';
import client, { getAccessToken } from './apiClient';

let watchId: number | null = null;
let intervalId: ReturnType<typeof setInterval> | null = null;
let lastPosition: { latitude: number; longitude: number; heading: number; speed: number; accuracy: number } | null = null;

async function requestPermission(): Promise<boolean> {
  if (Platform.OS === 'ios') {
    return new Promise((resolve) => {
      Geolocation.requestAuthorization(
        () => resolve(true),
        () => resolve(false),
      );
    });
  }
  // Android
  const granted = await PermissionsAndroid.request(
    PermissionsAndroid.PERMISSIONS.ACCESS_FINE_LOCATION,
  );
  return granted === PermissionsAndroid.RESULTS.GRANTED;
}

export async function configureBackgroundLocation() {
  Geolocation.setRNConfiguration({
    skipPermissionRequests: false,
    authorizationLevel: 'whenInUse',
    locationProvider: 'auto',
  });
  console.log('Location service configured');
}

export async function startTracking() {
  const hasPermission = await requestPermission();
  if (!hasPermission) {
    console.warn('Location permission denied');
    return;
  }

  // Watch position changes
  watchId = Geolocation.watchPosition(
    (position) => {
      lastPosition = {
        latitude: position.coords.latitude,
        longitude: position.coords.longitude,
        heading: position.coords.heading ?? 0,
        speed: position.coords.speed ?? 0,
        accuracy: position.coords.accuracy,
      };
    },
    (error) => console.warn('Location watch error:', error.message),
    { enableHighAccuracy: true, distanceFilter: 20 },
  );

  // Send location to backend every 15 seconds
  intervalId = setInterval(() => {
    if (lastPosition && getAccessToken()) {
      sendLocation(lastPosition).catch(() => {});
    }
  }, 15000);

  console.log('Location tracking started');
}

export function stopTracking() {
  if (watchId !== null) {
    Geolocation.clearWatch(watchId);
    watchId = null;
  }
  if (intervalId !== null) {
    clearInterval(intervalId);
    intervalId = null;
  }
  lastPosition = null;
  console.log('Location tracking stopped');
}

export function updateAuthToken() {
  // Token is read dynamically from getAccessToken(), no action needed
}

async function sendLocation(pos: NonNullable<typeof lastPosition>) {
  await client.post('/locations/report', {
    points: [{
      latitude: pos.latitude,
      longitude: pos.longitude,
      heading: pos.heading,
      speed: pos.speed,
      accuracy: pos.accuracy,
      recorded_at: new Date().toISOString(),
    }],
  });
}
