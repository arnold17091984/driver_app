/**
 * Background Location Service
 *
 * Uses react-native-background-geolocation (Transistorsoft) for battery-efficient
 * background GPS tracking.
 *
 * SETUP REQUIRED:
 * 1. Install: yarn add react-native-background-geolocation
 * 2. Follow Transistorsoft setup guide for iOS and Android native configuration
 * 3. Set API_BASE to your server URL
 */

import { getAccessToken } from './apiClient';

// This is a placeholder implementation. In production, replace with:
// import BackgroundGeolocation from 'react-native-background-geolocation';

const API_BASE = 'http://localhost:8080';

export async function configureBackgroundLocation() {
  // BackgroundGeolocation.ready({
  //   desiredAccuracy: BackgroundGeolocation.DESIRED_ACCURACY_HIGH,
  //   distanceFilter: 50,
  //   stopTimeout: 5,
  //   stationaryRadius: 100,
  //   locationUpdateInterval: 30000,
  //   fastestLocationUpdateInterval: 15000,
  //   foregroundService: true,
  //   notification: {
  //     title: '車両位置送信中',
  //     text: '位置情報を送信しています',
  //   },
  //   activityType: BackgroundGeolocation.ACTIVITY_TYPE_AUTOMOTIVE_NAVIGATION,
  //   url: `${API_BASE}/api/v1/locations/report`,
  //   headers: { Authorization: `Bearer ${getAccessToken()}` },
  //   method: 'POST',
  //   autoSync: true,
  //   batchSync: true,
  //   maxBatchSize: 10,
  //   maxDaysToPersist: 1,
  // });
  console.log('Background location configured (placeholder)');
}

export function startTracking() {
  // BackgroundGeolocation.start();
  console.log('Location tracking started (placeholder)');
}

export function stopTracking() {
  // BackgroundGeolocation.stop();
  console.log('Location tracking stopped (placeholder)');
}

export function updateAuthToken() {
  // BackgroundGeolocation.setConfig({
  //   headers: { Authorization: `Bearer ${getAccessToken()}` },
  // });
}
