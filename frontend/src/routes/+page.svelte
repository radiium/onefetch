<script lang="ts">
	import PageLayout from '$lib/components/PageLayout.svelte';
	import SelectDirectory from '$lib/components/SelectDirectory.svelte';
	import { createNewState } from '$lib/state/new-state.svelte';
	import { DownloadType } from '$lib/types/types';
	import { useClipboard } from '$lib/utils/clipboard.svelte';
	import { formatBytes } from '$lib/utils/format-bytes';
	import DatabaseIcon from 'phosphor-svelte/lib/DatabaseIcon';
	import FileArrowDownIcon from 'phosphor-svelte/lib/FileArrowDownIcon';
	import FolderIcon from 'phosphor-svelte/lib/FolderIcon';
	import PlayIcon from 'phosphor-svelte/lib/PlayIcon';

	import { Button, Flex, Input, Panel, Select, SelectOption, Separator, Text } from 'svxui';

	const id = $props.id();
	const newState = createNewState();
	const clipboard = useClipboard();
</script>

<PageLayout title="New task" error={newState.error}>
	<Flex direction="column" gap="6">
		<!-- Url -->
		<Flex gap="4" align="center" as="label">
			<Input
				id="url-{id}"
				name="url"
				size="3"
				fullWidth
				placeholder="Type 1fichier.com URL..."
				bind:value={newState.url}
				onfocus={async () => {
					const url = ((await clipboard.read()) ?? '').trim();
					if (newState.isValid1FichierUrl(url)) {
						newState.url = url;
					}
				}}
			/>
		</Flex>

		{#if newState.fileinfo?.url}
			<Panel variant="soft" p="0">
				<!-- Infos -->
				<Flex gap="4" align="stretch" class="p-5">
					<Panel p="2" style="width: 120px;" class="shrink-0 ">
						<Flex gap="3" align="center" justify="center" class="h-100">
							<DatabaseIcon class="shrink-0" />
							<Text muted weight="medium" wrap="nowrap" align="center" class="flex-auto">
								{formatBytes(newState.fileinfo.size)}
							</Text>
						</Flex>
					</Panel>

					<Panel p="2" class="flex-auto min-w-0">
						<Flex gap="3" align="center">
							<FileArrowDownIcon class="shrink-0" />
							<Text
								muted
								weight="medium"
								wrap="pretty"
								class="min-w-0"
								title="path where the file will be saved">{newState.pathPreview}</Text
							>
						</Flex>
					</Panel>
				</Flex>

				<Separator size="4" />

				<Flex direction="column" gap="4" as="form" class="p-5">
					<!-- Type -->
					<Flex gap="4" align="center">
						<span>Type</span>

						<Select
							id="type-{id}"
							name="type"
							size="3"
							style="min-width: 150px;"
							class="flex-auto"
							bind:value={newState.type}
						>
							{#each Object.values(DownloadType) as value, i (i)}
								<SelectOption {value}>{value}</SelectOption>
							{/each}
						</Select>
					</Flex>

					<!-- File dir -->
					<Flex gap="4" align="center" as="label">
						<span> Save to </span>

						<Flex gap="2" class="flex-auto">
							<Input
								id="fileDir-{id}"
								name="fileDir"
								fullWidth
								size="3"
								bind:value={newState.fileDir}
								disabled={newState.loading}
							/>

							{#if newState.directories.length}
								<SelectDirectory
									options={newState.directories}
									disabled={newState.directories.length === 0}
									onSelect={(dir) => (newState.fileDir = dir)}
								>
									<FolderIcon />
								</SelectDirectory>
							{/if}
						</Flex>
					</Flex>

					<!-- File name -->
					<Flex gap="4" align="center" as="label">
						<span> Rename </span>
						<Input
							id="fileName-{id}"
							name="fileName"
							size="3"
							fullWidth
							bind:value={newState.fileName}
							disabled={newState.loading}
						/>
					</Flex>
				</Flex>
			</Panel>

			<!-- Submit -->
			<Flex>
				<Button size="3" onclick={newState.create}>
					<PlayIcon weight="fill" />
					Start download
				</Button>
			</Flex>
		{/if}
	</Flex>
</PageLayout>

<style>
	span {
		min-width: 120px;
		text-align: right;
	}
</style>
