<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { auth } from '$lib/stores/auth';
	import { prediction } from '$lib/stores/prediction.svelte';
	import Map from './Map.svelte';
	import PredictionPanel from './PredictionPanel.svelte';

	function formatCoord(val: number, type: 'lat' | 'lng'): string {
		const dir = type === 'lat' ? (val >= 0 ? 'N' : 'S') : val >= 0 ? 'E' : 'W';
		return `${Math.abs(val).toFixed(4)}° ${dir}`;
	}

	let activeMode = $state<'prediction' | 'heatmap'>('prediction');

	onMount(() => {
		if (!auth.isAuthenticated()) goto(resolve('/login'));
	});

	function handleCoordinateSelect(lat: number, lng: number) {
		prediction.setCoords(lat, lng);
	}
</script>

<div class="flex h-screen flex-col">
	<nav
		class="flex h-14 shrink-0 items-center border-b border-outline-variant/30 bg-surface px-gutter backdrop-blur-md"
	>
		<div class="flex flex-1 items-center justify-center gap-stack-lg">
			{#each [{ mode: 'prediction', label: 'Predicción' }, { mode: 'heatmap', label: 'Mapa de calor' }] as const as tab (tab.mode)}
				<button
					onclick={() => (activeMode = tab.mode)}
					class="cursor-pointer border-b-2 px-1 py-4 text-label-md font-medium transition-colors
						{activeMode === tab.mode
						? 'border-on-surface text-on-surface'
						: 'border-transparent text-on-surface-variant hover:text-on-surface'}"
				>
					{tab.label}
				</button>
			{/each}
		</div>
		<div class="flex items-center">
			<button
				onclick={() => auth.logout()}
				class="flex cursor-pointer items-center gap-2 text-label-md text-on-surface-variant transition-colors hover:text-on-surface"
			>
				<svg
					class="h-5 w-5"
					fill="none"
					viewBox="0 0 24 24"
					stroke="currentColor"
					stroke-width="1.5"
				>
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						d="M15.75 9V5.25A2.25 2.25 0 0013.5 3h-6a2.25 2.25 0 00-2.25 2.25v13.5A2.25 2.25 0 007.5 21h6a2.25 2.25 0 002.25-2.25V15m3 0l3-3m0 0l-3-3m3 3H9"
					/>
				</svg>
				Cerrar sesión
			</button>
		</div>
	</nav>

	<div class="flex flex-1 overflow-hidden">
		<div class="flex-1 relative">
			<Map onCoordinateSelect={handleCoordinateSelect} />
			{#if prediction.lat !== null}
				<div
					class="absolute top-4 right-4 rounded-lg border border-outline-variant/30 bg-surface/80 px-3 py-2 shadow-sm backdrop-blur-sm"
				>
					<p class="font-mono text-sm text-on-surface">
						{formatCoord(prediction.lat!, 'lat')}, {formatCoord(prediction.lng!, 'lng')}
					</p>
				</div>
			{/if}
		</div>

		<aside class="flex w-90 flex-col border-l border-outline-variant/30 bg-surface">
			{#if activeMode === 'prediction'}
				<PredictionPanel />
			{:else}
				<div class="flex flex-1 items-center justify-center p-stack-lg text-center">
					<p class="text-body-md text-on-surface-variant">Mapa de Calor - Proximamente</p>
				</div>
			{/if}
		</aside>
	</div>
</div>
