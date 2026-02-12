import client from './client';

export interface LatLng {
  lat: number;
  lng: number;
}

export interface RouteLeg {
  duration_sec: number;
  distance_meters: number;
  duration_text: string;
  distance_text: string;
}

export interface RouteResult {
  polyline: string;
  duration_sec: number;
  distance_meters: number;
  legs: RouteLeg[];
}

export async function computeRoute(
  origin: LatLng,
  destination: LatLng,
  intermediates?: LatLng[],
): Promise<RouteResult> {
  const { data } = await client.post<RouteResult>('/routes/compute', {
    origin,
    destination,
    intermediates: intermediates || [],
  });
  return data;
}
