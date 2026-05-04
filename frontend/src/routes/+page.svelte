<script lang="ts">
	import PageLayout from '$lib/components/PageLayout.svelte';
	import SelectDirectory from '$lib/components/SelectDirectory.svelte';
	import { createNewState } from '$lib/state/new-state.svelte';
	import { DownloadType } from '$lib/types/types';
	import { formatBytes } from '$lib/utils/format-bytes';
	import { typeIcons } from '$lib/utils/type-icons';
	import { FileCloudIcon, FilmReelIcon } from 'phosphor-svelte';
	import FolderIcon from 'phosphor-svelte/lib/FolderIcon';
	import PlayIcon from 'phosphor-svelte/lib/PlayIcon';
	import { Button, Clipboard, Flex, Grid, Input, Panel, Separator, Text } from 'svxui';

	const id = $props.id();
	const newState = createNewState();
	const clipboard = new Clipboard();
</script>

<PageLayout title="New download" error={newState.error}>
	<Flex direction="column" gap="4">
		<!-- Step 1 -->
		<Panel variant="surface" p="5">
			<Flex direction="column" gap="4">
				<!-- Input URL -->
				<Input
					id="url-{id}"
					name="url"
					size="3"
					fullWidth
					placeholder="paste a 1fichier.com link..."
					bind:value={newState.url}
					onfocus={async () => {
						clipboard.read();
						const url = ((await clipboard.read()) ?? '').trim();
						if (newState.isValid1FichierUrl(url)) {
							newState.url = url;
						}
					}}
					style="padding: var(--space-4) var(--space-4); font-size: var(--font-size-6)"
				/>

				<!-- Download infos -->
				{#if newState.fileinfo?.url}
					<Panel variant="soft">
						<Flex align="center" gap="3">
							<Panel variant="solid" p="3">
								<Flex align="center">
									<FilmReelIcon size="30px" />
								</Flex>
							</Panel>
							<Flex direction="column" gap="1">
								<Text weight="medium">{newState.fileinfo.filename}</Text>
								<Flex gap="1" align="center">
									<Text size="2" muted>{formatBytes(newState.fileinfo.size)} • 1fichier •</Text>
									<Text size="2" color="green">✓ resolved</Text>
								</Flex>
							</Flex>
						</Flex>
					</Panel>
				{/if}
			</Flex>
		</Panel>

		<!-- Step 2 -->
		{#if newState.fileinfo?.url}
			<Panel variant="surface" p="5">
				<Flex gap="4" direction="column" as="form">
					<!-- type -->
					<Grid cols="80px 1fr" gap="3" as="label">
						<Flex justify="start" align="center" as="span">Type</Flex>

						<Flex gap="3">
							{#each Object.values(DownloadType) as value, i (i)}
								{@const Icon = typeIcons[value]}
								{@const isSelected = newState.type === value}

								<Panel as="label" variant={isSelected ? 'soft' : 'surface'} outline p="0">
									<!-- color={isSelected ? 'orange' : 'neutral'} -->
									<Flex align="center" gap="3" px="3" height="40px">
										<Icon size="16px" />
										<Flex direction="column" gap="1">
											<input
												type="radio"
												name="type"
												{value}
												bind:group={newState.type}
												style="position: fixed; opacity: 0; pointer-events: none;"
											/>
											<Text size="4" weight="bold">{value}</Text>
										</Flex>
									</Flex>
								</Panel>
							{/each}
						</Flex>
					</Grid>

					<!-- Sub directory -->
					<Grid cols="1fr 40px" gap="3">
						<Grid cols="80px 1fr" gap="3" as="label">
							<Flex justify="start" align="center" as="span">Save to</Flex>
							<Input
								id="fileDir-{id}"
								name="fileDir"
								fullWidth
								size="3"
								bind:value={newState.fileDir}
								disabled={newState.loading}
							/>
						</Grid>

						{#if newState.directories.length}
							<SelectDirectory
								options={newState.directories}
								disabled={newState.directories.length === 0}
								onSelect={(dir) => (newState.fileDir = dir)}
							>
								<FolderIcon />
							</SelectDirectory>
						{/if}
					</Grid>

					<!-- File name -->
					<Grid cols="80px 1fr" gap="3" as="label">
						<Flex justify="start" align="center" as="span">Rename</Flex>
						<Input
							id="fileName-{id}"
							name="fileName"
							size="3"
							fullWidth
							bind:value={newState.fileName}
							disabled={newState.loading}
						/>
					</Grid>

					<!-- Preview -->
					<Separator size="4" />
					<Grid cols="80px minmax(0, 1fr)" gap="3">
						<Flex justify="start" align="center" as="span">Preview</Flex>
						<Panel variant="soft">
							<Flex align="center" gap="3">
								<FileCloudIcon class="shrink-0" />
								<Text weight="medium" style="word-break: break-word;">{newState.pathPreview}</Text>
							</Flex>
						</Panel>
					</Grid>
				</Flex>
			</Panel>

			<!-- Submit -->
			<Flex justify="end">
				<Button size="3" variant="soft" color="orange" radius="full" onclick={newState.create}>
					<PlayIcon weight="fill" />
					Start download
				</Button>
			</Flex>
		{/if}
	</Flex>
</PageLayout>
