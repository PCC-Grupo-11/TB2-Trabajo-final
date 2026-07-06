<script lang="ts">
	import MetricsChart from './MetricsChart.svelte';
	import type { HistoryPoint } from '$lib/types/metrics';

	let {
		cpuPercent,
		memoryBytes,
		history,
		entityKey,
		peticiones,
		showReq = false
	}: {
		cpuPercent: number;
		memoryBytes: number;
		history: HistoryPoint[];
		entityKey: string;
		peticiones?: number;
		showReq?: boolean;
	} = $props();

	function formatBytes(bytes: number): string {
		if (bytes === 0) return '0 B';
		const units = ['B', 'KB', 'MB', 'GB'];
		const i = Math.floor(Math.log(bytes) / Math.log(1024));
		return `${(bytes / Math.pow(1024, i)).toFixed(2)} ${units[i]}`;
	}
</script>

<div class="flex gap-4">
	<div class="flex flex-col justify-center gap-3 min-w-0 shrink-0 w-28">
		<div>
			<p class="text-[11px] font-medium uppercase tracking-[0.12em] text-on-surface-variant">CPU</p>
			<p class="mt-0.5 text-[1.3rem] leading-tight font-extrabold text-on-surface">
				{cpuPercent.toFixed(2)}%
			</p>
		</div>
		<div>
			<p class="text-[11px] font-medium uppercase tracking-[0.12em] text-on-surface-variant">RAM</p>
			<p class="mt-0.5 text-[1.3rem] leading-tight font-extrabold text-on-surface">
				{formatBytes(memoryBytes)}
			</p>
		</div>
		{#if peticiones !== undefined}
			<div>
				<p class="text-[11px] font-medium uppercase tracking-[0.12em] text-on-surface-variant">
					Peticiones
				</p>
				<p class="mt-0.5 text-[1.3rem] leading-tight font-extrabold text-on-surface">
					{peticiones}
				</p>
			</div>
		{/if}
	</div>

	<div class="flex-1 min-w-0">
		<MetricsChart data={history} {entityKey} {showReq} />
	</div>
</div>
