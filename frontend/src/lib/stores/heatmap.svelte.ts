import { prediction } from '$lib/stores/prediction.svelte';
import { bulkPredict } from '$lib/api/bulk-prediction';
import { SvelteMap } from 'svelte/reactivity';
import {
	zoomToResolution,
	getCellsForViewport,
	cellToFeature,
	emptyFeatureCollection
} from '$lib/utils/h3';
import type { HexPrediction, HeatmapViewport } from '$lib/types/heatmap';

class HeatmapStore {
	enabled = $state(false);
	resolution = $state(7);
	viewport = $state<HeatmapViewport | null>(null);
	cache = new SvelteMap<string, HexPrediction>();
	loading = $state(false);
	error = $state<string | null>(null);

	private abortController: AbortController | null = null;

	updateViewport(v: HeatmapViewport, zoom: number) {
		this.viewport = v;
		this.resolution = zoomToResolution(zoom);
		this.fetchMissing();
	}

	enable() {
		this.enabled = true;
		this.fetchMissing();
	}

	disable() {
		this.enabled = false;
		this.abortController?.abort();
	}

	refresh() {
		this.clearCache();
		this.fetchMissing();
	}

	getFeatures(): GeoJSON.FeatureCollection {
		if (!this.viewport) return emptyFeatureCollection();
		const cells = getCellsForViewport(
			this.viewport.north,
			this.viewport.south,
			this.viewport.east,
			this.viewport.west,
			this.resolution
		);
		const features: GeoJSON.Feature[] = [];
		for (const cell of cells) {
			const pred = this.cache.get(cell);
			if (pred) features.push(cellToFeature(cell, pred));
		}
		return { type: 'FeatureCollection', features };
	}

	clearCache() {
		this.cache.clear();
	}

	hasValidParams(): boolean {
		return prediction.agency !== '' && prediction.complaintType !== '';
	}

	fetchMissing() {
		if (!this.enabled || !this.viewport) return;
		if (!this.hasValidParams()) return;

		const cells = getCellsForViewport(
			this.viewport.north,
			this.viewport.south,
			this.viewport.east,
			this.viewport.west,
			this.resolution
		);
		const missing = cells.filter((c) => !this.cache.has(c));

		if (missing.length === 0) return;

		this.abortController?.abort();
		this.abortController = new AbortController();
		this.loading = true;
		this.error = null;

		this.doFetch(missing, this.abortController.signal);
	}

	private async doFetch(hexes: string[], signal: AbortSignal) {
		try {
			const response = await bulkPredict(
				{
					ts: prediction.ts,
					agency: prediction.agency,
					complaint_type: prediction.complaintType,
					descriptor: prediction.descriptor,
					location_type: prediction.locationType,
					borough: prediction.borough,
					h3_hexes: hexes
				},
				signal
			);

			for (const hex of response.results) {
				this.cache.set(hex.hex, hex);
			}
		} catch (err) {
			if (err instanceof Error && err.name === 'AbortError') return;
			this.error = err instanceof Error ? err.message : 'Error al obtener predicciones.';
		} finally {
			this.loading = false;
		}
	}
}

export const heatmap = new HeatmapStore();
