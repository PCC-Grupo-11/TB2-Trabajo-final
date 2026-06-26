<script lang="ts">
	import type { HTMLInputAttributes } from 'svelte/elements';

	let {
		position,
		class: className,
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
	class="w-full {positionClasses[position ?? ''] ?? ''} {className ?? ''}"
	{...restProps}
/>
