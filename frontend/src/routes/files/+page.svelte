<script lang="ts">
	import FSNodeItem from '$lib/components/FSNodeItem.svelte';
	import PageLayout from '$lib/components/PageLayout.svelte';
	import { createFilesState } from '$lib/state/files-state.svelte';
	import type { FSNode } from '$lib/types/types';
	import { onMount } from 'svelte';
	import { Button, clickOutsideAction, Dialog, Flexbox, Input } from 'svxui';

	const fileState = createFilesState();
	onMount(fileState.get);

	let current = $state<FSNode | undefined>();

	// new
	let newIsOpen = $state(false);
	let newCurrent = $state<FSNode | undefined>();
	let newDirname = $state<string>('');
	function newAction(fileNone: FSNode): void {
		newCurrent = fileNone;
		newIsOpen = true;
	}
	function newConfirm(): void {
		fileState.create(newCurrent?.path, newDirname);
		newCancel();
	}
	function newCancel(): void {
		newIsOpen = false;
		newCurrent = undefined;
	}

	// delete
	let deleteIsOpen = $state(false);
	let deleteCurrent = $state<FSNode | undefined>();
	function deleteAction(fileNone: FSNode): void {
		deleteCurrent = fileNone;
		deleteIsOpen = true;
	}
	function deleteConfirm(): void {
		fileState.delete(deleteCurrent?.path);
		deleteCancel();
	}
	function deleteCancel(): void {
		deleteIsOpen = false;
		deleteCurrent = undefined;
	}
</script>

<PageLayout title="Files" error={fileState.error}>
	<div
		use:clickOutsideAction
		onclickoutside={() => (current = undefined)}
		style="width: auto; overflow: auto"
	>
		{#if fileState.data}
			<FSNodeItem
				fileNode={fileState.data}
				fileNodeSelected={current}
				onSelect={(fileNode) => (current = fileNode)}
				onNew={newAction}
				onDelete={deleteAction}
			/>
		{/if}
	</div>
</PageLayout>

<Dialog bind:isOpen={newIsOpen} width="500px" maxWidth="90vw" closeOnBackdropClick closeOnEscape>
	<Flexbox direction="column" gap="3">
		<h2 class="my-0">Add new folder</h2>
		<div>
			<Input placeholder="new folder" bind:value={newDirname} />
		</div>
		<Flexbox gap="3" justify="end">
			<Button variant="outline" onclick={newCancel}>Cancel</Button>
			<Button onclick={newConfirm}>Confirm</Button>
		</Flexbox>
	</Flexbox>
</Dialog>

<Dialog bind:isOpen={deleteIsOpen} width="500px" maxWidth="90vw" closeOnBackdropClick closeOnEscape>
	<Flexbox direction="column" gap="3">
		<h2 class="my-0">Delete</h2>
		<div>
			Do you want delete {deleteCurrent?.name} item ?
		</div>
		<Flexbox gap="3" justify="end">
			<Button variant="outline" onclick={deleteCancel}>Cancel</Button>
			<Button onclick={deleteConfirm}>Confirm</Button>
		</Flexbox>
	</Flexbox>
</Dialog>
