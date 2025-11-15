<script lang="ts">
	import CaretDown from 'phosphor-svelte/lib/CaretDown';
	import Funnel from 'phosphor-svelte/lib/Funnel';
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
		name: string;
		value?: string[] | null;
		options: string[];
		floatingProps?: Partial<FloatingProps>;
		buttonOptionProps?: Partial<ButtonProps>;
		buttonProps?: Partial<ButtonProps>;
	};
	let {
		name,
		value = $bindable(),
		options = [],
		floatingProps,
		buttonOptionProps,
		buttonProps,
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
		<Button size="2" variant="outline" {...buttonProps} onclick={() => (isOpen = !isOpen)}>
			<Flexbox gap="2" align="center" class="pl-1">
				<Funnel size="1rem" weight={value?.length ? 'fill' : 'regular'} />

				{name}

				{#if valueCount > 0}
					<Badge size="1" variant="outline">{valueCount}</Badge>
				{/if}

				<CaretDown size="1.2rem" />
			</Flexbox>
		</Button>
	{/snippet}
	{#snippet content()}
		<Flexbox direction="column" align="center">
			{#each options as opt}
				<Button
					size="2"
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
				<Button size="2" variant="clear" fullWidth onclick={() => (value = [])}>
					Reset filter
				</Button>
			{/if}
		</Flexbox>
	{/snippet}
</Floating>
