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

// Safety cap: at very low zoom the viewport can cover a huge area and
// polygonToCells would produce tens of thousands of cells, freezing the tab
// and hammering the backend. Above this we skip fetching/rendering and ask
// the user to zoom in.
const MAX_CELLS = 2000;

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
		if (cells.length > MAX_CELLS) return emptyFeatureCollection();
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
		if (cells.length > MAX_CELLS) {
			this.error = 'Acerca el mapa para ver el mapa de calor.';
			return;
		}
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
			const response = await bulkPredict({ ...prediction.heatmapParams, h3_hexes: hexes }, signal);

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
