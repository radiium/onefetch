<script lang="ts">
	import { type Snippet } from 'svelte';
	import { fade } from 'svelte/transition';
	import { clickoutside, flip, FloatingEngine, offset, Panel, shift } from 'svxui';

	// Props
	type Props = {
		children?: Snippet<[{ close: () => void }]>;
	};
	let { children }: Props = $props();

	let isOpen = $state<boolean>(false);
	const floating = new FloatingEngine({
		transform: true,
		strategy: 'fixed',
		placement: 'bottom-start',
		middleware: [offset(), flip(), shift()]
	});

	export function open(evt: MouseEvent): void {
		evt.preventDefault();

		floating.reference = {
			getBoundingClientRect() {
				return {
					width: 0,
					height: 0,
					x: evt.clientX,
					y: evt.clientY,
					top: evt.clientY,
					left: evt.clientX,
					right: evt.clientX,
					bottom: evt.clientY
				};
			}
		};

		isOpen = true;
	}

	export function close() {
		isOpen = false;
	}

	function handleKeydown(evt: KeyboardEvent) {
		if (evt.key === 'Escape') {
			close();
		}
	}

	function handleClickOutside() {
		close();
	}
</script>

<!-- Handle globals events -->
<svelte:window onkeydown={handleKeydown} />

{#if isOpen}
	<div
		transition:fade={{ duration: 150, delay: 0 }}
		bind:this={floating.floating}
		{@attach clickoutside({ onClickOutside: handleClickOutside })}
		style={floating.style}
		class="context-menu"
		data-state={isOpen ? 'open' : 'close'}
		role="dialog"
	>
		<Panel size="2" variant="soft" outline>
			{@render children?.({ close })}
		</Panel>
	</div>
{/if}

<style>
	.context-menu {
		position: fixed;
		width: -moz-max-content;
		width: max-content;
		top: 0;
		left: 0;
		z-index: 1;

		&[data-state='open'] {
			display: block;
		}

		&[data-state='close'] {
			display: none;
		}
	}
</style>
