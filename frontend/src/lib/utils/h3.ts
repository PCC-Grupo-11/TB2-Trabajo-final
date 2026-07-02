import { polygonToCells, cellToBoundary, cellToLatLng } from 'h3-js';
import { getBorough } from '$lib/utils/borough';
import type { HexPrediction } from '$lib/types/heatmap';

const ZOOM_RESOLUTIONS: [number, number][] = [
	[0, 5],
	[8, 6],
	[10, 7],
	[12, 8],
	[14, 9]
];

export function zoomToResolution(zoom: number): number {
	let res = ZOOM_RESOLUTIONS[0][1];
	for (const [minZoom, resolution] of ZOOM_RESOLUTIONS) {
		if (zoom >= minZoom) res = resolution;
	}
	return res;
}

export function getCellsForViewport(
	north: number,
	south: number,
	east: number,
	west: number,
	resolution: number
): string[] {
	const pad = 0.04;
	const bbox: number[][] = [
		[north + pad, west - pad],
		[north + pad, east + pad],
		[south - pad, east + pad],
		[south - pad, west - pad],
		[north + pad, west - pad]
	];
	return polygonToCells(bbox, resolution, false);
}

export function cellToFeature(cell: string, prediction: HexPrediction): GeoJSON.Feature {
	const boundary = cellToBoundary(cell, true);
	return {
		type: 'Feature',
		properties: {
			hex: prediction.hex,
			class: prediction.class,
			confidence: prediction.confidence
		},
		geometry: {
			type: 'Polygon',
			coordinates: [boundary]
		}
	};
}

export function hexToBorough(hex: string): string {
	const [lat, lng] = cellToLatLng(hex);
	return getBorough(lat, lng);
}

export function emptyFeatureCollection(): GeoJSON.FeatureCollection {
	return { type: 'FeatureCollection', features: [] };
}
