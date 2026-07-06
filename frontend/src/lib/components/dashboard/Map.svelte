<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import maplibregl from 'maplibre-gl';
	import type { GeoJSONSource } from 'maplibre-gl';
	import 'maplibre-gl/dist/maplibre-gl.css';
	import { heatmap } from '$lib/stores/heatmap.svelte';

	export type MapAPI = {
		enableHeatmap: () => void;
		disableHeatmap: () => void;
	};

	let {
		onCoordinateSelect,
		onMapReady
	}: {
		onCoordinateSelect: (lat: number, lng: number) => void;
		onMapReady?: (api: MapAPI) => void;
	} = $props();

	let container: HTMLDivElement;
	let map: maplibregl.Map;
	let marker: maplibregl.Marker | null = null;
	let heatmapLayerAdded = $state(false);

	function addHeatmapLayer() {
		if (heatmapLayerAdded || !map || !map.isStyleLoaded()) return;
		map.addSource('heatmap-source', {
			type: 'geojson',
			data: { type: 'FeatureCollection', features: [] }
		});
		map.addLayer({
			id: 'heatmap-fill',
			type: 'fill',
			source: 'heatmap-source',
			paint: {
				'fill-color': [
					'match',
					['get', 'class'],
					0,
					'#98e9d5',
					1,
					'#7fd97d',
					2,
					'#c0e06c',
					3,
					'#f5cc51',
					4,
					'#f0a94f',
					5,
					'#df6d34',
					6,
					'#cd493b',
					7,
					'#aa2e32',
					'#aa2e32'
				],
				'fill-opacity': 0.4,
				'fill-outline-color': 'rgba(0,0,0,0.15)'
			}
		});
		heatmapLayerAdded = true;
	}

	function removeHeatmapLayer() {
		if (!heatmapLayerAdded || !map) return;
		if (map.getLayer('heatmap-fill')) map.removeLayer('heatmap-fill');
		if (map.getSource('heatmap-source')) map.removeSource('heatmap-source');
		heatmapLayerAdded = false;
	}

	function emitViewport() {
		if (!map) return;
		const bounds = map.getBounds();
		heatmap.updateViewport(
			{
				north: bounds.getNorth(),
				south: bounds.getSouth(),
				east: bounds.getEast(),
				west: bounds.getWest()
			},
			map.getZoom()
		);
	}

	function enableHeatmap() {
		addHeatmapLayer();
		emitViewport();
	}

	function disableHeatmap() {
		removeHeatmapLayer();
	}

	onMount(() => {
		map = new maplibregl.Map({
			container,
			style: 'https://basemaps.cartocdn.com/gl/positron-gl-style/style.json',
			center: [-74.006, 40.7128],
			zoom: 10
		});

		map.addControl(new maplibregl.NavigationControl(), 'top-left');

		map.on('click', (e: maplibregl.MapMouseEvent) => {
			const { lat, lng } = e.lngLat;
			if (marker) marker.remove();
			marker = new maplibregl.Marker({ color: '#000000' }).setLngLat([lng, lat]).addTo(map);
			onCoordinateSelect(lat, lng);
		});

		map.on('load', () => {
			if (heatmap.enabled) enableHeatmap();
			else emitViewport();
			onMapReady?.({ enableHeatmap, disableHeatmap });
		});

		// Keep the viewport in sync even while the heatmap is disabled. Otherwise,
		// panning/zooming in prediction mode leaves a stale viewport, and enabling
		// the heatmap fetches hexes for the old area (heatmap looks empty).
		// updateViewport() no-ops the fetch while disabled, so this stays cheap.
		map.on('moveend', () => {
			emitViewport();
		});
	});

	$effect(() => {
		const features = heatmap.getFeatures();
		if (heatmapLayerAdded && map?.getSource('heatmap-source')) {
			(map.getSource('heatmap-source') as GeoJSONSource).setData(features);
		}
	});

	onDestroy(() => {
		if (map) map.remove();
	});
</script>

<div bind:this={container} class="h-full w-full"></div>
