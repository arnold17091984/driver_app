/**
 * Push Notification Service
 *
 * Firebase/FCM push notifications. Gracefully disabled when Firebase is not configured.
 *
 * To enable:
 * 1. Add GoogleService-Info.plist (iOS) / google-services.json (Android)
 * 2. Enable Push Notifications capability in Xcode
 */

import { Alert } from 'react-native';
import client from './apiClient';

let messagingModule: any = null;

try {
  messagingModule = require('@react-native-firebase/messaging').default;
} catch {
  // Firebase not installed or not configured — notifications disabled
}

export async function setupNotifications() {
  if (!messagingModule) {
    console.log('Firebase not configured — push notifications disabled');
    return;
  }

  try {
    const authStatus = await messagingModule().requestPermission();
    const enabled =
      authStatus === messagingModule.AuthorizationStatus.AUTHORIZED ||
      authStatus === messagingModule.AuthorizationStatus.PROVISIONAL;

    if (!enabled) {
      console.log('Push notification permission denied');
      return;
    }

    const token = await messagingModule().getToken();
    await registerFCMToken(token);

    messagingModule().onTokenRefresh(async (newToken: string) => {
      await registerFCMToken(newToken);
    });

    messagingModule().onMessage(async (remoteMessage: any) => {
      const { title, body } = remoteMessage.notification || {};
      if (title) {
        Alert.alert(title, body || '');
      }
    });

    messagingModule().onNotificationOpenedApp((remoteMessage: any) => {
      console.log('Notification opened app:', remoteMessage.data);
    });

    const initialNotification = await messagingModule().getInitialNotification();
    if (initialNotification) {
      console.log('App opened from notification:', initialNotification.data);
    }

    console.log('Push notifications setup complete');
  } catch (error) {
    console.warn('Push notifications setup failed (Firebase may not be configured):', error);
  }
}

async function registerFCMToken(token: string) {
  try {
    await client.put('/notifications/fcm-token', { token });
  } catch (error) {
    console.error('Failed to register FCM token:', error);
  }
}

export { registerFCMToken };
