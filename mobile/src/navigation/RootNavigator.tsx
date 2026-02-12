import React from 'react';
import { NavigationContainer } from '@react-navigation/native';
import { createNativeStackNavigator } from '@react-navigation/native-stack';
import { useAuthStore } from '../stores/authStore';
import { LoginScreen } from '../screens/LoginScreen';
import { HomeScreen } from '../screens/HomeScreen';
import { TripDetailScreen } from '../screens/TripDetailScreen';
import { TripActiveScreen } from '../screens/TripActiveScreen';
import type { RootStackParamList } from '../types';

const Stack = createNativeStackNavigator<RootStackParamList>();

export function RootNavigator() {
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated);

  return (
    <NavigationContainer>
      <Stack.Navigator screenOptions={{ headerShown: false }}>
        {!isAuthenticated ? (
          <Stack.Screen name="Login" component={LoginScreen} />
        ) : (
          <>
            <Stack.Screen name="Home" component={HomeScreen} />
            <Stack.Screen
              name="TripDetail"
              component={TripDetailScreen}
              options={{
                headerShown: true,
                title: '配車詳細',
                headerBackTitle: '戻る',
              }}
            />
            <Stack.Screen
              name="TripActive"
              component={TripActiveScreen}
              options={{
                headerShown: true,
                title: '配車進行',
                headerBackTitle: '戻る',
              }}
            />
          </>
        )}
      </Stack.Navigator>
    </NavigationContainer>
  );
}
