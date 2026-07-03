<script lang="ts">
	import { metrics } from '$lib/stores/metrics.svelte';
	import X from '@lucide/svelte/icons/x';

	let { open, onclose }: { open: boolean; onclose: () => void } = $props();

	function formatBytes(bytes: number): string {
		if (bytes === 0) return '0 B';
		const units = ['B', 'KB', 'MB', 'GB'];
		const i = Math.floor(Math.log(bytes) / Math.log(1024));
		return `${(bytes / Math.pow(1024, i)).toFixed(2)} ${units[i]}`;
	}

	function formatPercent(val: number): string {
		return `${val.toFixed(2)}%`;
	}
</script>

{#if open}
	<div class="fixed inset-0 z-40 bg-black/20" onclick={onclose} role="presentation"></div>
{/if}

<aside
	class="fixed top-12 bottom-0 left-0 z-50 w-80 overflow-y-auto border-r border-outline-variant/30 bg-surface shadow-lg transition-transform duration-300 ease-in-out
		{open ? 'translate-x-0' : '-translate-x-full'}"
>
	<div class="flex items-center justify-between border-b border-outline-variant/30 px-4 py-3">
		<h2 class="text-title-md font-medium text-on-surface">Métricas del cluster</h2>
		<button
			onclick={onclose}
			class="cursor-pointer text-on-surface-variant transition-colors hover:text-on-surface"
		>
			<X size={20} strokeWidth={1.5} />
		</button>
	</div>

	{#if !metrics.connected}
		<div class="p-4 text-body-md text-on-surface-variant">Conectando...</div>
	{:else if !metrics.info || !metrics.metrics}
		<div class="p-4 text-body-md text-on-surface-variant">Esperando datos...</div>
	{:else}
		<!-- API Node -->
		<div class="border-b border-outline-variant/30 px-4 py-3">
			<h3 class="text-label-md font-medium text-on-surface-variant">API</h3>
			<div class="mt-2 space-y-1.5">
				<div class="flex justify-between text-body-sm">
					<span class="text-on-surface-variant">CPU</span>
					<span class="font-mono text-on-surface"
						>{formatPercent(metrics.metrics.api.cpu_percent)}</span
					>
				</div>
				<div class="h-1.5 w-full rounded-full bg-outline-variant/30">
					<div
						class="h-1.5 rounded-full bg-primary transition-all duration-500"
						style="width: {Math.min(metrics.metrics.api.cpu_percent, 100)}%"
					></div>
				</div>
				<div class="flex justify-between text-body-sm">
					<span class="text-on-surface-variant">Memoria</span>
					<span class="font-mono text-on-surface"
						>{formatBytes(metrics.metrics.api.memory_bytes)}</span
					>
				</div>
				<div class="flex justify-between text-body-sm">
					<span class="text-on-surface-variant">CPU</span>
					<span class="font-mono text-on-surface">{metrics.info.api.cpu_name}</span>
				</div>
				<div class="flex justify-between text-body-sm">
					<span class="text-on-surface-variant">Núcleos</span>
					<span class="font-mono text-on-surface">{metrics.info.api.cores}</span>
				</div>
				<div class="flex justify-between text-body-sm">
					<span class="text-on-surface-variant">RAM</span>
					<span class="font-mono text-on-surface">{metrics.info.api.total_ram_gb} GB</span>
				</div>
			</div>
		</div>

		<!-- Inference Nodes -->
		<div class="border-b border-outline-variant/30 px-4 py-3">
			<h3 class="text-label-md font-medium text-on-surface-variant">
				Nodos de Inferencia ({metrics.metrics.cluster.healthy_nodes}/{metrics.metrics.cluster
					.total_nodes})
			</h3>
			<div class="mt-2 space-y-3">
				{#each metrics.metrics.nodes as node (node.addr)}
					{@const info = metrics.info.nodes.find((n) => n.addr === node.addr)}
					<div class="rounded-lg border border-outline-variant/30 p-2.5">
						<div class="flex items-center justify-between">
							<span class="text-body-sm font-medium text-on-surface">{node.addr}</span>
							<span
								class="rounded-full px-2 py-0.5 text-label-sm
									{node.status === 'healthy' ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'}"
							>
								{node.status === 'healthy' ? 'Saludable' : 'Inalcanzable'}
							</span>
						</div>
						{#if node.status === 'healthy'}
							<div class="mt-2 space-y-1">
								<div class="flex justify-between text-body-sm">
									<span class="text-on-surface-variant">CPU</span>
									<span class="font-mono text-on-surface">{formatPercent(node.cpu_percent)}</span>
								</div>
								<div class="h-1.5 w-full rounded-full bg-outline-variant/30">
									<div
										class="h-1.5 rounded-full bg-primary transition-all duration-500"
										style="width: {Math.min(node.cpu_percent, 100)}%"
									></div>
								</div>
								<div class="flex justify-between text-body-sm">
									<span class="text-on-surface-variant">Memoria</span>
									<span class="font-mono text-on-surface">{formatBytes(node.memory_bytes)}</span>
								</div>
								<div class="flex justify-between text-body-sm">
									<span class="text-on-surface-variant">Peticiones</span>
									<span class="font-mono text-on-surface">{node.requests_served}</span>
								</div>
								{#if info}
									<div class="flex justify-between text-body-sm">
										<span class="text-on-surface-variant">CPU</span>
										<span class="font-mono text-on-surface">{info.cpu_name}</span>
									</div>
									<div class="flex justify-between text-body-sm">
										<span class="text-on-surface-variant">Núcleos</span>
										<span class="font-mono text-on-surface">{info.cores}</span>
									</div>
									<div class="flex justify-between text-body-sm">
										<span class="text-on-surface-variant">RAM</span>
										<span class="font-mono text-on-surface">{info.total_ram_gb} GB</span>
									</div>
								{/if}
							</div>
						{/if}
					</div>
				{/each}
			</div>
		</div>

		<!-- Cluster -->
		<div class="px-4 py-3">
			<h3 class="text-label-md font-medium text-on-surface-variant">Cluster</h3>
			<div class="mt-2 space-y-1.5">
				<div class="flex justify-between text-body-sm">
					<span class="text-on-surface-variant">Predicciones totales</span>
					<span class="font-mono text-on-surface">{metrics.metrics.cluster.total_predictions}</span>
				</div>
				<div class="flex justify-between text-body-sm">
					<span class="text-on-surface-variant">Latencia promedio</span>
					<span class="font-mono text-on-surface"
						>{metrics.metrics.cluster.avg_latency_ms.toFixed(2)} ms</span
					>
				</div>
				<div class="flex justify-between text-body-sm">
					<span class="text-on-surface-variant">Cache hit rate</span>
					<span class="font-mono text-on-surface"
						>{(metrics.metrics.cluster.cache_hit_rate * 100).toFixed(2)}%</span
					>
				</div>
			</div>
		</div>
	{/if}
</aside>
