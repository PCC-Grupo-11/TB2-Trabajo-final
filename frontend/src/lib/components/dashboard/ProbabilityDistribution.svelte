<script lang="ts">
	import { TIME_BUCKET_SHORT, getSeverityColor } from '$lib/types/prediction';

	let { probabilities }: { probabilities: number[] } = $props();

	let maxProb = $derived(Math.max(...probabilities));
</script>

<div class="rounded border border-outline-variant/30 bg-result-card-bg p-stack-md">
	<h4 class="mb-3 text-label-sm font-medium uppercase tracking-wider text-on-surface-variant">
		Distribución de probabilidad
	</h4>
	<div class="flex flex-col gap-2.5">
		{#each probabilities as prob, i (i)}
			{@const pct = maxProb > 0 ? (prob / maxProb) * 100 : 0}
			<div class="flex items-center gap-3">
				<span class="w-20 shrink-0 text-right text-xs text-on-surface-variant">
					{TIME_BUCKET_SHORT[i]}
				</span>
				<div class="h-1.5 flex-1 overflow-hidden rounded-full bg-prob-bar-bg">
					<div
						class="h-full rounded-full {getSeverityColor(i)} transition-all duration-500"
						style="width: {pct}%"
					></div>
				</div>
				<span
					class="w-12 shrink-0 text-right font-mono text-xs tabular-nums text-on-surface-variant"
				>
					{(prob * 100).toFixed(1)}%
				</span>
			</div>
		{/each}
	</div>
</div>
