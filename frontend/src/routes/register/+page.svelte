<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { auth } from '$lib/stores/auth';
	import { ApiRequestError } from '$lib/api/client';

	let username = $state('');
	let password = $state('');
	let confirmPassword = $state('');
	let loading = $state(false);
	let error = $state('');
	let success = $state('');

	async function handleSubmit(e: Event) {
		e.preventDefault();

		if (password !== confirmPassword) {
			error = 'Las contrasenas no coinciden';
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
		<header class="mb-stack-lg flex flex-col items-center justify-center">
			<p class="mt-1 text-center text-sm text-on-surface-variant">Crear una cuenta</p>
		</header>

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
					<div class="relative">
						<label class="sr-only" for="username">Usuario</label>
						<input
							bind:value={username}
							id="username"
							name="username"
							type="text"
							placeholder="Usuario"
							required
							class="relative block w-full rounded-t-lg border border-outline-variant/50 bg-transparent px-4 py-3 text-sm text-on-surface placeholder-on-surface-variant/60 focus:border-primary focus:outline-none focus:ring-1 focus:ring-primary"
						/>
					</div>

					<div class="relative">
						<label class="sr-only" for="password">Contraseña</label>
						<input
							bind:value={password}
							id="password"
							name="password"
							type="password"
							placeholder="Contraseña"
							required
							class="relative block w-full border border-t-0 border-outline-variant/50 bg-transparent px-4 py-3 text-sm text-on-surface placeholder-on-surface-variant/60 focus:border-primary focus:outline-none focus:ring-1 focus:ring-primary"
						/>
					</div>

					<div class="relative">
						<label class="sr-only" for="confirmPassword">Confirmar contraseña</label>
						<input
							bind:value={confirmPassword}
							id="confirmPassword"
							name="confirmPassword"
							type="password"
							placeholder="Confirmar contraseña"
							required
							class="relative block w-full rounded-b-lg border border-t-0 border-outline-variant/50 bg-transparent px-4 py-3 text-sm text-on-surface placeholder-on-surface-variant/60 focus:border-primary focus:outline-none focus:ring-1 focus:ring-primary"
						/>
					</div>
				</div>

				<button
					type="submit"
					disabled={loading}
					class="mt-2 w-full rounded bg-primary py-3 px-4 text-sm font-medium text-on-primary transition-colors hover:bg-inverse-surface focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary disabled:opacity-50"
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
