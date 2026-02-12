/**
 * Push Notification Service
 *
 * Uses @react-native-firebase/messaging for FCM push notifications.
 *
 * SETUP REQUIRED:
 * 1. Install: yarn add @react-native-firebase/app @react-native-firebase/messaging
 * 2. Configure Firebase in iOS (GoogleService-Info.plist) and Android (google-services.json)
 * 3. iOS: Enable Push Notifications capability in Xcode
 */

import { Alert } from 'react-native';
import messaging from '@react-native-firebase/messaging';
import client from './apiClient';

export async function setupNotifications() {
  try {
    const authStatus = await messaging().requestPermission();
    const enabled =
      authStatus === messaging.AuthorizationStatus.AUTHORIZED ||
      authStatus === messaging.AuthorizationStatus.PROVISIONAL;

    if (!enabled) {
      console.log('Push notification permission denied');
      return;
    }

    // Get FCM token and register with backend
    const token = await messaging().getToken();
    await registerFCMToken(token);

    // Handle token refresh
    messaging().onTokenRefresh(async (newToken) => {
      await registerFCMToken(newToken);
    });

    // Handle foreground messages
    messaging().onMessage(async (remoteMessage) => {
      const { title, body } = remoteMessage.notification || {};
      if (title) {
        Alert.alert(title, body || '');
      }
    });

    // Handle background/quit state notification tap
    messaging().onNotificationOpenedApp((remoteMessage) => {
      console.log('Notification opened app:', remoteMessage.data);
    });

    // Check if app was opened from a notification (cold start)
    const initialNotification = await messaging().getInitialNotification();
    if (initialNotification) {
      console.log('App opened from notification:', initialNotification.data);
    }

    console.log('Push notifications setup complete');
  } catch (error) {
    console.error('Failed to setup notifications:', error);
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
