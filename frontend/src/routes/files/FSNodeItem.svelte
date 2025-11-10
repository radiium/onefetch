<script lang="ts">
	import type { FSNode } from '$lib/state/files-state.svelte';
	import { CaretDown, CaretRight, Folder, Plus, Trash } from 'phosphor-svelte';
	import { slide } from 'svelte/transition';
	import { AccordionItem, AccordionRoot, Button, Flexbox, Text } from 'svxui';
	import FSNodeItem from './FSNodeItem.svelte';
	import ContextMenu from './ContextMenu.svelte';

	type Props = {
		level?: number;
		fileNode: FSNode;
		fileNodeSelected?: FSNode;
		disabled?: boolean;
		onSelect?: (fileNode: FSNode) => void;
		onNew?: (fileNode: FSNode) => void;
		onDelete?: (fileNode: FSNode) => void;
	};
	let {
		level = 0,
		fileNode,
		fileNodeSelected,
		disabled = false,
		onSelect,
		onNew,
		onDelete
	}: Props = $props();

	let active = $derived(fileNode === fileNodeSelected);
	let offsetStyle = $derived(level > 0 ? `padding-left: calc(var(--space-6) * ${level})` : '');
	let value = $derived(level === 0 && fileNode ? fileNode.path : undefined);

	let contextMenu = $state<{ open: (evt: MouseEvent) => void; close: () => void } | undefined>(
		undefined
	);
	let oncontextmenu = $derived(!disabled && (onNew || onDelete) ? contextMenu?.open : undefined);
</script>

<ContextMenu bind:this={contextMenu}>
	<Flexbox direction="column">
		{#if fileNode.isDir}
			<Button
				size="2"
				variant="clear"
				align="start"
				onclick={() => {
					contextMenu?.close();
					onNew?.(fileNode);
				}}
			>
				<Plus />
				New folder
			</Button>
		{/if}

		<Button
			size="2"
			variant="clear"
			align="start"
			onclick={() => {
				contextMenu?.close();
				onDelete?.(fileNode);
			}}
		>
			<Trash />
			Delete
		</Button>
	</Flexbox>
</ContextMenu>

{#if fileNode?.isDir}
	<AccordionRoot orientation="vertical" bind:value>
		{#snippet children(root)}
			<Flexbox direction="column" class="w-100" {...root.rootAttrs}>
				<AccordionItem value={fileNode.path}>
					{#snippet children(item)}
						<!-- Item -->
						<Flexbox direction="column" class="w-100" {...item.itemAttrs}>
							<!-- Heading -->

							<Button
								variant="clear"
								align="start"
								fullWidth
								style={offsetStyle}
								{active}
								{oncontextmenu}
							>
								<Flexbox
									gap="1"
									align="center"
									justify="start"
									class="w-100 py-1"
									style="cursor: pointer; "
									{...item.headingAttrs}
								>
									<Flexbox align="center" {...item.triggerAttrs}>
										{#if item.expanded}
											<CaretDown size="1.2rem" />
										{:else}
											<CaretRight size="1.2rem" />
										{/if}
									</Flexbox>

									<Folder
										size="1.2rem"
										ondblclick={(e) => (item.triggerAttrs?.onclick as any)?.(e)}
										onclick={(e) => {
											onSelect?.(fileNode);
											e.stopPropagation();
										}}
									/>
									<Text
										class="shrink-0 flex-auto"
										truncate
										ondblclick={(e: MouseEvent) => (item.triggerAttrs?.onclick as any)?.(e)}
										onclick={(e: MouseEvent) => {
											onSelect?.(fileNode);
											e.stopPropagation();
										}}
									>
										{fileNode?.name}
									</Text>
								</Flexbox>
							</Button>

							<!-- Content -->
							{#if item.expanded}
								<div transition:slide={{ duration: 150 }} {...item.contentAttrs}>
									{#each fileNode.children as childFileNode}
										<FSNodeItem
											{fileNodeSelected}
											level={level + 1}
											fileNode={childFileNode}
											{disabled}
											{onSelect}
											{onNew}
											{onDelete}
										/>
									{/each}
								</div>
							{/if}
						</Flexbox>
					{/snippet}
				</AccordionItem>
			</Flexbox>
		{/snippet}
	</AccordionRoot>
{:else if fileNode}
	<Button
		variant="clear"
		align="start"
		fullWidth
		style={offsetStyle}
		{active}
		onclick={() => onSelect?.(fileNode)}
		{oncontextmenu}
	>
		<Text truncate class="flex-auto">{fileNode?.name}</Text>
	</Button>
{/if}
