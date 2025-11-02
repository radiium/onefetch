<script lang="ts">
	import PageLayout from '$lib/components/PageLayout.svelte';
	import SelectDirectory from '$lib/components/SelectDirectory.svelte';
	import { createNewState } from '$lib/state/new-state.svelte';
	import { DownloadType } from '$lib/types/types';
	import { useClipboard } from '$lib/utils/clipboard.svelte';
	import { formatBytes } from '$lib/utils/format-bytes';
	import Database from 'phosphor-svelte/lib/Database';
	import FileArrowDown from 'phosphor-svelte/lib/FileArrowDown';
	import Play from 'phosphor-svelte/lib/Play';
	import Folder from 'phosphor-svelte/lib/Folder';

	import { Button, Flexbox, Input, InputGroup, Panel, Select, Separator, Text } from 'svxui';

	const id = $props.id();
	const newState = createNewState();
	const clipboard = useClipboard();
</script>

<PageLayout title="New task">
	<Flexbox direction="column" gap="6">
		<!-- Url -->
		<Flexbox gap="4" align="center" as="label">
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
		</Flexbox>

		{#if newState.fileinfo?.url}
			<Panel variant="soft" size="0">
				<!-- Infos -->
				<Flexbox gap="4" align="stretch" class="p-5">
					<Panel size="2" style="width: 120px;" class="shrink-0 ">
						<Flexbox gap="3" align="center" justify="center" class="h-100">
							<Database class="shrink-0" />
							<Text muted weight="medium" wrap="nowrap" align="center" class="flex-auto">
								{formatBytes(newState.fileinfo.size)}
							</Text>
						</Flexbox>
					</Panel>

					<Panel size="2" class="flex-auto min-w-0">
						<Flexbox gap="3" align="center">
							<FileArrowDown class="shrink-0" />
							<Text
								muted
								weight="medium"
								wrap="pretty"
								class="min-w-0"
								title="path where the file will be saved">{newState.pathPreview}</Text
							>
						</Flexbox>
					</Panel>
				</Flexbox>
				<Separator size="4" />

				<Flexbox direction="column" gap="4" as="form" class="p-5">
					<!-- Type -->
					<Flexbox gap="4" align="center">
						<span>Type</span>

						<Select
							id="type-{id}"
							name="type"
							size="3"
							style="min-width: 150px;"
							class="flex-auto"
							options={Object.values(DownloadType)}
							bind:value={newState.type}
						/>
						<!-- <Flexbox gap="4" align="center">
							{#each Object.values(DownloadType) as type (type)}
								<Button
									size="3"
									variant={newState.type === type ? 'solid' : 'soft'}
									onclick={() => {
										newState.type = type;
									}}
								>
									{@const Icon = typeIcons[type]}
									<Icon  />
									{type}
								</Button>
							{/each}
						</Flexbox> -->
					</Flexbox>

					<!-- File dir -->
					<Flexbox gap="4" align="center" as="label">
						<span> Save to </span>

						<Flexbox gap="2" class="flex-auto">
							<Input
								id="fileDir-{id}"
								name="fileDir"
								fullWidth
								size="3"
								bind:value={newState.fileDir}
								disabled={newState.loading}
							/>

							<SelectDirectory
								options={newState.directories}
								disabled={newState.directories.length === 0}
								onSelect={(dir) => (newState.fileDir = dir)}
							>
								<Folder />
							</SelectDirectory>
						</Flexbox>
					</Flexbox>

					<!-- File name -->
					<Flexbox gap="4" align="center" as="label">
						<span> Rename </span>
						<Input
							id="fileName-{id}"
							name="fileName"
							size="3"
							fullWidth
							bind:value={newState.fileName}
							disabled={newState.loading}
						/>
					</Flexbox>
				</Flexbox>
			</Panel>

			<!-- Submit -->
			<Flexbox>
				<Button size="3" onclick={newState.create}>
					<Play weight="fill" />
					Start download
				</Button>
			</Flexbox>
		{/if}
	</Flexbox>
</PageLayout>

<style>
	span {
		min-width: 120px;
		text-align: right;
	}
</style>
