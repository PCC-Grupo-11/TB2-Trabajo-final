import boroughs from '$lib/assets/boroughs.json';
import { point } from '@turf/helpers';
import booleanPointInPolygon from '@turf/boolean-point-in-polygon';
import type { Feature, Polygon } from 'geojson';

export function getBorough(lat: number, lng: number): string {
	const p = point([lng, lat]);

	const match = boroughs.features.find((f) => booleanPointInPolygon(p, f as Feature<Polygon>));

	return match?.properties?.BoroName
		? (match.properties.BoroName as string).toUpperCase()
		: 'Unspecified';
}
