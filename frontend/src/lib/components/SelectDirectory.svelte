<script lang="ts">
	import { XIcon } from 'phosphor-svelte';
	import { type Snippet } from 'svelte';
	import { Button, Dialog, Flex, Input, Panel, Separator, Text } from 'svxui';

	type Props = {
		options: string[];
		disabled?: boolean;
		onSelect?: (value: string) => void;
		children: Snippet;
	};
	let { options = [], disabled, onSelect, children }: Props = $props();

	let isOpen: boolean = $state(false);
	let query: string = $state('');
	let inputEl: HTMLInputElement | undefined = $state();
	let filteredOptions: string[] = $derived(
		query.trim()
			? options.filter((opt) => opt.toLowerCase().includes(query.trim().toLowerCase()))
			: options
	);
	const open = async () => {
		isOpen = true;
		query = '';
		setTimeout(() => inputEl?.focus(), 800);
	};
	const close = () => {
		isOpen = false;
		query = '';
	};
</script>

<Button size="3" iconOnly variant="outline" onclick={open} {disabled}>
	{@render children?.()}
</Button>

<Dialog closeOnBackdropClick closeOnEscape bind:isOpen onClose={close} style="margin-top: 10%">
	<Panel variant="soft" outline>
		<Flex direction="column" gap="3" width="350px" maxHeight="70vh">
			<Flex align="center" gap="3">
				<Input
					id="search-directory"
					type="text"
					placeholder="Search directory..."
					fullWidth
					bind:value={query}
					bind:ref={inputEl}
				/>
				<Button iconOnly onclick={close}>
					<XIcon />
				</Button>
			</Flex>

			<Separator size="4" />
			<Flex direction="column" gap="1" overflow="auto">
				{#if filteredOptions.length === 0}
					<Text wrap="nowrap">No result found...</Text>
				{/if}
				{#each filteredOptions as opt, i (i)}
					<Button
						variant="clear"
						align="start"
						fullWidth
						onclick={() => {
							close();
							onSelect?.(opt);
						}}
						style="--button-background-hover: var(--accent-5);"
					>
						<Text wrap="nowrap">{opt}</Text>
					</Button>
				{/each}
			</Flex>
		</Flex>
	</Panel>
</Dialog>
