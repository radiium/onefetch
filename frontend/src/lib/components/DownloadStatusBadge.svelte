<script lang="ts">
	import { DownloadStatus, type Download } from '$lib/types/types';
	import { statusColor } from '$lib/utils/status-color';
	import Info from 'phosphor-svelte/lib/Info';
	import WarningCircle from 'phosphor-svelte/lib/WarningCircle';
	import { Badge, Flexbox, Floating, Text } from 'svxui';

	type Props = {
		dl: Download;
	};

	let { dl }: Props = $props();
	let isOpen = $state(false);
</script>

{#if dl.status === DownloadStatus.FAILED && dl.errorMessage}
	<Floating bind:isOpen variant="soft" color="red" size="4" offset={8} arrow placement="bottom-end">
		{#snippet trigger()}
			<Badge
				variant="soft"
				size="3"
				color={statusColor[dl.status]}
				onmouseenter={() => (isOpen = true)}
				onmouseleave={() => (isOpen = false)}
			>
				<Info />
				{dl.status}
			</Badge>
		{/snippet}
		{#snippet content()}
			<Flexbox gap="3" style="max-width: 500px;">
				<WarningCircle size="24px" weight="fill" color="var(--tomato-9)" class="shrink-0" />
				<Text color="red">{dl.errorMessage}</Text>
			</Flexbox>
		{/snippet}
	</Floating>
{:else}
	<Badge variant="soft" size="3" color={statusColor[dl.status]} style="">
		{dl.status}
	</Badge>
{/if}
