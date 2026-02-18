import React, { useEffect, useState } from 'react';
import { View, Text, TouchableOpacity, StyleSheet, Alert, ActivityIndicator } from 'react-native';
import MapView, { Marker, PROVIDER_DEFAULT } from 'react-native-maps';
import Geolocation from '@react-native-community/geolocation';
import { useTripStore } from '../stores/tripStore';
import type { DispatchStatus } from '../types';

const STEPS: { status: DispatchStatus; label: string }[] = [
  { status: 'accepted', label: '受領' },
  { status: 'en_route', label: '移動中' },
  { status: 'arrived', label: '到着' },
  { status: 'completed', label: '完了' },
];

export function TripActiveScreen({ navigation }: any) {
  const { currentTrip, fetchCurrentTrip, startEnRoute, markArrived, completeTrip } = useTripStore();
  const [driverLocation, setDriverLocation] = useState<{
    latitude: number;
    longitude: number;
  } | null>(null);

  useEffect(() => {
    fetchCurrentTrip();
  }, [fetchCurrentTrip]);

  useEffect(() => {
    Geolocation.getCurrentPosition(
      (pos) => {
        setDriverLocation({
          latitude: pos.coords.latitude,
          longitude: pos.coords.longitude,
        });
      },
      () => {},
      { enableHighAccuracy: true, timeout: 10000 },
    );

    const watchId = Geolocation.watchPosition(
      (pos) => {
        setDriverLocation({
          latitude: pos.coords.latitude,
          longitude: pos.coords.longitude,
        });
      },
      () => {},
      { enableHighAccuracy: true, distanceFilter: 20 },
    );

    return () => Geolocation.clearWatch(watchId);
  }, []);

  if (!currentTrip) {
    return (
      <View style={styles.center}>
        <Text style={styles.emptyText}>配車データが見つかりません</Text>
      </View>
    );
  }

  const currentStepIndex = STEPS.findIndex((s) => s.status === currentTrip.status);

  const handleNext = async () => {
    try {
      switch (currentTrip.status) {
        case 'accepted':
          await startEnRoute(currentTrip.id);
          break;
        case 'en_route':
          await markArrived(currentTrip.id);
          break;
        case 'arrived':
          Alert.alert('完了確認', '配車を完了しますか？', [
            { text: 'キャンセル', style: 'cancel' },
            {
              text: '完了',
              onPress: async () => {
                await completeTrip(currentTrip.id);
                navigation.navigate('Home');
              },
            },
          ]);
          return;
      }
    } catch {
      Alert.alert('エラー', 'ステータス更新に失敗しました');
    }
  };

  const getNextAction = (): { label: string; color: string } => {
    switch (currentTrip.status) {
      case 'accepted':
        return { label: '出発する', color: '#8b5cf6' };
      case 'en_route':
        return { label: '到着した', color: '#22c55e' };
      case 'arrived':
        return { label: '完了する', color: '#3b82f6' };
      default:
        return { label: '', color: '#6b7280' };
    }
  };

  const nextAction = getNextAction();

  const pickupCoord = currentTrip.pickup_lat && currentTrip.pickup_lng
    ? { latitude: currentTrip.pickup_lat, longitude: currentTrip.pickup_lng }
    : null;
  const dropoffCoord = currentTrip.dropoff_lat && currentTrip.dropoff_lng
    ? { latitude: currentTrip.dropoff_lat, longitude: currentTrip.dropoff_lng }
    : null;

  const mapCenter = driverLocation || pickupCoord;

  return (
    <View style={styles.container}>
      {/* Map */}
      {mapCenter ? (
        <MapView
          style={styles.map}
          provider={PROVIDER_DEFAULT}
          showsUserLocation
          initialRegion={{
            ...mapCenter,
            latitudeDelta: 0.02,
            longitudeDelta: 0.02,
          }}>
          {pickupCoord && (
            <Marker
              coordinate={pickupCoord}
              title="乗車地点"
              description={currentTrip.pickup_address}
              pinColor="#16a34a"
            />
          )}
          {dropoffCoord && (
            <Marker
              coordinate={dropoffCoord}
              title="降車地点"
              description={currentTrip.dropoff_address}
              pinColor="#ef4444"
            />
          )}
        </MapView>
      ) : (
        <View style={[styles.map, styles.mapLoading]}>
          <ActivityIndicator size="large" color="#3b82f6" />
          <Text style={styles.mapLoadingText}>位置情報を取得中...</Text>
        </View>
      )}

      {/* Bottom panel */}
      <View style={styles.bottomPanel}>
        {/* Progress Stepper */}
        <View style={styles.stepper}>
          {STEPS.map((step, index) => (
            <View key={step.status} style={styles.stepItem}>
              <View
                style={[
                  styles.stepCircle,
                  index <= currentStepIndex ? styles.stepActive : styles.stepInactive,
                ]}
              >
                <Text style={styles.stepNumber}>{index + 1}</Text>
              </View>
              <Text
                style={[
                  styles.stepLabel,
                  index <= currentStepIndex ? styles.stepLabelActive : styles.stepLabelInactive,
                ]}
              >
                {step.label}
              </Text>
              {index < STEPS.length - 1 && (
                <View
                  style={[
                    styles.stepLine,
                    index < currentStepIndex ? styles.stepLineActive : styles.stepLineInactive,
                  ]}
                />
              )}
            </View>
          ))}
        </View>

        {/* Trip Info */}
        <View style={styles.infoCard}>
          <Text style={styles.purpose}>{currentTrip.purpose}</Text>
          <View style={styles.addressRow}>
            <Text style={styles.addressDot}>●</Text>
            <Text style={styles.addressText}>{currentTrip.pickup_address}</Text>
          </View>
          {currentTrip.dropoff_address && (
            <View style={styles.addressRow}>
              <Text style={[styles.addressDot, {color: '#ef4444'}]}>●</Text>
              <Text style={styles.addressText}>{currentTrip.dropoff_address}</Text>
            </View>
          )}
          {currentTrip.passenger_name && (
            <Text style={styles.passenger}>
              乗客: {currentTrip.passenger_name}
            </Text>
          )}
        </View>

        {/* Action Button */}
        {nextAction.label !== '' && (
          <TouchableOpacity
            style={[styles.actionButton, { backgroundColor: nextAction.color }]}
            onPress={handleNext}
          >
            <Text style={styles.actionButtonText}>{nextAction.label}</Text>
          </TouchableOpacity>
        )}
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: '#f1f5f9' },
  center: { flex: 1, justifyContent: 'center', alignItems: 'center' },
  emptyText: { color: '#94a3b8', fontSize: 16 },
  map: {
    flex: 2,
  },
  mapLoading: {
    justifyContent: 'center',
    alignItems: 'center',
    backgroundColor: '#e2e8f0',
    gap: 8,
  },
  mapLoadingText: {
    color: '#64748b',
    fontSize: 14,
  },
  bottomPanel: {
    flex: 3,
    padding: 16,
  },
  stepper: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    backgroundColor: '#fff',
    borderRadius: 12,
    padding: 16,
    marginBottom: 12,
  },
  stepItem: { alignItems: 'center', flex: 1 },
  stepCircle: {
    width: 32,
    height: 32,
    borderRadius: 16,
    justifyContent: 'center',
    alignItems: 'center',
  },
  stepActive: { backgroundColor: '#3b82f6' },
  stepInactive: { backgroundColor: '#e2e8f0' },
  stepNumber: { color: '#fff', fontWeight: '700', fontSize: 13 },
  stepLabel: { fontSize: 11, marginTop: 4, fontWeight: '500' },
  stepLabelActive: { color: '#3b82f6' },
  stepLabelInactive: { color: '#94a3b8' },
  stepLine: {
    position: 'absolute',
    top: 16,
    left: '60%',
    right: '-40%',
    height: 2,
  },
  stepLineActive: { backgroundColor: '#3b82f6' },
  stepLineInactive: { backgroundColor: '#e2e8f0' },
  infoCard: {
    backgroundColor: '#fff',
    borderRadius: 12,
    padding: 16,
    marginBottom: 12,
  },
  purpose: { fontSize: 18, fontWeight: '700', color: '#1e293b', marginBottom: 12 },
  addressRow: { flexDirection: 'row', alignItems: 'center', marginBottom: 6 },
  addressDot: { fontSize: 10, marginRight: 8, color: '#16a34a' },
  addressText: { fontSize: 14, color: '#475569', flex: 1 },
  passenger: { fontSize: 13, color: '#64748b', marginTop: 6 },
  actionButton: {
    borderRadius: 14,
    padding: 18,
    alignItems: 'center',
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 4 },
    shadowOpacity: 0.2,
    shadowRadius: 8,
    elevation: 6,
  },
  actionButtonText: { color: '#fff', fontSize: 18, fontWeight: '700' },
});
