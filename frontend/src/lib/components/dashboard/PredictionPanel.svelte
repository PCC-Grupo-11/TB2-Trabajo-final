<script lang="ts">
	import { prediction } from '$lib/stores/prediction.svelte';
	import { TIME_BUCKETS, getSeverityColor, getSeverityLabel } from '$lib/types/prediction';
	import { AGENCIES, LOCATION_TYPES, COMPLAINT_TYPES, DESCRIPTORS } from '$lib/constants/mappings';

	const selectFields = [
		{ id: 'agency', label: 'Agencia', placeholder: 'Seleccionar agencia', options: AGENCIES },
		{ id: 'complaintType', label: 'Tipo de queja', placeholder: 'Seleccionar tipo', options: COMPLAINT_TYPES },
		{ id: 'descriptor', label: 'Descriptor', placeholder: 'Seleccionar descriptor', options: DESCRIPTORS },
		{ id: 'locationType', label: 'Tipo de ubicacion', placeholder: 'Seleccionar ubicacion', options: LOCATION_TYPES },
	] as const;

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
			la configuracion de parametros.
		</p>
	</div>
{:else}
	<div class="flex flex-1 flex-col overflow-y-auto">
		<div class="flex flex-1 flex-col gap-stack-md px-stack-lg py-stack-md">
			<div>
				<label for="ts" class="mb-1 block text-label-sm font-medium text-on-surface-variant"
					>Fecha y hora</label
				>
				<input
					type="datetime-local"
					id="ts"
					bind:value={prediction.tsLocal}
					class="w-full rounded border border-outline-variant/50 bg-transparent px-4 py-3 text-on-surface focus:outline-none focus:ring-inset focus:ring-2 focus:ring-primary"
				/>
			</div>

			<div class="grid grid-cols-2 gap-3">
				{#each selectFields as field (field.id)}
					<div>
						<label for={field.id} class="mb-1 block text-label-sm font-medium text-on-surface-variant">
							{field.label}
						</label>
						<select
							id={field.id}
							bind:value={prediction[field.id]}
							class="w-full appearance-none rounded border border-outline-variant/50 bg-transparent px-4 py-3 text-on-surface focus:outline-none focus:ring-inset focus:ring-2 focus:ring-primary"
						>
							<option value="">{field.placeholder}</option>
							{#each field.options as opt (opt)}
								<option value={opt}>{opt}</option>
							{/each}
						</select>
					</div>
				{/each}
			</div>

			{#if prediction.error}
				<div class="rounded border border-red-200 bg-red-50 p-3 text-sm text-red-700">
					{prediction.error}
				</div>
			{/if}

		</div>

		{#if prediction.result}
			{@const res = prediction.result!}
			{@const maxProb = Math.max(...res.probabilities)}
			<div class="border-t border-outline-variant/30 px-stack-lg py-stack-md">
				<h3 class="mb-stack-md text-headline-md font-semibold text-on-surface">Resultado</h3>

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

				<h4 class="mb-2 text-label-sm font-medium uppercase tracking-wider text-on-surface-variant">
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
		{/if}
	</div>
{/if}
