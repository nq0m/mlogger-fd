<script>
	import '../app.css';

	let { children } = $props();
</script>

<div class="app-layout">
	{@render children()}
</div>

<style>
	.app-layout {
		display: flex;
		flex-direction: column;
		height: 100dvh;
		overflow: hidden;
	}
</style>
