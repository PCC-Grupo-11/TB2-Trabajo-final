<script lang="ts">
	import { prediction } from '$lib/stores/prediction.svelte';
	import { AGENCIES, AGENCY_NAMES, COMPLAINT_TYPES } from '$lib/constants/mappings';
	import { COMPLAINT_DESCRIPTORS, DESCRIPTOR_LOCATIONS } from '$lib/constants/combo_mappings';

	let selectedComplaint = $state(prediction.complaintType);
	let selectedDescriptor = $state(prediction.descriptor);
	let selectedLocation = $state(prediction.locationType);

	let availableDescriptors = $derived(
		selectedComplaint ? (COMPLAINT_DESCRIPTORS[selectedComplaint] ?? []) : []
	);

	let availableLocations = $derived(
		selectedComplaint && selectedDescriptor
			? (DESCRIPTOR_LOCATIONS[`${selectedComplaint}|${selectedDescriptor}`] ?? [])
			: []
	);

	let realDescriptors = $derived(availableDescriptors.filter((d) => d !== 'UNKNOWN'));
	let realLocations = $derived(availableLocations.filter((l) => l !== 'UNKNOWN'));

	let allLocationsForComplaint = $derived(
		selectedComplaint
			? [
					...new Set(
						(COMPLAINT_DESCRIPTORS[selectedComplaint] ?? []).flatMap(
							(d) => DESCRIPTOR_LOCATIONS[`${selectedComplaint}|${d}`] ?? []
						)
					)
				]
			: []
	);

	let complaintForcesLocation = $derived(
		allLocationsForComplaint.length === 1 && allLocationsForComplaint[0] !== 'UNKNOWN'
	);

	let descriptorLocked = $derived(realDescriptors.length <= 1);
	let locationLocked = $derived(realLocations.length <= 1 || complaintForcesLocation);

	$effect(() => {
		prediction.complaintType = selectedComplaint;
	});

	$effect(() => {
		prediction.descriptor = selectedDescriptor;
	});

	$effect(() => {
		prediction.locationType = selectedLocation;
	});

	function autoSelectDescriptor() {
		if (realDescriptors.length === 0) {
			selectedDescriptor = 'UNKNOWN';
		} else if (realDescriptors.length === 1) {
			selectedDescriptor = realDescriptors[0];
		}
	}

	function autoSelectLocation() {
		if (realLocations.length === 0) {
			selectedLocation = 'UNKNOWN';
		} else if (realLocations.length === 1) {
			selectedLocation = realLocations[0];
		}
	}

	function onComplaintChange() {
		selectedDescriptor = '';
		selectedLocation = '';
		autoSelectDescriptor();
		if (complaintForcesLocation) {
			selectedLocation = allLocationsForComplaint[0];
		} else if (descriptorLocked) {
			autoSelectLocation();
		}
	}

	function onDescriptorChange() {
		selectedLocation = '';
		autoSelectLocation();
	}
</script>

<div class="flex flex-col gap-4">
	<div>
		<label
			for="ts"
			class="mb-0.5 block text-[11px] font-medium tracking-[0.12em] uppercase text-on-surface-variant/70"
		>
			Fecha y hora
		</label>
		<input
			type="datetime-local"
			id="ts"
			bind:value={prediction.tsLocal}
			class="h-10 w-full border-0 border-b border-outline-variant/50 bg-transparent px-0 py-2 text-sm text-on-surface placeholder:text-on-surface-variant/40 focus:border-on-surface focus:outline-none focus:ring-0"
		/>
	</div>

	<div>
		<label
			for="agency"
			class="mb-0.5 block text-[11px] font-medium tracking-[0.12em] uppercase text-on-surface-variant/70"
		>
			Agencia
		</label>
		<select
			id="agency"
			bind:value={prediction.agency}
			class="h-10 w-full border-0 border-b border-outline-variant/50 bg-transparent px-0 py-2 text-sm text-on-surface focus:border-on-surface focus:outline-none focus:ring-0"
		>
			<option value="" disabled selected hidden>Seleccionar agencia</option>
			{#each AGENCIES as a (a)}
				<option value={a}>{AGENCY_NAMES[a]}</option>
			{/each}
		</select>
	</div>

	<div>
		<label
			for="complaintType"
			class="mb-0.5 block text-[11px] font-medium tracking-[0.12em] uppercase text-on-surface-variant/70"
		>
			Tipo de queja
		</label>
		<select
			id="complaintType"
			bind:value={selectedComplaint}
			onchange={onComplaintChange}
			class="h-10 w-full border-0 border-b border-outline-variant/50 bg-transparent px-0 py-2 text-sm text-on-surface focus:border-on-surface focus:outline-none focus:ring-0"
		>
			<option value="" disabled selected hidden>Seleccionar tipo</option>
			{#each COMPLAINT_TYPES as ct (ct)}
				<option value={ct}>{ct}</option>
			{/each}
		</select>
	</div>

	<div>
		<label
			for="descriptor"
			class="mb-0.5 block text-[11px] font-medium tracking-[0.12em] uppercase text-on-surface-variant/70"
		>
			Descriptor
		</label>
		{#if descriptorLocked}
			<input
				type="text"
				id="descriptor"
				value={selectedDescriptor === ''
					? 'Seleccionar tipo primero'
					: selectedDescriptor === 'UNKNOWN'
						? 'No aplica'
						: selectedDescriptor}
				readonly
				class="h-10 w-full border-0 border-b border-outline-variant/50 bg-transparent px-0 py-2 text-sm text-on-surface-variant/50 cursor-not-allowed"
			/>
		{:else}
			<select
				id="descriptor"
				bind:value={selectedDescriptor}
				onchange={onDescriptorChange}
				disabled={!selectedComplaint}
				class="h-10 w-full border-0 border-b border-outline-variant/50 bg-transparent px-0 py-2 text-sm text-on-surface disabled:cursor-not-allowed disabled:opacity-40 focus:border-on-surface focus:outline-none focus:ring-0"
			>
				<option value="" disabled selected hidden>Seleccionar descriptor</option>
				{#each realDescriptors as d (d)}
					<option value={d}>{d}</option>
				{/each}
			</select>
		{/if}
	</div>

	<div>
		<label
			for="locationType"
			class="mb-0.5 block text-[11px] font-medium tracking-[0.12em] uppercase text-on-surface-variant/70"
		>
			Tipo de ubicacion
		</label>
		{#if locationLocked}
			<input
				type="text"
				id="locationType"
				value={selectedLocation === ''
					? selectedComplaint
						? 'Seleccionar descriptor primero'
						: 'Seleccionar tipo primero'
					: selectedLocation === 'UNKNOWN'
						? 'No aplica'
						: selectedLocation}
				readonly
				class="h-10 w-full border-0 border-b border-outline-variant/50 bg-transparent px-0 py-2 text-sm text-on-surface-variant/50 cursor-not-allowed"
			/>
		{:else}
			<select
				id="locationType"
				bind:value={selectedLocation}
				disabled={!selectedDescriptor}
				class="h-10 w-full border-0 border-b border-outline-variant/50 bg-transparent px-0 py-2 text-sm text-on-surface disabled:cursor-not-allowed disabled:opacity-40 focus:border-on-surface focus:outline-none focus:ring-0"
			>
				<option value="" disabled selected hidden>Seleccionar ubicacion</option>
				{#each realLocations as loc (loc)}
					<option value={loc}>{loc}</option>
				{/each}
			</select>
		{/if}
	</div>
</div>
