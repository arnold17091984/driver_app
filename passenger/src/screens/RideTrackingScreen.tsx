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

const STATUS_LABELS: Record<DispatchStatus, string> = {
  pending: 'Finding a driver...',
  assigned: 'Driver assigned',
  accepted: 'Driver is on the way',
  en_route: 'Driver is heading to pickup',
  arrived: 'Driver has arrived',
  completed: 'Ride completed',
  cancelled: 'Ride cancelled',
};

const STATUS_COLORS: Record<DispatchStatus, string> = {
  pending: '#f39c12',
  assigned: '#3498db',
  accepted: '#2ecc71',
  en_route: '#1a73e8',
  arrived: '#27ae60',
  completed: '#666',
  cancelled: '#e74c3c',
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

  // Also poll for ride status updates
  useEffect(() => {
    const interval = setInterval(fetchCurrentRide, 10000);
    return () => clearInterval(interval);
  }, [fetchCurrentRide]);

  // Navigate back when ride is completed or cancelled
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
            const msg =
              err.response?.data?.error?.message || 'Failed to cancel ride';
            Alert.alert('Error', msg);
          }
        },
      },
    ]);
  }, [cancelRide, rideId, navigation]);

  const status = currentRide?.status ?? 'pending';
  const canCancel = status === 'pending' || status === 'assigned';

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
      {/* Map showing driver location */}
      <MapView
        style={styles.map}
        provider={PROVIDER_DEFAULT}
        region={mapRegion}
        showsUserLocation>
        {/* Pickup marker */}
        {currentRide?.pickup_lat && currentRide?.pickup_lng && (
          <Marker
            coordinate={{
              latitude: currentRide.pickup_lat,
              longitude: currentRide.pickup_lng,
            }}
            title="Pickup"
            pinColor="#1a73e8"
          />
        )}

        {/* Driver location marker */}
        {driverLocation && (
          <Marker
            coordinate={{
              latitude: driverLocation.latitude,
              longitude: driverLocation.longitude,
            }}
            title="Driver"
            pinColor="#27ae60"
          />
        )}
      </MapView>

      {/* Status card */}
      <View style={styles.statusCard}>
        <View
          style={[
            styles.statusBadge,
            {backgroundColor: STATUS_COLORS[status] || '#666'},
          ]}>
          <Text style={styles.statusBadgeText}>
            {STATUS_LABELS[status] || status}
          </Text>
        </View>

        {status === 'pending' && (
          <View style={styles.loadingRow}>
            <ActivityIndicator size="small" color="#f39c12" />
            <Text style={styles.loadingText}>
              Looking for available drivers...
            </Text>
          </View>
        )}

        {currentRide && (
          <View style={styles.rideInfo}>
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

        {canCancel && (
          <TouchableOpacity
            style={styles.cancelButton}
            onPress={handleCancel}>
            <Text style={styles.cancelButtonText}>Cancel Ride</Text>
          </TouchableOpacity>
        )}

        {!canCancel && status !== 'completed' && status !== 'cancelled' && (
          <Text style={styles.cantCancelText}>
            Ride is in progress and cannot be cancelled
          </Text>
        )}
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
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
    borderTopLeftRadius: 20,
    borderTopRightRadius: 20,
    paddingHorizontal: 24,
    paddingTop: 20,
    paddingBottom: 40,
    shadowColor: '#000',
    shadowOffset: {width: 0, height: -3},
    shadowOpacity: 0.2,
    shadowRadius: 8,
    elevation: 10,
  },
  statusBadge: {
    alignSelf: 'flex-start',
    paddingHorizontal: 16,
    paddingVertical: 8,
    borderRadius: 20,
    marginBottom: 16,
  },
  statusBadgeText: {
    color: '#fff',
    fontSize: 15,
    fontWeight: '600',
  },
  loadingRow: {
    flexDirection: 'row',
    alignItems: 'center',
    marginBottom: 16,
    gap: 8,
  },
  loadingText: {
    fontSize: 14,
    color: '#888',
  },
  rideInfo: {
    marginBottom: 16,
  },
  infoRow: {
    flexDirection: 'row',
    marginBottom: 8,
  },
  infoLabel: {
    width: 60,
    fontSize: 13,
    fontWeight: '600',
    color: '#888',
  },
  infoValue: {
    flex: 1,
    fontSize: 15,
    color: '#333',
  },
  cancelButton: {
    backgroundColor: '#fee2e2',
    borderRadius: 12,
    paddingVertical: 14,
    alignItems: 'center',
  },
  cancelButtonText: {
    color: '#e74c3c',
    fontSize: 16,
    fontWeight: '600',
  },
  cantCancelText: {
    fontSize: 13,
    color: '#999',
    textAlign: 'center',
    fontStyle: 'italic',
  },
});
