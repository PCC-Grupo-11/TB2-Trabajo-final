<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { auth } from '$lib/stores/auth';
	import { ApiRequestError } from '$lib/api/client';
	import Input from '$lib/components/Input.svelte';

	let username = $state('');
	let password = $state('');
	let loading = $state(false);
	let error = $state('');

	onMount(() => {
		if (auth.isAuthenticated()) goto(resolve('/dashboard'));
	});

	async function handleSubmit(e: Event) {
		e.preventDefault();
		loading = true;

		try {
			await auth.login(username, password);
			goto(resolve('/dashboard'));
		} catch (err) {
			if (err instanceof ApiRequestError && err.status === 401) {
				error = 'El usuario y la contraseña no coinciden.';
			} else if (err instanceof ApiRequestError && err.status === 400) {
				error = 'Solicitud invalida.';
			} else {
				error = 'Error de conexion.';
			}
		} finally {
			loading = false;
		}
	}
</script>

<svelte:head>
	<title>Login</title>
</svelte:head>

<div class="flex min-h-screen items-center justify-center p-gutter">
	<main class="flex w-full max-w-sm flex-col items-center">
		<section
			aria-labelledby="login-heading"
			class="w-full rounded border border-outline-variant/30 bg-surface p-stack-lg shadow-sm"
		>
			<h2 class="sr-only" id="login-heading">Formulario de Login</h2>

			{#if error}
				<div class="mb-stack-md rounded border border-red-200 bg-red-50 p-3 text-sm text-red-700">
					{error}
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
					/>
					<label class="sr-only" for="password">Contraseña</label>
					<Input
						position="last"
						bind:value={password}
						id="password"
						name="password"
						type="password"
						placeholder="Contraseña"
						required
					/>
				</div>

				<button
					type="submit"
					disabled={loading}
					class="mt-2 w-full cursor-pointer rounded bg-primary py-3 px-4 text-sm font-medium text-on-primary transition-colors hover:bg-inverse-surface focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary disabled:opacity-50"
				>
					{loading ? 'Ingresando...' : 'Iniciar sesión'}
				</button>
			</form>

			<div class="mt-stack-md text-center">
				<a
					href={resolve('/register')}
					class="text-sm text-on-surface-variant transition-colors hover:text-primary"
				>
					Crear cuenta
				</a>
			</div>
		</section>
	</main>
</div>
