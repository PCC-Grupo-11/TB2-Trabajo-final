<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { auth } from '$lib/stores/auth';
	import { ApiRequestError } from '$lib/api/client';
	import Input from '$lib/components/Input.svelte';

	let username = $state('');
	let password = $state('');
	let confirmPassword = $state('');
	let loading = $state(false);
	let error = $state('');
	let success = $state('');

	onMount(() => {
		if (auth.isAuthenticated()) goto(resolve('/dashboard'));
	});

	async function handleSubmit(e: Event) {
		e.preventDefault();

		if (password !== confirmPassword) {
			error = 'Las contraseñas no coinciden';
			return;
		}

		loading = true;
		error = '';
		success = '';

		try {
			await auth.register(username, password);
			success = 'Cuenta creada. Redirigiendo al login...';
			setTimeout(() => goto(resolve('/login')), 1500);
		} catch (err) {
			if (err instanceof ApiRequestError) {
				error = err.message;
			} else {
				error = 'Error de conexion';
			}
		} finally {
			loading = false;
		}
	}
</script>

<svelte:head>
	<title>Registro</title>
</svelte:head>

<div class="flex min-h-screen items-center justify-center p-gutter">
	<main class="flex w-full max-w-sm flex-col items-center">
		<section
			aria-labelledby="register-heading"
			class="w-full rounded border border-outline-variant/30 bg-surface p-stack-lg shadow-sm"
		>
			<h2 class="sr-only" id="register-heading">Formulario de registro</h2>

			{#if error}
				<div class="mb-stack-md rounded border border-red-200 bg-red-50 p-3 text-sm text-red-700">
					{error}
				</div>
			{/if}

			{#if success}
				<div
					class="mb-stack-md rounded border border-green-200 bg-green-50 p-3 text-sm text-green-700"
				>
					{success}
				</div>
			{/if}

			<form onsubmit={handleSubmit} class="flex flex-col gap-stack-md">
				<div class="flex flex-col rounded shadow-sm">
					<label class="sr-only" for="username">Usuario</label>
					<Input
						position="first"
						bind:value={username}
						id="username"
						name="username"
						type="text"
						placeholder="Usuario"
						required
						class="border-outline-variant/50 bg-transparent px-4 py-3 text-on-surface placeholder:text-[0.8rem] placeholder:font-light placeholder:text-on-surface-variant/50 focus:outline-none focus:ring-inset focus:ring-2 focus:ring-black/80"
					/>
					<label class="sr-only" for="password">Contraseña</label>
					<Input
						position="middle"
						bind:value={password}
						id="password"
						name="password"
						type="password"
						placeholder="Contraseña"
						required
						class="border-outline-variant/50 bg-transparent px-4 py-3 text-on-surface placeholder:text-[0.8rem] placeholder:font-light placeholder:text-on-surface-variant/50 focus:outline-none focus:ring-inset focus:ring-2 focus:ring-black/80"
					/>
					<label class="sr-only" for="confirmPassword">Confirmar contraseña</label>
					<Input
						position="last"
						bind:value={confirmPassword}
						id="confirmPassword"
						name="confirmPassword"
						type="password"
						placeholder="Confirmar contraseña"
						required
						class="border-outline-variant/50 bg-transparent px-4 py-3 text-on-surface placeholder:text-[0.8rem] placeholder:font-light placeholder:text-on-surface-variant/50 focus:outline-none focus:ring-inset focus:ring-2 focus:ring-black/80"
					/>
				</div>

				<button
					type="submit"
					disabled={loading}
					class="mt-2 w-full cursor-pointer rounded bg-primary py-3 px-4 text-sm font-medium text-on-primary transition-colors hover:bg-inverse-surface focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary disabled:opacity-50"
				>
					{loading ? 'Creando cuenta...' : 'Crear cuenta'}
				</button>
			</form>

			<div class="mt-stack-md text-center">
				<a
					href={resolve('/login')}
					class="text-sm text-on-surface-variant transition-colors hover:text-primary"
				>
					Ya tengo una cuenta
				</a>
			</div>
		</section>
	</main>
</div>
