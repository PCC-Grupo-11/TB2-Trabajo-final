<script lang="ts">
	import { Chart, Svg, Area } from 'layerchart';
	import type { HistoryPoint } from '$lib/types/metrics';

	let {
		data,
		entityKey,
		showReq = false
	}: { data: HistoryPoint[]; entityKey: string; showReq?: boolean } = $props();

	let cpuData = $derived(
		data.map((p) => ({
			ts: p.ts.getTime(),
			value: entityKey === 'api' ? p.api.cpu : (p.nodes[entityKey]?.cpu ?? 0)
		}))
	);

	let memData = $derived(
		data.map((p) => ({
			ts: p.ts.getTime(),
			value:
				entityKey === 'api'
					? p.api.mem / (1024 * 1024)
					: (p.nodes[entityKey]?.mem ?? 0) / (1024 * 1024)
		}))
	);

	let reqData = $derived(
		showReq
			? data.map((p, i) => {
					const prev = i > 0 ? data[i - 1] : null;
					const curr = p.nodes[entityKey]?.req ?? 0;
					const prevVal = prev?.nodes[entityKey]?.req ?? curr;
					const dt = prev ? (p.ts.getTime() - prev.ts.getTime()) / 1000 : 1;
					return { ts: p.ts.getTime(), value: dt > 0 ? Math.max(0, (curr - prevVal) / dt) : 0 };
				})
			: []
	);
</script>

{#if cpuData.length > 1}
	<div class="flex gap-3">
		<div class="flex-1 min-w-0">
			<div class="text-[11px] uppercase tracking-[0.12em] text-on-surface-variant/60 mb-1">
				CPU %
			</div>
			<div class="h-20">
				<Chart
					data={cpuData}
					x="ts"
					y="value"
					yNice
					padding={{ left: 0, right: 0, top: 4, bottom: 4 }}
				>
					<Svg>
						<defs>
							<linearGradient id="grad-cpu-{entityKey}" x1="0" y1="0" x2="0" y2="1">
								<stop offset="0%" stop-color="#f59e0b" stop-opacity="0.35" />
								<stop offset="100%" stop-color="#f59e0b" stop-opacity="0.04" />
							</linearGradient>
						</defs>
						<Area
							fill="url(#grad-cpu-{entityKey})"
							line={{ class: 'stroke-amber-500 stroke-[1.5] fill-none' }}
						/>
					</Svg>
				</Chart>
			</div>
		</div>

		<div class="flex-1 min-w-0">
			<div class="text-[11px] uppercase tracking-[0.12em] text-on-surface-variant/60 mb-1">
				RAM (MB)
			</div>
			<div class="h-20">
				<Chart
					data={memData}
					x="ts"
					y="value"
					yNice
					padding={{ left: 0, right: 0, top: 4, bottom: 4 }}
				>
					<Svg>
						<defs>
							<linearGradient id="grad-mem-{entityKey}" x1="0" y1="0" x2="0" y2="1">
								<stop offset="0%" stop-color="#3b82f6" stop-opacity="0.4" />
								<stop offset="100%" stop-color="#3b82f6" stop-opacity="0.05" />
							</linearGradient>
						</defs>
						<Area
							fill="url(#grad-mem-{entityKey})"
							line={{ class: 'stroke-blue-500 stroke-[1.5] fill-none' }}
						/>
					</Svg>
				</Chart>
			</div>
		</div>
	</div>

	{#if showReq && reqData.length > 1}
		<div class="mt-2">
			<div class="text-[11px] uppercase tracking-[0.12em] text-on-surface-variant/60 mb-1">
				Peticiones/s
			</div>
			<div class="h-10">
				<Chart
					data={reqData}
					x="ts"
					y="value"
					yNice
					padding={{ left: 0, right: 0, top: 4, bottom: 4 }}
				>
					<Svg>
						<defs>
							<linearGradient id="grad-req-{entityKey}" x1="0" y1="0" x2="0" y2="1">
								<stop offset="0%" stop-color="#22c55e" stop-opacity="0.35" />
								<stop offset="100%" stop-color="#22c55e" stop-opacity="0.04" />
							</linearGradient>
						</defs>
						<Area
							fill="url(#grad-req-{entityKey})"
							line={{ class: 'stroke-green-500 stroke-[1.5] fill-none' }}
						/>
					</Svg>
				</Chart>
			</div>
		</div>
	{/if}
{:else}
	<div class="h-20 flex items-center justify-center text-[11px] text-on-surface-variant/40">
		Esperando datos...
	</div>
{/if}
