<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { auth } from '$lib/stores/auth';

	let checked = $state(false);

	onMount(() => {
		const unsubscribe = auth.subscribe((token) => {
			if (checked) return;

			checked = true;
			if (token) {
				goto(resolve('/dashboard'));
			} else {
				goto(resolve('/login'));
			}
		});
		return unsubscribe;
	});
</script>

{#if !checked}
	<div class="flex min-h-screen items-center justify-center">
		<div
			class="h-6 w-6 animate-spin rounded-full border-2 border-outline-variant border-t-primary"
		></div>
	</div>
{/if}
