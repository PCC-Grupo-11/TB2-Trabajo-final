<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { auth } from '$lib/stores/auth';
	import { prediction } from '$lib/stores/prediction.svelte';
	import Map from './Map.svelte';
	import PredictionPanel from './PredictionPanel.svelte';
	import LogOut from '@lucide/svelte/icons/log-out';

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
		class="grid h-12 shrink-0 grid-cols-[1fr_auto_1fr] items-center border-b border-outline-variant/30 bg-surface px-gutter backdrop-blur-md"
	>
		<div></div>
		<div class="flex items-center gap-stack-lg">
			{#each [{ mode: 'prediction', label: 'Predicción' }, { mode: 'heatmap', label: 'Mapa de calor' }] as const as tab (tab.mode)}
				<button
					onclick={() => (activeMode = tab.mode)}
					class="cursor-pointer border-b-2 px-1 py-2 text-label-md font-medium transition-colors
						{activeMode === tab.mode
						? 'border-on-surface text-on-surface'
						: 'border-transparent text-on-surface-variant hover:text-on-surface'}"
				>
					{tab.label}
				</button>
			{/each}
		</div>
		<div class="flex justify-end">
			<button
				onclick={() => auth.logout()}
				title="Cerrar sesión"
				class="flex cursor-pointer items-center text-on-surface-variant transition-colors hover:text-on-surface"
			>
				<LogOut size={20} strokeWidth={1.5} />
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
