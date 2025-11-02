<script lang="ts">
	import CaretDown from 'phosphor-svelte/lib/CaretDown';
	import type { Snippet } from 'svelte';
	import {
		Badge,
		Button,
		Checkbox,
		Flexbox,
		Floating,
		Separator,
		type ButtonProps,
		type FloatingProps
	} from 'svxui';

	type Props = {
		value?: string[] | null;
		options: string[];
		floatingProps?: Partial<FloatingProps>;
		buttonOptionProps?: Partial<ButtonProps>;
		buttonProps?: Partial<ButtonProps>;
		buttonContent: Snippet;
	};
	let {
		value = $bindable(),
		options = [],
		floatingProps,
		buttonOptionProps,
		buttonProps,
		buttonContent
	}: Props = $props();

	let isOpen = $state(false);
	let valueCount = $derived(value?.length ?? 0);

	const select = (opt: string) => {
		if (!Array.isArray(value)) {
			value = [];
		}

		if (value.includes(opt)) {
			value = [...value.filter((v) => v !== opt)];
		} else {
			value = [...value, opt];
		}
	};
</script>

<Floating
	size="1"
	variant="outline"
	placement="bottom-start"
	offset={4}
	autoUpdate
	flip
	closeOnClickOutside
	{...floatingProps}
	bind:isOpen
>
	{#snippet trigger()}
		<Button size="3" variant="outline" {...buttonProps} onclick={() => (isOpen = !isOpen)}>
			<Flexbox gap="2" align="center" class="pl-1">
				{@render buttonContent?.()}
				{#if valueCount > 0}
					<Badge size="1" variant="outline">{valueCount}</Badge>
				{/if}

				<CaretDown />
			</Flexbox>
		</Button>
	{/snippet}
	{#snippet content()}
		<Flexbox direction="column" align="center">
			{#each options as opt}
				<Button
					size="3"
					variant="clear"
					align="start"
					fullWidth
					{...buttonOptionProps}
					onclick={() => select(opt)}
					style="--button-background-hover: var(--accent-5);"
				>
					<Flexbox gap="2" align="center">
						<Checkbox size="3" tabindex={-1} checked={value?.includes(opt)} />
						{opt}
					</Flexbox>
				</Button>
			{/each}

			{#if valueCount > 0}
				<Separator size="4" class="my-1" />
				<Button size="3" variant="clear" fullWidth onclick={() => (value = [])}>
					Reset filter
				</Button>
			{/if}
		</Flexbox>
	{/snippet}
</Floating>
