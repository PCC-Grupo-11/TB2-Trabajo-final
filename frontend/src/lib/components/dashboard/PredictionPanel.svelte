<script lang="ts">
	import { prediction } from '$lib/stores/prediction.svelte';
	import { TIME_BUCKET_FRIENDLY, getSeverityBorder } from '$lib/types/prediction';
	import PredictionForm from './PredictionForm.svelte';
	import ProbabilityDistribution from './ProbabilityDistribution.svelte';
	import Clock from '@lucide/svelte/icons/clock';

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

	<div class="min-h-0 flex-1 overflow-y-auto flex flex-col">
		{#if activeTab === 'config'}
			<div class="px-stack-lg py-stack-md">
				<PredictionForm />
			</div>
		{:else if prediction.result}
			{@const res = prediction.result!}
			<div class="flex-1 px-stack-lg pt-5 pb-stack-md">
				<div>
					<div class="border-l-[3px] pl-4 {getSeverityBorder(res.class)}">
						<p class="text-[11px] font-medium uppercase tracking-[0.12em] text-on-surface-variant">
							Tiempo estimado de resolución
						</p>
						<p class="mt-2 text-[1.6rem] leading-tight font-extrabold text-on-surface">
							{TIME_BUCKET_FRIENDLY[res.class] ?? `Clase ${res.class}`}
						</p>
						<p
							class="mt-3 text-[11px] font-medium uppercase tracking-[0.12em] text-on-surface-variant"
						>
							Confianza
						</p>
						<p class="mt-0.5 text-[1.3rem] leading-tight font-extrabold text-on-surface">
							{(res.confidence * 100).toFixed(1)}%
						</p>
					</div>
				</div>

				<div class="my-5 border-t border-outline-variant/10"></div>

				<ProbabilityDistribution probabilities={res.probabilities} winnerIndex={res.class} />
			</div>

			<div
				class="shrink-0 flex items-center justify-end gap-1.5 px-stack-lg py-3 border-t border-outline-variant/10 mx-6 text-xs text-on-surface-variant/40"
			>
				<Clock size={12} strokeWidth={2} />
				{#if res.cached}
					<span>Almacenado en cache</span>
				{:else}
					<span>Procesado en {res.latency_ms.toFixed(2)} ms</span>
				{/if}
			</div>
		{:else}
			<div class="flex items-center justify-center py-stack-lg text-on-surface-variant">
				<p class="text-body-md">Esperando predicción...</p>
			</div>
		{/if}
	</div>
</div>
