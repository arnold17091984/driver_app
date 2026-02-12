import React, {useEffect, useState, useRef, useCallback} from 'react';
import {
  View,
  Text,
  TextInput,
  TouchableOpacity,
  StyleSheet,
  Alert,
  ActivityIndicator,
  Dimensions,
  Animated,
  Keyboard,
} from 'react-native';
import MapView, {Marker, PROVIDER_DEFAULT} from 'react-native-maps';
import {useRideStore} from '../stores/rideStore';
import {useAuthStore} from '../stores/authStore';
import {getCurrentPosition} from '../services/locationService';
import type {NativeStackScreenProps} from '@react-navigation/native-stack';

const {height: SCREEN_HEIGHT} = Dimensions.get('window');
const SHEET_HEIGHT = 320;

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
  const user = useAuthStore(s => s.user);
  const logout = useAuthStore(s => s.logout);

  // Fetch current location and check for active ride on mount
  useEffect(() => {
    getCurrentPosition().then(coords => {
      setUserLocation(coords);
      mapRef.current?.animateToRegion(
        {
          ...coords,
          latitudeDelta: 0.01,
          longitudeDelta: 0.01,
        },
        500,
      );
    });
    fetchCurrentRide();
  }, [fetchCurrentRide]);

  // Navigate to tracking screen when ride becomes active
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

  const handleRequestRide = async () => {
    if (!userLocation) {
      Alert.alert('Error', 'Unable to determine your location');
      return;
    }
    if (!dropoffAddress.trim()) {
      Alert.alert('Error', 'Please enter a destination');
      return;
    }

    try {
      await requestRide(
        'Current Location',
        userLocation.latitude,
        userLocation.longitude,
        dropoffAddress.trim(),
        undefined,
        undefined,
        user?.name,
      );
      closeSheet();
      setDropoffAddress('');
    } catch (err: any) {
      const msg =
        err.response?.data?.error?.message || 'Failed to request ride';
      Alert.alert('Error', msg);
    }
  };

  const sheetTranslateY = sheetAnim.interpolate({
    inputRange: [0, 1],
    outputRange: [SHEET_HEIGHT, 0],
  });

  return (
    <View style={styles.container}>
      {/* Map */}
      <MapView
        ref={mapRef}
        style={styles.map}
        provider={PROVIDER_DEFAULT}
        showsUserLocation
        showsMyLocationButton
        initialRegion={{
          latitude: userLocation?.latitude ?? 35.6812,
          longitude: userLocation?.longitude ?? 139.7671,
          latitudeDelta: 0.01,
          longitudeDelta: 0.01,
        }}>
        {userLocation && (
          <Marker
            coordinate={userLocation}
            title="Your Location"
            pinColor="#1a73e8"
          />
        )}
      </MapView>

      {/* Top bar */}
      <View style={styles.topBar}>
        <TouchableOpacity
          style={styles.historyButton}
          onPress={() => navigation.navigate('RideHistory')}>
          <Text style={styles.historyButtonText}>History</Text>
        </TouchableOpacity>
        <TouchableOpacity style={styles.logoutButton} onPress={logout}>
          <Text style={styles.logoutButtonText}>Logout</Text>
        </TouchableOpacity>
      </View>

      {/* Destination input bar */}
      {!showSheet && (
        <View style={styles.searchBar}>
          <TouchableOpacity style={styles.searchInput} onPress={openSheet}>
            <Text style={styles.searchPlaceholder}>Where to?</Text>
          </TouchableOpacity>
        </View>
      )}

      {/* Bottom sheet for ride request */}
      {showSheet && (
        <Animated.View
          style={[
            styles.bottomSheet,
            {transform: [{translateY: sheetTranslateY}]},
          ]}>
          <View style={styles.sheetHandle} />
          <Text style={styles.sheetTitle}>Request a Ride</Text>

          <View style={styles.sheetField}>
            <Text style={styles.fieldLabel}>Pickup</Text>
            <Text style={styles.fieldValue}>Current Location</Text>
          </View>

          <View style={styles.sheetField}>
            <Text style={styles.fieldLabel}>Destination</Text>
            <TextInput
              style={styles.destinationInput}
              placeholder="Enter destination address"
              placeholderTextColor="#999"
              value={dropoffAddress}
              onChangeText={setDropoffAddress}
              autoFocus
            />
          </View>

          <View style={styles.sheetButtons}>
            <TouchableOpacity style={styles.cancelButton} onPress={closeSheet}>
              <Text style={styles.cancelButtonText}>Cancel</Text>
            </TouchableOpacity>
            <TouchableOpacity
              style={[
                styles.requestButton,
                isRequesting && styles.requestButtonDisabled,
              ]}
              onPress={handleRequestRide}
              disabled={isRequesting}>
              {isRequesting ? (
                <ActivityIndicator color="#fff" />
              ) : (
                <Text style={styles.requestButtonText}>Request Ride</Text>
              )}
            </TouchableOpacity>
          </View>
        </Animated.View>
      )}
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
  topBar: {
    position: 'absolute',
    top: 60,
    left: 16,
    right: 16,
    flexDirection: 'row',
    justifyContent: 'space-between',
  },
  historyButton: {
    backgroundColor: '#fff',
    paddingHorizontal: 16,
    paddingVertical: 10,
    borderRadius: 20,
    shadowColor: '#000',
    shadowOffset: {width: 0, height: 2},
    shadowOpacity: 0.15,
    shadowRadius: 4,
    elevation: 3,
  },
  historyButtonText: {
    fontSize: 14,
    fontWeight: '600',
    color: '#333',
  },
  logoutButton: {
    backgroundColor: '#fff',
    paddingHorizontal: 16,
    paddingVertical: 10,
    borderRadius: 20,
    shadowColor: '#000',
    shadowOffset: {width: 0, height: 2},
    shadowOpacity: 0.15,
    shadowRadius: 4,
    elevation: 3,
  },
  logoutButtonText: {
    fontSize: 14,
    fontWeight: '600',
    color: '#e74c3c',
  },
  searchBar: {
    position: 'absolute',
    bottom: 40,
    left: 16,
    right: 16,
  },
  searchInput: {
    backgroundColor: '#fff',
    borderRadius: 12,
    paddingHorizontal: 20,
    paddingVertical: 18,
    shadowColor: '#000',
    shadowOffset: {width: 0, height: 2},
    shadowOpacity: 0.2,
    shadowRadius: 6,
    elevation: 5,
  },
  searchPlaceholder: {
    fontSize: 18,
    color: '#999',
    fontWeight: '500',
  },
  bottomSheet: {
    position: 'absolute',
    bottom: 0,
    left: 0,
    right: 0,
    height: SHEET_HEIGHT,
    backgroundColor: '#fff',
    borderTopLeftRadius: 20,
    borderTopRightRadius: 20,
    paddingHorizontal: 24,
    paddingTop: 12,
    shadowColor: '#000',
    shadowOffset: {width: 0, height: -3},
    shadowOpacity: 0.2,
    shadowRadius: 8,
    elevation: 10,
  },
  sheetHandle: {
    width: 40,
    height: 4,
    backgroundColor: '#ddd',
    borderRadius: 2,
    alignSelf: 'center',
    marginBottom: 16,
  },
  sheetTitle: {
    fontSize: 20,
    fontWeight: '700',
    color: '#333',
    marginBottom: 20,
  },
  sheetField: {
    marginBottom: 16,
  },
  fieldLabel: {
    fontSize: 13,
    color: '#888',
    fontWeight: '600',
    marginBottom: 6,
    textTransform: 'uppercase',
    letterSpacing: 0.5,
  },
  fieldValue: {
    fontSize: 16,
    color: '#333',
    backgroundColor: '#f0f0f0',
    borderRadius: 8,
    padding: 12,
  },
  destinationInput: {
    fontSize: 16,
    color: '#333',
    backgroundColor: '#f0f0f0',
    borderRadius: 8,
    padding: 12,
  },
  sheetButtons: {
    flexDirection: 'row',
    marginTop: 12,
    gap: 12,
  },
  cancelButton: {
    flex: 1,
    backgroundColor: '#f0f0f0',
    borderRadius: 12,
    paddingVertical: 14,
    alignItems: 'center',
  },
  cancelButtonText: {
    fontSize: 16,
    fontWeight: '600',
    color: '#666',
  },
  requestButton: {
    flex: 2,
    backgroundColor: '#1a73e8',
    borderRadius: 12,
    paddingVertical: 14,
    alignItems: 'center',
  },
  requestButtonDisabled: {
    opacity: 0.7,
  },
  requestButtonText: {
    fontSize: 16,
    fontWeight: '600',
    color: '#fff',
  },
});
