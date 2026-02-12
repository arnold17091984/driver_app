import { useEffect, useRef, useState } from 'react';

const API_KEY = import.meta.env.VITE_GOOGLE_MAPS_API_KEY || '';

let loadPromise: Promise<void> | null = null;

function loadGoogleMaps(): Promise<void> {
  if (loadPromise) return loadPromise;

  // Already loaded (e.g. via script tag in HTML)
  if (window.google?.maps) {
    loadPromise = Promise.resolve();
    return loadPromise;
  }

  loadPromise = new Promise<void>((resolve, reject) => {
    // Google Maps calls this callback on auth errors
    (window as unknown as Record<string, () => void>).gm_authFailure = () => {
      console.error('Google Maps AUTH FAILURE - check API key, billing, and enabled APIs');
    };

    const script = document.createElement('script');
    script.src = `https://maps.googleapis.com/maps/api/js?key=${API_KEY}&v=weekly`;
    script.async = true;
    script.defer = true;
    script.onload = () => resolve();
    script.onerror = () => reject(new Error('Failed to load Google Maps script'));
    document.head.appendChild(script);
  });

  return loadPromise;
}

export function useGoogleMap(
  containerRef: React.RefObject<HTMLDivElement | null>,
  options: google.maps.MapOptions,
) {
  const mapRef = useRef<google.maps.Map | null>(null);
  const [ready, setReady] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!containerRef.current) return;
    let cancelled = false;

    loadGoogleMaps().then(() => {
      if (cancelled || !containerRef.current) return;
      try {
        const map = new google.maps.Map(containerRef.current, options);
        mapRef.current = map;
        setReady(true);
      } catch (err) {
        const msg = err instanceof Error ? err.message : String(err);
        console.error('Google Maps Map creation error:', msg);
        setError(msg);
      }
    }).catch((err) => {
      const msg = err instanceof Error ? err.message : String(err);
      console.error('Google Maps load error:', msg);
      if (!cancelled) setError(msg);
    });

    return () => {
      cancelled = true;
      mapRef.current = null;
      setReady(false);
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [containerRef]);

  return { map: mapRef.current, ready, error };
}
