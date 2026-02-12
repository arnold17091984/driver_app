package maps

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type LatLng struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type RouteLeg struct {
	DurationSec    int    `json:"duration_sec"`
	DistanceMeters int    `json:"distance_meters"`
	DurationText   string `json:"duration_text"`
	DistanceText   string `json:"distance_text"`
}

type RouteResult struct {
	Polyline       string     `json:"polyline"`
	DurationSec    int        `json:"duration_sec"`
	DistanceMeters int        `json:"distance_meters"`
	Legs           []RouteLeg `json:"legs"`
}

type Client struct {
	apiKey     string
	httpClient *http.Client
}

func NewClient(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// ComputeRoute calls the Google Routes API to get a driving route.
func (c *Client) ComputeRoute(ctx context.Context, origin, destination LatLng, intermediates []LatLng) (*RouteResult, error) {
	reqBody := buildRoutesRequest(origin, destination, intermediates)

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST",
		"https://routes.googleapis.com/directions/v2:computeRoutes", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Goog-Api-Key", c.apiKey)
	req.Header.Set("X-Goog-FieldMask",
		"routes.duration,routes.distanceMeters,routes.polyline.encodedPolyline,routes.legs.duration,routes.legs.distanceMeters")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("routes API request: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("routes API error (status %d): %s", resp.StatusCode, string(body))
	}

	var result routesAPIResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	if len(result.Routes) == 0 {
		return nil, fmt.Errorf("no routes returned")
	}

	route := result.Routes[0]
	r := &RouteResult{
		Polyline:       route.Polyline.EncodedPolyline,
		DurationSec:    parseDurationStr(route.Duration),
		DistanceMeters: route.DistanceMeters,
	}

	for _, leg := range route.Legs {
		dur := parseDurationStr(leg.Duration)
		r.Legs = append(r.Legs, RouteLeg{
			DurationSec:    dur,
			DistanceMeters: leg.DistanceMeters,
			DurationText:   formatDuration(dur),
			DistanceText:   formatDistance(leg.DistanceMeters),
		})
	}

	return r, nil
}

// --- Google Routes API request/response types ---

type routesAPIRequest struct {
	Origin                *routeWaypoint   `json:"origin"`
	Destination           *routeWaypoint   `json:"destination"`
	Intermediates         []routeWaypoint  `json:"intermediates,omitempty"`
	TravelMode            string           `json:"travelMode"`
	RoutingPreference     string           `json:"routingPreference"`
	ComputeAlternativeRoutes bool          `json:"computeAlternativeRoutes"`
}

type routeWaypoint struct {
	Location *routeLocation `json:"location"`
}

type routeLocation struct {
	LatLng *routeLatLng `json:"latLng"`
}

type routeLatLng struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type routesAPIResponse struct {
	Routes []routeData `json:"routes"`
}

type routeData struct {
	Duration       string          `json:"duration"`
	DistanceMeters int             `json:"distanceMeters"`
	Polyline       routePolyline   `json:"polyline"`
	Legs           []routeLegData  `json:"legs"`
}

type routePolyline struct {
	EncodedPolyline string `json:"encodedPolyline"`
}

type routeLegData struct {
	Duration       string `json:"duration"`
	DistanceMeters int    `json:"distanceMeters"`
}

func buildRoutesRequest(origin, destination LatLng, intermediates []LatLng) routesAPIRequest {
	req := routesAPIRequest{
		Origin: &routeWaypoint{
			Location: &routeLocation{LatLng: &routeLatLng{Latitude: origin.Lat, Longitude: origin.Lng}},
		},
		Destination: &routeWaypoint{
			Location: &routeLocation{LatLng: &routeLatLng{Latitude: destination.Lat, Longitude: destination.Lng}},
		},
		TravelMode:        "DRIVE",
		RoutingPreference: "TRAFFIC_AWARE",
	}

	for _, wp := range intermediates {
		req.Intermediates = append(req.Intermediates, routeWaypoint{
			Location: &routeLocation{LatLng: &routeLatLng{Latitude: wp.Lat, Longitude: wp.Lng}},
		})
	}

	return req
}

// parseDurationStr parses "123s" format from Routes API into seconds.
func parseDurationStr(s string) int {
	var secs int
	fmt.Sscanf(s, "%ds", &secs)
	return secs
}

func formatDuration(secs int) string {
	if secs < 60 {
		return fmt.Sprintf("%d sec", secs)
	}
	mins := secs / 60
	if mins < 60 {
		return fmt.Sprintf("%d min", mins)
	}
	hours := mins / 60
	remaining := mins % 60
	if remaining == 0 {
		return fmt.Sprintf("%d hr", hours)
	}
	return fmt.Sprintf("%d hr %d min", hours, remaining)
}

func formatDistance(meters int) string {
	if meters < 1000 {
		return fmt.Sprintf("%d m", meters)
	}
	km := float64(meters) / 1000.0
	return fmt.Sprintf("%.1f km", km)
}
