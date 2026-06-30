<script lang="ts">
	import { prediction } from '$lib/stores/prediction.svelte';
	import { TIME_BUCKETS, getSeverityColor, getSeverityLabel } from '$lib/types/prediction';
	import PredictionForm from './PredictionForm.svelte';

	let activeTab = $state<'config' | 'result'>('config');
	let hasResult = $derived(prediction.result !== null);

	$effect(() => {
		void prediction.lat;
		void prediction.lng;
		void prediction.tsLocal;
		void prediction.agency;
		void prediction.complaintType;
		void prediction.descriptor;
		void prediction.locationType;
		prediction.submitPrediction();
	});
</script>

{#if prediction.lat === null}
	<div
		class="flex flex-1 flex-col items-center justify-center gap-stack-md px-stack-lg text-center"
	>
		<div class="flex h-16 w-16 items-center justify-center rounded-full bg-surface-container-high">
			<svg
				class="h-8 w-8 text-on-surface-variant"
				fill="none"
				viewBox="0 0 24 24"
				stroke="currentColor"
				stroke-width="1.5"
			>
				<path
					stroke-linecap="round"
					stroke-linejoin="round"
					d="M15 10.5a3 3 0 11-6 0 3 3 0 016 0z"
				/>
				<path
					stroke-linecap="round"
					stroke-linejoin="round"
					d="M19.5 10.5c0 7.142-7.5 11.25-7.5 11.25S4.5 17.642 4.5 10.5a7.5 7.5 0 1115 0z"
				/>
			</svg>
		</div>
		<h2 class="text-headline-md font-semibold text-on-surface">Iniciar Prediccion</h2>
		<p class="max-w-60 text-body-md text-on-surface-variant">
			Haga clic en cualquier lugar del mapa para seleccionar una coordenada geoespacial y comenzar
			la configuración de parametros.
		</p>
	</div>
{:else}
	<div class="flex flex-1 flex-col">
		<div class="flex shrink-0 border-b border-outline-variant/30">
			{#each [{ id: 'config', label: 'Configuración' }, { id: 'result', label: 'Resultado' }] as tab (tab.id)}
				{@const locked = tab.id === 'result' && !hasResult}
				<button
					type="button"
					onclick={() => {
						if (!locked) activeTab = tab.id as typeof activeTab;
					}}
					class="flex-1 border-b-2 px-3 py-2.5 text-label-md font-medium transition-colors
						{activeTab === tab.id
						? 'border-on-surface text-on-surface cursor-pointer'
						: locked
							? 'border-transparent text-on-surface-variant/40 cursor-not-allowed'
							: 'border-transparent text-on-surface-variant hover:text-on-surface cursor-pointer'}"
				>
					{tab.label}
				</button>
			{/each}
		</div>

		<div class="min-h-0 flex-1 overflow-y-auto">
			{#if activeTab === 'config'}
				<div class="px-stack-lg py-stack-md">
					{#if prediction.error}
						<div class="mb-4 rounded border border-red-200 bg-red-50 p-3 text-sm text-red-700">
							{prediction.error}
						</div>
					{/if}
					<PredictionForm />
				</div>
			{:else if prediction.result}
				{@const res = prediction.result!}
				{@const maxProb = Math.max(...res.probabilities)}
				<div class="px-stack-lg py-stack-md">
					<div
						class="mb-stack-md rounded border border-outline-variant/30 bg-surface-container-low p-stack-md"
					>
						<p class="text-label-sm font-medium uppercase tracking-wider text-on-surface-variant">
							Tiempo estimado de resolucion
						</p>
						<p class="mt-1 text-headline-lg font-bold text-on-surface">
							{TIME_BUCKETS[res.class] ?? `Clase ${res.class}`}
						</p>
						<p class="mt-1 text-label-md text-on-surface-variant">
							Confianza: <span class="font-semibold text-on-surface"
								>{(res.confidence * 100).toFixed(1)}%</span
							>
							<span class="ml-1 text-label-sm">({getSeverityLabel(res.confidence)})</span>
						</p>
					</div>

					<h4
						class="mb-2 text-label-sm font-medium uppercase tracking-wider text-on-surface-variant"
					>
						Distribucion de probabilidad
					</h4>
					<div class="flex flex-col gap-2">
						{#each res.probabilities as prob, i (i)}
							{@const pct = maxProb > 0 ? (prob / maxProb) * 100 : 0}
							<div class="flex items-center gap-3">
								<span class="w-36 shrink-0 text-right text-label-sm text-on-surface-variant">
									{TIME_BUCKETS[i] ?? `Clase ${i}`}
								</span>
								<div class="h-2 flex-1 overflow-hidden rounded bg-surface-container-high">
									<div
										class="h-full rounded {getSeverityColor(i)} transition-all"
										style="width: {pct}%"
									></div>
								</div>
								<span
									class="w-14 shrink-0 text-right font-mono text-label-sm text-on-surface-variant"
								>
									{(prob * 100).toFixed(1)}%
								</span>
							</div>
						{/each}
					</div>

					<div
						class="mt-stack-md flex items-center gap-stack-sm text-label-sm text-on-surface-variant"
					>
						{#if res.cached}
							<span class="rounded bg-surface-container-high px-2 py-0.5 text-label-sm">Cache</span>
						{/if}
						{#if res.latency_ms > 0}
							<span>{res.latency_ms.toFixed(3)}ms</span>
						{/if}
					</div>
				</div>
			{:else}
				<div class="flex items-center justify-center py-stack-lg text-on-surface-variant">
					<p class="text-body-md">Esperando prediccion...</p>
				</div>
			{/if}
		</div>
	</div>
{/if}
