<script lang="ts">
	import { prediction } from '$lib/stores/prediction.svelte';
	import {
		AGENCIES,
		AGENCY_NAMES,
		LOCATION_TYPES,
		COMPLAINT_TYPES,
		DESCRIPTORS
	} from '$lib/constants/mappings';

	const selectFields = [
		{
			id: 'agency',
			label: 'Agencia',
			placeholder: 'Seleccionar agencia',
			options: AGENCIES.map((v) => ({ value: v, label: AGENCY_NAMES[v] }))
		},
		{
			id: 'complaintType',
			label: 'Tipo de queja',
			placeholder: 'Seleccionar tipo',
			options: COMPLAINT_TYPES.map((v) => ({ value: v, label: v }))
		},
		{
			id: 'descriptor',
			label: 'Descriptor',
			placeholder: 'Seleccionar descriptor',
			options: DESCRIPTORS.map((v) => ({ value: v, label: v }))
		},
		{
			id: 'locationType',
			label: 'Tipo de ubicacion',
			placeholder: 'Seleccionar ubicacion',
			options: LOCATION_TYPES.map((v) => ({ value: v, label: v }))
		}
	] as const;
</script>

<div class="flex flex-col gap-3">
	<div>
		<label
			for="ts"
			class="mb-0.5 block text-xs font-medium tracking-wide text-on-surface-variant/70 uppercase"
		>
			Fecha y hora
		</label>
		<input
			type="datetime-local"
			id="ts"
			bind:value={prediction.tsLocal}
			class="h-9 w-full border-0 border-b border-outline-variant/50 bg-transparent px-0 py-2 text-sm text-on-surface placeholder:text-on-surface-variant/40 focus:border-on-surface focus:outline-none focus:ring-0"
		/>
	</div>

	{#each selectFields as field (field.id)}
		<div>
			<label
				for={field.id}
				class="mb-0.5 block text-xs font-medium tracking-wide text-on-surface-variant/70 uppercase"
			>
				{field.label}
			</label>
			<select
				id={field.id}
				bind:value={prediction[field.id]}
				class="h-9 w-full border-0 border-b border-outline-variant/50 bg-transparent px-0 py-2 text-sm text-on-surface focus:border-on-surface focus:outline-none focus:ring-0"
			>
				<option value="">{field.placeholder}</option>
				{#each field.options as opt (opt.value)}
					<option value={opt.value}>{opt.label}</option>
				{/each}
			</select>
		</div>
	{/each}
</div>
