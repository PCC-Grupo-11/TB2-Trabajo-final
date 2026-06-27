<script lang="ts">
	import type { HTMLInputAttributes } from 'svelte/elements';

	const defaultClass =
		'border-outline-variant/50 bg-transparent px-4 py-3 text-on-surface placeholder:text-[0.8rem] placeholder:font-light placeholder:text-on-surface-variant/50 focus:outline-none focus:ring-inset focus:ring-2 focus:ring-black/80';

	let {
		position,
		class: className = defaultClass,
		value = $bindable(),
		...restProps
	}: HTMLInputAttributes & { position?: 'first' | 'middle' | 'last' } = $props();

	const positionClasses: Record<string, string> = {
		first: 'rounded-t-lg',
		middle: 'rounded-none border-t-0',
		last: 'rounded-b-lg border-t-0'
	};
</script>

<input
	bind:value
	class="w-full {positionClasses[position ?? ''] ?? ''} {className}"
	{...restProps}
/>
