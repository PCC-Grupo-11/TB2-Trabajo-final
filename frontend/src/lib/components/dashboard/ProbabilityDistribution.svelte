<script lang="ts">
	import { TIME_BUCKET_SHORT, getSeverityColor } from '$lib/types/prediction';

	let { probabilities, winnerIndex }: { probabilities: number[]; winnerIndex: number } = $props();

	let maxProb = $derived(Math.max(...probabilities));
</script>

<div>
	<h4 class="mb-3 text-[12px] font-semibold uppercase tracking-[0.12em] text-on-surface-variant">
		Distribución de probabilidad
	</h4>
	<div class="flex flex-col gap-3">
		{#each probabilities as prob, i (i)}
			{@const pct = maxProb > 0 ? (prob / maxProb) * 100 : 0}
			<div class="flex items-center gap-3">
				<span
					class="w-20 shrink-0 text-right text-sm {i === winnerIndex
						? 'font-semibold text-on-surface'
						: 'text-on-surface-variant'}"
				>
					{TIME_BUCKET_SHORT[i]}
				</span>
				<div class="h-2.5 flex-1 overflow-hidden rounded-full bg-prob-bar-bg">
					<div
						class="h-full rounded-full {getSeverityColor(i)} transition-all duration-500"
						style="width: {pct}%"
					></div>
				</div>
				<span
					class="w-12 shrink-0 text-right font-mono text-sm tabular-nums {i === winnerIndex
						? 'font-semibold text-on-surface'
						: 'text-on-surface-variant'}"
				>
					{(prob * 100).toFixed(1)}%
				</span>
			</div>
		{/each}
	</div>
</div>
