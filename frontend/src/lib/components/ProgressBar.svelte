<script lang="ts">
	import type { Color } from 'svxui';

	interface Props {
		value: number;
		color?: Color;
	}

	let { value, color }: Props = $props();

	const percentage = $derived(Math.min(Math.max(value, 0), 100));
</script>

<div class="progressbar" data-color={color}>
	<div
		class="progress"
		role="progressbar"
		style:width="{percentage}%"
		aria-valuenow={percentage}
		aria-valuemin="0"
		aria-valuemax="100"
	></div>
</div>

<style>
	.progressbar {
		width: 100%;
		min-height: 8px;
		background: #2d2c2c;
		border-radius: var(--radius-3);
		overflow: hidden;
	}

	.progress {
		height: 100%;
		background-color: var(--accent-9);
		transition: width 0.3s ease;
	}
</style>
