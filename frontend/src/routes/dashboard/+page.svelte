<script lang="ts">
	import { auth } from '$lib/stores/auth';
</script>

<h1>Dashboard</h1>
<button onclick={() => auth.logout()}>Logout</button>
