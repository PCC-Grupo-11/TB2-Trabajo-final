<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import maplibregl from 'maplibre-gl';
	import 'maplibre-gl/dist/maplibre-gl.css';

	let { onCoordinateSelect }: { onCoordinateSelect: (lat: number, lng: number) => void } = $props();

	let container: HTMLDivElement;
	let map: maplibregl.Map;
	let marker: maplibregl.Marker | null = null;

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
	});

	onDestroy(() => {
		if (map) map.remove();
	});
</script>

<div bind:this={container} class="h-full w-full"></div>
