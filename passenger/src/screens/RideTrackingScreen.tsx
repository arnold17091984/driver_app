import React, {useEffect, useCallback} from 'react';
import {
  View,
  Text,
  TouchableOpacity,
  StyleSheet,
  Alert,
  ActivityIndicator,
} from 'react-native';
import MapView, {Marker, PROVIDER_DEFAULT} from 'react-native-maps';
import {useRideStore} from '../stores/rideStore';
import type {NativeStackScreenProps} from '@react-navigation/native-stack';
import type {DispatchStatus} from '../types';

type AppStackParamList = {
  Home: undefined;
  RideTracking: {rideId: string};
  RideHistory: undefined;
};

type Props = NativeStackScreenProps<AppStackParamList, 'RideTracking'>;

const STATUS_CONFIG: Record<
  DispatchStatus,
  {label: string; color: string; bg: string}
> = {
  pending: {label: 'Finding a driver...', color: '#d97706', bg: '#fffbeb'},
  assigned: {label: 'Driver assigned', color: '#2563eb', bg: '#eff6ff'},
  accepted: {label: 'Driver is on the way', color: '#16a34a', bg: '#f0fdf4'},
  en_route: {label: 'Heading to pickup', color: '#2563eb', bg: '#eff6ff'},
  arrived: {label: 'Driver has arrived', color: '#16a34a', bg: '#f0fdf4'},
  completed: {label: 'Ride completed', color: '#64748b', bg: '#f8fafc'},
  cancelled: {label: 'Ride cancelled', color: '#dc2626', bg: '#fef2f2'},
};

export default function RideTrackingScreen({route, navigation}: Props) {
  const {rideId} = route.params;

  const currentRide = useRideStore(s => s.currentRide);
  const driverLocation = useRideStore(s => s.driverLocation);
  const fetchCurrentRide = useRideStore(s => s.fetchCurrentRide);
  const pollDriverLocation = useRideStore(s => s.pollDriverLocation);
  const stopPolling = useRideStore(s => s.stopPolling);
  const cancelRide = useRideStore(s => s.cancelRide);

  useEffect(() => {
    fetchCurrentRide();
    pollDriverLocation(rideId);
    return () => stopPolling();
  }, [rideId, fetchCurrentRide, pollDriverLocation, stopPolling]);

  useEffect(() => {
    const interval = setInterval(fetchCurrentRide, 10000);
    return () => clearInterval(interval);
  }, [fetchCurrentRide]);

  useEffect(() => {
    if (
      currentRide &&
      (currentRide.status === 'completed' || currentRide.status === 'cancelled')
    ) {
      const message =
        currentRide.status === 'completed'
          ? 'Your ride has been completed!'
          : 'Your ride has been cancelled.';
      Alert.alert('Ride Update', message, [
        {text: 'OK', onPress: () => navigation.goBack()},
      ]);
    }
  }, [currentRide, navigation]);

  const handleCancel = useCallback(() => {
    Alert.alert('Cancel Ride', 'Are you sure you want to cancel this ride?', [
      {text: 'No', style: 'cancel'},
      {
        text: 'Yes, Cancel',
        style: 'destructive',
        onPress: async () => {
          try {
            await cancelRide(rideId);
            navigation.goBack();
          } catch (err: any) {
            Alert.alert(
              'Error',
              err.response?.data?.error?.message || 'Failed to cancel ride',
            );
          }
        },
      },
    ]);
  }, [cancelRide, rideId, navigation]);

  const status = currentRide?.status ?? 'pending';
  const canCancel = status === 'pending' || status === 'assigned';
  const config = STATUS_CONFIG[status] || STATUS_CONFIG.pending;

  const mapRegion = driverLocation
    ? {
        latitude: driverLocation.latitude,
        longitude: driverLocation.longitude,
        latitudeDelta: 0.02,
        longitudeDelta: 0.02,
      }
    : currentRide?.pickup_lat && currentRide?.pickup_lng
      ? {
          latitude: currentRide.pickup_lat,
          longitude: currentRide.pickup_lng,
          latitudeDelta: 0.02,
          longitudeDelta: 0.02,
        }
      : {
          latitude: 35.6812,
          longitude: 139.7671,
          latitudeDelta: 0.05,
          longitudeDelta: 0.05,
        };

  return (
    <View style={styles.container}>
      <MapView
        style={styles.map}
        provider={PROVIDER_DEFAULT}
        region={mapRegion}
        showsUserLocation>
        {currentRide?.pickup_lat && currentRide?.pickup_lng && (
          <Marker
            coordinate={{
              latitude: currentRide.pickup_lat,
              longitude: currentRide.pickup_lng,
            }}
            title="Pickup"
            pinColor="#2563eb"
          />
        )}
        {driverLocation && (
          <Marker
            coordinate={{
              latitude: driverLocation.latitude,
              longitude: driverLocation.longitude,
            }}
            title="Driver"
            pinColor="#16a34a"
          />
        )}
      </MapView>

      {/* Status card (web-style panel) */}
      <View style={styles.statusCard}>
        {/* Status badge */}
        <View style={[styles.statusBadge, {backgroundColor: config.bg}]}>
          <View
            style={[styles.statusDot, {backgroundColor: config.color}]}
          />
          <Text style={[styles.statusText, {color: config.color}]}>
            {config.label}
          </Text>
        </View>

        {/* Searching spinner */}
        {status === 'pending' && (
          <View style={styles.searchingContainer}>
            <ActivityIndicator size="large" color="#16a34a" />
            <Text style={styles.searchingTitle}>Looking for drivers</Text>
            <Text style={styles.searchingSubtitle}>
              This usually takes a few moments...
            </Text>
          </View>
        )}

        {/* Ride info (web info-box style) */}
        {currentRide && (
          <View style={styles.infoBox}>
            <View style={styles.infoRow}>
              <Text style={styles.infoLabel}>Pickup</Text>
              <Text style={styles.infoValue} numberOfLines={1}>
                {currentRide.pickup_address}
              </Text>
            </View>
            {currentRide.dropoff_address && (
              <View style={styles.infoRow}>
                <Text style={styles.infoLabel}>Dropoff</Text>
                <Text style={styles.infoValue} numberOfLines={1}>
                  {currentRide.dropoff_address}
                </Text>
              </View>
            )}
          </View>
        )}

        {/* Cancel button (web danger style) */}
        {canCancel && (
          <TouchableOpacity
            style={styles.cancelButton}
            onPress={handleCancel}
            activeOpacity={0.7}>
            <Text style={styles.cancelButtonText}>Cancel Ride</Text>
          </TouchableOpacity>
        )}

        {!canCancel && status !== 'completed' && status !== 'cancelled' && (
          <Text style={styles.inProgressText}>
            Ride is in progress
          </Text>
        )}
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#f1f5f9',
  },
  map: {
    flex: 1,
  },
  statusCard: {
    position: 'absolute',
    bottom: 0,
    left: 0,
    right: 0,
    backgroundColor: '#fff',
    borderTopLeftRadius: 16,
    borderTopRightRadius: 16,
    paddingHorizontal: 20,
    paddingTop: 20,
    paddingBottom: 40,
    shadowColor: '#000',
    shadowOffset: {width: 0, height: -3},
    shadowOpacity: 0.1,
    shadowRadius: 12,
    elevation: 10,
  },
  statusBadge: {
    flexDirection: 'row',
    alignItems: 'center',
    alignSelf: 'flex-start',
    paddingHorizontal: 12,
    paddingVertical: 6,
    borderRadius: 8,
    gap: 8,
    marginBottom: 16,
  },
  statusDot: {
    width: 8,
    height: 8,
    borderRadius: 4,
  },
  statusText: {
    fontSize: 14,
    fontWeight: '600',
  },
  searchingContainer: {
    alignItems: 'center',
    paddingVertical: 16,
  },
  searchingTitle: {
    fontSize: 16,
    fontWeight: '700',
    color: '#0f172a',
    marginTop: 12,
  },
  searchingSubtitle: {
    fontSize: 13,
    color: '#94a3b8',
    marginTop: 4,
  },
  infoBox: {
    backgroundColor: '#f8fafc',
    borderRadius: 12,
    padding: 14,
    gap: 10,
    marginBottom: 16,
  },
  infoRow: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
  },
  infoLabel: {
    fontSize: 13,
    color: '#94a3b8',
    fontWeight: '500',
  },
  infoValue: {
    flex: 1,
    fontSize: 14,
    fontWeight: '600',
    color: '#0f172a',
    textAlign: 'right',
    marginLeft: 12,
  },
  cancelButton: {
    backgroundColor: '#fff',
    borderWidth: 1,
    borderColor: '#fecaca',
    borderRadius: 10,
    paddingVertical: 12,
    alignItems: 'center',
  },
  cancelButtonText: {
    color: '#ef4444',
    fontSize: 15,
    fontWeight: '600',
  },
  inProgressText: {
    fontSize: 13,
    color: '#94a3b8',
    textAlign: 'center',
    fontStyle: 'italic',
  },
});
