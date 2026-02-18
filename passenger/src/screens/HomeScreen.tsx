import React, {useEffect, useState, useRef, useCallback} from 'react';
import {
  View,
  Text,
  TextInput,
  TouchableOpacity,
  StyleSheet,
  Alert,
  ActivityIndicator,
  Animated,
  Keyboard,
  FlatList,
} from 'react-native';
import MapView, {Marker, PROVIDER_DEFAULT} from 'react-native-maps';
import {useRideStore} from '../stores/rideStore';
import {useAuthStore} from '../stores/authStore';
import {getCurrentPosition} from '../services/locationService';
import type {NativeStackScreenProps} from '@react-navigation/native-stack';
import type {VehicleETA} from '../types';

type AppStackParamList = {
  Home: undefined;
  RideTracking: {rideId: string};
  RideHistory: undefined;
};

type Props = NativeStackScreenProps<AppStackParamList, 'Home'>;

export default function HomeScreen({navigation}: Props) {
  const mapRef = useRef<MapView>(null);
  const sheetAnim = useRef(new Animated.Value(0)).current;

  const [userLocation, setUserLocation] = useState<{
    latitude: number;
    longitude: number;
  } | null>(null);
  const [dropoffAddress, setDropoffAddress] = useState('');
  const [showSheet, setShowSheet] = useState(false);

  const currentRide = useRideStore(s => s.currentRide);
  const isRequesting = useRideStore(s => s.isRequesting);
  const requestRide = useRideStore(s => s.requestRide);
  const fetchCurrentRide = useRideStore(s => s.fetchCurrentRide);
  const nearbyVehicles = useRideStore(s => s.nearbyVehicles);
  const selectedVehicle = useRideStore(s => s.selectedVehicle);
  const isLoadingVehicles = useRideStore(s => s.isLoadingVehicles);
  const fetchNearbyVehicles = useRideStore(s => s.fetchNearbyVehicles);
  const selectVehicle = useRideStore(s => s.selectVehicle);
  const user = useAuthStore(s => s.user);
  const logout = useAuthStore(s => s.logout);

  const [sheetStep, setSheetStep] = useState<'destination' | 'vehicles'>('destination');

  useEffect(() => {
    getCurrentPosition().then(coords => {
      setUserLocation(coords);
      mapRef.current?.animateToRegion(
        {...coords, latitudeDelta: 0.01, longitudeDelta: 0.01},
        500,
      );
    });
    fetchCurrentRide();
  }, [fetchCurrentRide]);

  useEffect(() => {
    if (
      currentRide &&
      currentRide.status !== 'completed' &&
      currentRide.status !== 'cancelled'
    ) {
      navigation.navigate('RideTracking', {rideId: currentRide.id});
    }
  }, [currentRide, navigation]);

  const openSheet = useCallback(() => {
    setShowSheet(true);
    Animated.spring(sheetAnim, {
      toValue: 1,
      useNativeDriver: true,
      tension: 60,
      friction: 12,
    }).start();
  }, [sheetAnim]);

  const closeSheet = useCallback(() => {
    Keyboard.dismiss();
    Animated.timing(sheetAnim, {
      toValue: 0,
      duration: 200,
      useNativeDriver: true,
    }).start(() => setShowSheet(false));
  }, [sheetAnim]);

  const handleSearchVehicles = async () => {
    if (!userLocation) {
      Alert.alert('Error', 'Unable to determine your location');
      return;
    }
    if (!dropoffAddress.trim()) {
      Alert.alert('Error', 'Please enter a destination');
      return;
    }
    Keyboard.dismiss();
    await fetchNearbyVehicles(userLocation.latitude, userLocation.longitude);
    setSheetStep('vehicles');
  };

  const handleSelectVehicle = async (vehicle: VehicleETA) => {
    if (!userLocation) return;
    selectVehicle(vehicle);
    try {
      await requestRide(
        'Current Location',
        userLocation.latitude,
        userLocation.longitude,
        dropoffAddress.trim(),
        undefined,
        undefined,
        user?.name,
        vehicle.vehicle_id,
      );
      closeSheet();
      setDropoffAddress('');
      setSheetStep('destination');
    } catch (err: any) {
      Alert.alert(
        'Error',
        err.response?.data?.error?.message || 'Failed to request ride',
      );
    }
  };

  const handleBackToDestination = () => {
    setSheetStep('destination');
    selectVehicle(null);
  };

  const formatETA = (seconds: number) => {
    const mins = Math.round(seconds / 60);
    return mins <= 1 ? '1 min' : `${mins} min`;
  };

  const formatDistance = (meters: number) => {
    if (meters < 1000) return `${meters} m`;
    return `${(meters / 1000).toFixed(1)} km`;
  };

  const sheetTranslateY = sheetAnim.interpolate({
    inputRange: [0, 1],
    outputRange: [400, 0],
  });

  return (
    <View style={styles.container}>
      {/* Map */}
      {userLocation ? (
        <MapView
          ref={mapRef}
          style={styles.map}
          provider={PROVIDER_DEFAULT}
          showsUserLocation
          showsMyLocationButton
          initialRegion={{
            latitude: userLocation.latitude,
            longitude: userLocation.longitude,
            latitudeDelta: 0.01,
            longitudeDelta: 0.01,
          }}>
          <Marker
            coordinate={userLocation}
            title="Your Location"
            pinColor="#2563eb"
          />
        </MapView>
      ) : (
        <View style={[styles.map, styles.mapLoading]}>
          <ActivityIndicator size="large" color="#16a34a" />
          <Text style={styles.mapLoadingText}>Getting your location...</Text>
        </View>
      )}

      {/* Stats overlay (like web dashboard) */}
      <View style={styles.statsOverlay}>
        <View style={styles.statCard}>
          <View style={[styles.statDot, {backgroundColor: '#16a34a'}]} />
          <Text style={styles.statLabel}>Ready</Text>
        </View>
      </View>

      {/* Top bar */}
      <View style={styles.topBar}>
        <TouchableOpacity
          style={styles.topButton}
          onPress={() => navigation.navigate('RideHistory')}
          activeOpacity={0.7}>
          <Text style={styles.topButtonText}>History</Text>
        </TouchableOpacity>
        <TouchableOpacity
          style={styles.topButton}
          onPress={logout}
          activeOpacity={0.7}>
          <Text style={[styles.topButtonText, {color: '#dc2626'}]}>
            Logout
          </Text>
        </TouchableOpacity>
      </View>

      {/* "Where to?" search bar (like web booking FAB) */}
      {!showSheet && (
        <View style={styles.fabContainer}>
          <TouchableOpacity
            style={styles.fabButton}
            onPress={openSheet}
            activeOpacity={0.8}>
            <Text style={styles.fabIcon}>📍</Text>
            <Text style={styles.fabText}>Where to?</Text>
          </TouchableOpacity>
        </View>
      )}

      {/* Booking bottom sheet (matches web BookingOverlay) */}
      {showSheet && (
        <Animated.View
          style={[
            styles.bottomSheet,
            {transform: [{translateY: sheetTranslateY}]},
          ]}>
          {/* Header */}
          <View style={styles.sheetHeader}>
            <TouchableOpacity
              style={styles.backButton}
              onPress={sheetStep === 'vehicles' ? handleBackToDestination : closeSheet}
              activeOpacity={0.7}>
              <Text style={styles.backButtonText}>{sheetStep === 'vehicles' ? '<' : '✕'}</Text>
            </TouchableOpacity>
            <Text style={styles.sheetTitle}>
              {sheetStep === 'destination' ? 'Request a Ride' : 'Select Vehicle'}
            </Text>
            <View style={{width: 32}} />
          </View>

          {sheetStep === 'destination' ? (
            <>
              {/* Location inputs */}
              <View style={styles.locationContainer}>
                <View style={styles.dotsColumn}>
                  <View style={[styles.locationDot, {backgroundColor: '#16a34a'}]} />
                  <View style={styles.locationLine} />
                  <View style={[styles.locationDot, {backgroundColor: '#ef4444'}]} />
                </View>
                <View style={styles.inputsColumn}>
                  <View style={styles.locationField}>
                    <Text style={styles.locationActive}>Current Location</Text>
                  </View>
                  <View style={styles.locationDivider} />
                  <TextInput
                    style={styles.locationInput}
                    placeholder="Enter destination address"
                    placeholderTextColor="#94a3b8"
                    value={dropoffAddress}
                    onChangeText={setDropoffAddress}
                    autoFocus
                  />
                </View>
              </View>

              {/* Search vehicles button */}
              <TouchableOpacity
                style={[
                  styles.requestButton,
                  (!dropoffAddress.trim() || isLoadingVehicles) &&
                    styles.requestButtonDisabled,
                ]}
                onPress={handleSearchVehicles}
                disabled={!dropoffAddress.trim() || isLoadingVehicles}
                activeOpacity={0.8}>
                {isLoadingVehicles ? (
                  <ActivityIndicator color="#fff" />
                ) : (
                  <Text style={styles.requestButtonText}>Search Vehicles</Text>
                )}
              </TouchableOpacity>
            </>
          ) : (
            <>
              {/* Vehicle list */}
              {isLoadingVehicles ? (
                <View style={styles.vehicleLoading}>
                  <ActivityIndicator size="large" color="#16a34a" />
                  <Text style={styles.vehicleLoadingText}>Finding nearby vehicles...</Text>
                </View>
              ) : nearbyVehicles.length === 0 ? (
                <View style={styles.vehicleLoading}>
                  <Text style={styles.noVehiclesText}>No vehicles available nearby</Text>
                </View>
              ) : (
                <FlatList
                  data={nearbyVehicles.filter(v => v.is_available)}
                  keyExtractor={item => item.vehicle_id}
                  style={styles.vehicleList}
                  renderItem={({item}) => (
                    <TouchableOpacity
                      style={[
                        styles.vehicleCard,
                        selectedVehicle?.vehicle_id === item.vehicle_id && styles.vehicleCardSelected,
                      ]}
                      onPress={() => handleSelectVehicle(item)}
                      disabled={isRequesting}
                      activeOpacity={0.7}>
                      <View style={styles.vehicleInfo}>
                        <Text style={styles.vehicleName}>{item.vehicle_name}</Text>
                        <Text style={styles.vehicleDriver}>{item.driver_name}</Text>
                        <Text style={styles.vehiclePlate}>{item.plate}</Text>
                      </View>
                      <View style={styles.vehicleEta}>
                        <Text style={styles.vehicleEtaTime}>{formatETA(item.duration_sec)}</Text>
                        <Text style={styles.vehicleEtaDist}>{formatDistance(item.distance_m)}</Text>
                      </View>
                      {isRequesting && selectedVehicle?.vehicle_id === item.vehicle_id && (
                        <ActivityIndicator size="small" color="#16a34a" style={styles.vehicleSpinner} />
                      )}
                    </TouchableOpacity>
                  )}
                />
              )}
            </>
          )}
        </Animated.View>
      )}
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
  mapLoading: {
    justifyContent: 'center',
    alignItems: 'center',
    backgroundColor: '#e2e8f0',
    gap: 12,
  },
  mapLoadingText: {
    color: '#64748b',
    fontSize: 14,
  },
  statsOverlay: {
    position: 'absolute',
    top: 60,
    left: 16,
    flexDirection: 'row',
    gap: 8,
    zIndex: 10,
  },
  statCard: {
    backgroundColor: '#fff',
    borderRadius: 10,
    paddingHorizontal: 14,
    paddingVertical: 8,
    flexDirection: 'row',
    alignItems: 'center',
    gap: 8,
    shadowColor: '#000',
    shadowOffset: {width: 0, height: 2},
    shadowOpacity: 0.1,
    shadowRadius: 8,
    elevation: 3,
  },
  statDot: {
    width: 8,
    height: 8,
    borderRadius: 4,
  },
  statLabel: {
    fontSize: 13,
    fontWeight: '500',
    color: '#64748b',
  },
  topBar: {
    position: 'absolute',
    top: 60,
    right: 16,
    flexDirection: 'row',
    gap: 8,
    zIndex: 10,
  },
  topButton: {
    backgroundColor: '#fff',
    paddingHorizontal: 14,
    paddingVertical: 8,
    borderRadius: 10,
    shadowColor: '#000',
    shadowOffset: {width: 0, height: 2},
    shadowOpacity: 0.1,
    shadowRadius: 8,
    elevation: 3,
  },
  topButtonText: {
    fontSize: 13,
    fontWeight: '600',
    color: '#334155',
  },
  fabContainer: {
    position: 'absolute',
    bottom: 40,
    left: 16,
    right: 16,
  },
  fabButton: {
    backgroundColor: '#16a34a',
    borderRadius: 12,
    paddingVertical: 16,
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
    gap: 8,
    shadowColor: '#16a34a',
    shadowOffset: {width: 0, height: 4},
    shadowOpacity: 0.3,
    shadowRadius: 16,
    elevation: 6,
  },
  fabIcon: {
    fontSize: 18,
  },
  fabText: {
    color: '#fff',
    fontSize: 16,
    fontWeight: '700',
  },
  bottomSheet: {
    position: 'absolute',
    bottom: 0,
    left: 0,
    right: 0,
    backgroundColor: '#fff',
    borderTopLeftRadius: 16,
    borderTopRightRadius: 16,
    paddingBottom: 40,
    shadowColor: '#000',
    shadowOffset: {width: 0, height: -3},
    shadowOpacity: 0.15,
    shadowRadius: 12,
    elevation: 10,
  },
  sheetHeader: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    paddingHorizontal: 20,
    paddingTop: 18,
    paddingBottom: 16,
    borderBottomWidth: 1,
    borderBottomColor: '#f1f5f9',
  },
  backButton: {
    width: 32,
    height: 32,
    borderRadius: 8,
    backgroundColor: '#f1f5f9',
    alignItems: 'center',
    justifyContent: 'center',
  },
  backButtonText: {
    fontSize: 16,
    color: '#475569',
    fontWeight: '600',
  },
  sheetTitle: {
    fontSize: 16,
    fontWeight: '700',
    color: '#0f172a',
  },
  locationContainer: {
    flexDirection: 'row',
    backgroundColor: '#f8fafc',
    borderRadius: 12,
    marginHorizontal: 20,
    marginTop: 16,
    padding: 14,
    gap: 12,
  },
  dotsColumn: {
    alignItems: 'center',
    paddingTop: 6,
  },
  locationDot: {
    width: 10,
    height: 10,
    borderRadius: 5,
  },
  locationLine: {
    width: 2,
    height: 24,
    backgroundColor: '#cbd5e1',
    marginVertical: 4,
  },
  inputsColumn: {
    flex: 1,
  },
  locationField: {
    backgroundColor: '#fff',
    borderRadius: 8,
    padding: 10,
    borderWidth: 2,
    borderColor: '#16a34a',
  },
  locationActive: {
    fontSize: 14,
    fontWeight: '600',
    color: '#16a34a',
  },
  locationDivider: {
    height: 8,
  },
  locationInput: {
    backgroundColor: '#fff',
    borderRadius: 8,
    padding: 10,
    fontSize: 14,
    color: '#0f172a',
    borderWidth: 2,
    borderColor: '#e2e8f0',
  },
  requestButton: {
    backgroundColor: '#16a34a',
    borderRadius: 12,
    paddingVertical: 14,
    alignItems: 'center',
    marginHorizontal: 20,
    marginTop: 20,
    shadowColor: '#16a34a',
    shadowOffset: {width: 0, height: 4},
    shadowOpacity: 0.3,
    shadowRadius: 16,
    elevation: 6,
  },
  requestButtonDisabled: {
    backgroundColor: '#94a3b8',
    shadowColor: '#94a3b8',
  },
  requestButtonText: {
    color: '#fff',
    fontSize: 16,
    fontWeight: '700',
  },
  vehicleLoading: {
    paddingVertical: 40,
    alignItems: 'center',
    gap: 12,
  },
  vehicleLoadingText: {
    color: '#64748b',
    fontSize: 14,
  },
  noVehiclesText: {
    color: '#94a3b8',
    fontSize: 15,
    fontWeight: '500',
  },
  vehicleList: {
    maxHeight: 300,
    paddingHorizontal: 20,
    marginTop: 8,
  },
  vehicleCard: {
    flexDirection: 'row',
    alignItems: 'center',
    backgroundColor: '#f8fafc',
    borderRadius: 12,
    padding: 14,
    marginBottom: 8,
    borderWidth: 2,
    borderColor: 'transparent',
  },
  vehicleCardSelected: {
    borderColor: '#16a34a',
    backgroundColor: '#f0fdf4',
  },
  vehicleInfo: {
    flex: 1,
  },
  vehicleName: {
    fontSize: 15,
    fontWeight: '700',
    color: '#0f172a',
  },
  vehicleDriver: {
    fontSize: 13,
    color: '#475569',
    marginTop: 2,
  },
  vehiclePlate: {
    fontSize: 12,
    color: '#94a3b8',
    marginTop: 2,
  },
  vehicleEta: {
    alignItems: 'flex-end',
    marginLeft: 12,
  },
  vehicleEtaTime: {
    fontSize: 16,
    fontWeight: '700',
    color: '#16a34a',
  },
  vehicleEtaDist: {
    fontSize: 12,
    color: '#64748b',
    marginTop: 2,
  },
  vehicleSpinner: {
    marginLeft: 8,
  },
});
