import React, {useEffect} from 'react';
import {
  View,
  Text,
  FlatList,
  TouchableOpacity,
  StyleSheet,
  ActivityIndicator,
  RefreshControl,
} from 'react-native';
import {useRideStore} from '../stores/rideStore';
import type {Dispatch, DispatchStatus} from '../types';

const STATUS_LABELS: Record<DispatchStatus, string> = {
  pending: 'Pending',
  assigned: 'Assigned',
  accepted: 'Accepted',
  en_route: 'En Route',
  arrived: 'Arrived',
  completed: 'Completed',
  cancelled: 'Cancelled',
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

function formatDate(dateStr: string): string {
  const d = new Date(dateStr);
  return d.toLocaleDateString('en-US', {
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  });
}

function RideItem({ride}: {ride: Dispatch}) {
  return (
    <View style={styles.card}>
      <View style={styles.cardHeader}>
        <Text style={styles.date}>{formatDate(ride.created_at)}</Text>
        <View
          style={[
            styles.statusBadge,
            {backgroundColor: STATUS_COLORS[ride.status] || '#666'},
          ]}>
          <Text style={styles.statusText}>
            {STATUS_LABELS[ride.status] || ride.status}
          </Text>
        </View>
      </View>

      <View style={styles.routeContainer}>
        <View style={styles.routeRow}>
          <View style={[styles.dot, {backgroundColor: '#1a73e8'}]} />
          <Text style={styles.address} numberOfLines={1}>
            {ride.pickup_address}
          </Text>
        </View>
        {ride.dropoff_address && (
          <>
            <View style={styles.routeLine} />
            <View style={styles.routeRow}>
              <View style={[styles.dot, {backgroundColor: '#e74c3c'}]} />
              <Text style={styles.address} numberOfLines={1}>
                {ride.dropoff_address}
              </Text>
            </View>
          </>
        )}
      </View>

      {ride.estimated_duration_sec && (
        <Text style={styles.duration}>
          {Math.round(ride.estimated_duration_sec / 60)} min
          {ride.estimated_distance_m
            ? ` · ${(ride.estimated_distance_m / 1000).toFixed(1)} km`
            : ''}
        </Text>
      )}
    </View>
  );
}

export default function RideHistoryScreen() {
  const rideHistory = useRideStore(s => s.rideHistory);
  const isLoadingHistory = useRideStore(s => s.isLoadingHistory);
  const fetchHistory = useRideStore(s => s.fetchHistory);

  useEffect(() => {
    fetchHistory();
  }, [fetchHistory]);

  return (
    <View style={styles.container}>
      <Text style={styles.title}>Ride History</Text>
      <FlatList
        data={rideHistory}
        keyExtractor={item => item.id}
        renderItem={({item}) => <RideItem ride={item} />}
        contentContainerStyle={styles.list}
        refreshControl={
          <RefreshControl
            refreshing={isLoadingHistory}
            onRefresh={fetchHistory}
          />
        }
        ListEmptyComponent={
          isLoadingHistory ? (
            <ActivityIndicator
              size="large"
              color="#1a73e8"
              style={styles.loader}
            />
          ) : (
            <View style={styles.empty}>
              <Text style={styles.emptyText}>No ride history yet</Text>
              <Text style={styles.emptySubtext}>
                Your completed rides will appear here
              </Text>
            </View>
          )
        }
      />
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#f8f9fa',
  },
  title: {
    fontSize: 28,
    fontWeight: '700',
    color: '#333',
    paddingHorizontal: 20,
    paddingTop: 60,
    paddingBottom: 16,
  },
  list: {
    paddingHorizontal: 16,
    paddingBottom: 32,
  },
  card: {
    backgroundColor: '#fff',
    borderRadius: 14,
    padding: 16,
    marginBottom: 12,
    shadowColor: '#000',
    shadowOffset: {width: 0, height: 1},
    shadowOpacity: 0.08,
    shadowRadius: 3,
    elevation: 2,
  },
  cardHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: 12,
  },
  date: {
    fontSize: 13,
    color: '#888',
  },
  statusBadge: {
    paddingHorizontal: 10,
    paddingVertical: 4,
    borderRadius: 12,
  },
  statusText: {
    color: '#fff',
    fontSize: 12,
    fontWeight: '600',
  },
  routeContainer: {
    marginBottom: 8,
  },
  routeRow: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 10,
  },
  dot: {
    width: 10,
    height: 10,
    borderRadius: 5,
  },
  routeLine: {
    width: 2,
    height: 16,
    backgroundColor: '#ddd',
    marginLeft: 4,
  },
  address: {
    flex: 1,
    fontSize: 15,
    color: '#333',
  },
  duration: {
    fontSize: 13,
    color: '#888',
    marginTop: 4,
  },
  loader: {
    marginTop: 60,
  },
  empty: {
    alignItems: 'center',
    marginTop: 80,
  },
  emptyText: {
    fontSize: 18,
    fontWeight: '600',
    color: '#666',
  },
  emptySubtext: {
    fontSize: 14,
    color: '#999',
    marginTop: 8,
  },
});
