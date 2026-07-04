<script lang="ts">
	import { metrics } from '$lib/stores/metrics.svelte';
	import NodeMetricsBlock from './NodeMetricsBlock.svelte';

	let { open, onclose }: { open: boolean; onclose: () => void } = $props();
</script>

{#if open}
	<div
		class="fixed top-12 inset-x-0 bottom-0 z-40 bg-black/20"
		onclick={onclose}
		role="presentation"
	></div>
{/if}

<aside
	class="fixed top-12 bottom-0 left-0 z-50 w-md overflow-y-auto border-r border-outline-variant/30 bg-surface shadow-lg transition-transform duration-300 ease-in-out
		{open ? 'translate-x-0' : '-translate-x-full'}"
>
	<div class="px-stack-lg py-stack-md">
		<h2 class="text-[13px] font-semibold uppercase tracking-[0.12em] text-on-surface-variant">
			Métricas del backend
		</h2>
	</div>

	<div class="mx-6 border-t border-outline-variant/20"></div>

	{#if !metrics.connected}
		<div class="px-stack-lg py-stack-md text-[13px] text-on-surface-variant">Conectando...</div>
	{:else if !metrics.info || !metrics.metrics}
		<div class="px-stack-lg py-stack-md text-[13px] text-on-surface-variant">
			Esperando datos...
		</div>
	{:else}
		<!-- API -->
		<div class="px-stack-lg pt-5 pb-stack-md">
			<h3 class="text-[13px] font-semibold text-on-surface">API</h3>
			<p class="mt-0.5 text-[11px] text-on-surface-variant/60">
				{metrics.info.api.cpu_name} · {metrics.info.api.cores} cores · {metrics.info.api.total_ram_gb.toFixed(
					2
				)} GB RAM
			</p>

			<div class="mt-4">
				<NodeMetricsBlock
					cpuPercent={metrics.metrics.api.cpu_percent}
					memoryBytes={metrics.metrics.api.memory_bytes}
					history={metrics.history}
					entityKey="api"
				/>
			</div>
		</div>

		<div class="mx-6 border-t border-outline-variant/20"></div>

		<!-- Inference Nodes -->
		<div class="px-stack-lg pt-5 pb-stack-md">
			<h3 class="text-[12px] font-semibold uppercase tracking-[0.12em] text-on-surface-variant">
				Nodos de inferencia ({metrics.metrics.cluster.healthy_nodes}/{metrics.metrics.cluster
					.total_nodes})
			</h3>

			{#each metrics.metrics.nodes as node, i (node.addr)}
				{@const info = metrics.info.nodes.find((n) => n.addr === node.addr)}

				{#if i > 0}
					<div class="my-4 border-t border-outline-variant/10"></div>
				{/if}

				<div class={i === 0 ? 'mt-4' : ''}>
					<div class="flex items-center gap-2">
						<span class="text-[13px] font-semibold text-on-surface">{node.addr}</span>
						<span
							class="rounded-full px-2 py-0.5 text-[10px] font-medium
								{node.status === 'healthy' ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'}"
						>
							{node.status === 'healthy' ? 'Saludable' : 'Inalcanzable'}
						</span>
					</div>

					{#if info}
						<p class="mt-0.5 text-[11px] text-on-surface-variant/60">
							{info.cpu_name} · {info.cores} cores · {info.total_ram_gb.toFixed(2)} GB RAM
						</p>
					{/if}

					{#if node.status === 'healthy'}
						<div class="mt-4">
							<NodeMetricsBlock
								cpuPercent={node.cpu_percent}
								memoryBytes={node.memory_bytes}
								history={metrics.history}
								entityKey={node.addr}
								peticiones={node.requests_served}
								showReq
							/>
						</div>
					{:else}
						<p class="mt-2 text-[11px] text-on-surface-variant/40">Nodo inalcanzable</p>
					{/if}
				</div>
			{/each}
		</div>

		<div class="mx-6 border-t border-outline-variant/20"></div>

		<!-- Cluster -->
		<div class="px-stack-lg py-stack-md">
			<h3 class="text-[12px] font-semibold uppercase tracking-[0.12em] text-on-surface-variant">
				Cluster
			</h3>
			<div class="mt-3 space-y-2">
				<div class="flex items-center justify-between">
					<span
						class="text-[11px] font-medium uppercase tracking-[0.12em] text-on-surface-variant/70"
						>Predicciones totales</span
					>
					<span class="font-mono text-[13px] text-on-surface"
						>{metrics.metrics.cluster.total_predictions}</span
					>
				</div>
				<div class="flex items-center justify-between">
					<span
						class="text-[11px] font-medium uppercase tracking-[0.12em] text-on-surface-variant/70"
						>Latencia promedio</span
					>
					<span class="font-mono text-[13px] text-on-surface"
						>{metrics.metrics.cluster.avg_latency_ms.toFixed(3)} ms</span
					>
				</div>
				<div class="flex items-center justify-between">
					<span
						class="text-[11px] font-medium uppercase tracking-[0.12em] text-on-surface-variant/70"
						>Cache hit rate</span
					>
					<span class="font-mono text-[13px] text-on-surface"
						>{(metrics.metrics.cluster.cache_hit_rate * 100).toFixed(2)}%</span
					>
				</div>
			</div>
		</div>
	{/if}
</aside>
