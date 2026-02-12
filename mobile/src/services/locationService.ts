/**
 * Location Service with Background Support
 *
 * iOS: Uses Background Mode "location" + Always authorization
 * Android: Uses react-native-background-actions for foreground service
 *
 * Sends location updates to the backend every 15 seconds while tracking is active.
 */

import { Platform, PermissionsAndroid } from 'react-native';
import Geolocation from '@react-native-community/geolocation';
import client, { getAccessToken } from './apiClient';

// Only import on Android - iOS handles background location natively
const BackgroundService = Platform.OS === 'android'
  ? require('react-native-background-actions').default
  : null;

let watchId: number | null = null;
let lastPosition: {
  latitude: number;
  longitude: number;
  heading: number;
  speed: number;
  accuracy: number;
} | null = null;

// ── Permission Handling ──

async function requestForegroundPermission(): Promise<boolean> {
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
      title: '位置情報の許可',
      message: '配車管理のために位置情報を使用します。',
      buttonPositive: '許可する',
    },
  );
  return granted === PermissionsAndroid.RESULTS.GRANTED;
}

async function requestBackgroundPermission(): Promise<boolean> {
  if (Platform.OS === 'ios') {
    // iOS handles this via Info.plist + authorizationLevel: 'always'
    return true;
  }
  // Android 10+ requires separate background permission
  if (Platform.Version >= 29) {
    const granted = await PermissionsAndroid.request(
      PermissionsAndroid.PERMISSIONS.ACCESS_BACKGROUND_LOCATION,
      {
        title: 'バックグラウンド位置情報',
        message: 'アプリがバックグラウンドでも位置情報を送信し続けるために許可が必要です。',
        buttonPositive: '許可する',
      },
    );
    return granted === PermissionsAndroid.RESULTS.GRANTED;
  }
  return true;
}

// ── Configuration ──

export async function configureBackgroundLocation() {
  Geolocation.setRNConfiguration({
    skipPermissionRequests: false,
    authorizationLevel: 'always',
    locationProvider: 'auto',
  });
  console.log('[Location] Service configured with always authorization');
}

// ── Location Watching ──

function startWatchingPosition() {
  if (watchId !== null) return;

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
    (error) => console.warn('[Location] Watch error:', error.message),
    {
      enableHighAccuracy: true,
      distanceFilter: 20,
      interval: 10000,
      fastestInterval: 5000,
    },
  );
}

function stopWatchingPosition() {
  if (watchId !== null) {
    Geolocation.clearWatch(watchId);
    watchId = null;
  }
}

// ── Backend Reporting ──

async function sendLocation(pos: NonNullable<typeof lastPosition>) {
  try {
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
  } catch (err: any) {
    console.warn('[Location] Send failed:', err.message);
  }
}

// ── Background Task (Android) ──

const BACKGROUND_TASK_OPTIONS = {
  taskName: 'LocationTracking',
  taskTitle: 'FleetTrack - 位置情報送信中',
  taskDesc: 'バックグラウンドで位置情報を送信しています',
  taskIcon: { name: 'ic_launcher', type: 'mipmap' },
  color: '#1e293b',
  parameters: { delay: 15000 },
};

async function backgroundTask(params: { delay: number }) {
  await new Promise<void>(async (resolve) => {
    startWatchingPosition();

    const loop = async () => {
      while (BackgroundService?.isRunning()) {
        if (lastPosition && getAccessToken()) {
          await sendLocation(lastPosition);
        }
        await sleep(params.delay);
      }
      resolve();
    };
    loop();
  });
}

function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

// ── iOS Background Loop ──

let iosIntervalId: ReturnType<typeof setInterval> | null = null;

function startIOSBackgroundLoop() {
  if (iosIntervalId !== null) return;
  iosIntervalId = setInterval(() => {
    if (lastPosition && getAccessToken()) {
      sendLocation(lastPosition);
    }
  }, 15000);
}

function stopIOSBackgroundLoop() {
  if (iosIntervalId !== null) {
    clearInterval(iosIntervalId);
    iosIntervalId = null;
  }
}

// ── Public API ──

export async function startTracking() {
  const hasForeground = await requestForegroundPermission();
  if (!hasForeground) {
    console.warn('[Location] Foreground permission denied');
    return;
  }

  const hasBackground = await requestBackgroundPermission();
  if (!hasBackground) {
    console.warn('[Location] Background permission denied, using foreground only');
  }

  if (Platform.OS === 'android' && BackgroundService) {
    // Android: use foreground service via BackgroundService
    try {
      await BackgroundService.start(backgroundTask, BACKGROUND_TASK_OPTIONS);
      console.log('[Location] Android background service started');
    } catch (err: any) {
      console.warn('[Location] Background service failed, falling back to foreground:', err.message);
      startWatchingPosition();
      startIOSBackgroundLoop();
    }
  } else {
    // iOS: watchPosition continues in background with Background Mode enabled
    startWatchingPosition();
    startIOSBackgroundLoop();
    console.log('[Location] iOS background tracking started');
  }
}

export async function stopTracking() {
  if (Platform.OS === 'android' && BackgroundService?.isRunning()) {
    await BackgroundService.stop();
  }

  stopWatchingPosition();
  stopIOSBackgroundLoop();
  lastPosition = null;
  console.log('[Location] Tracking stopped');
}

export function isTracking(): boolean {
  if (Platform.OS === 'android' && BackgroundService) {
    return BackgroundService.isRunning();
  }
  return watchId !== null;
}

export function updateAuthToken() {
  // Token is read dynamically from getAccessToken(), no action needed
}
