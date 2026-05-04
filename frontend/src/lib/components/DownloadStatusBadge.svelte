<script lang="ts">
	import { DownloadStatus, type Download } from '$lib/types/types';
	import { statusColor } from '$lib/utils/status-color';
	import InfoIcon from 'phosphor-svelte/lib/InfoIcon';
	import WarningCircleIcon from 'phosphor-svelte/lib/WarningCircleIcon';
	import { Badge, Flex, Floating, Text } from 'svxui';

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
				<InfoIcon />
				{dl.status}
			</Badge>
		{/snippet}
		{#snippet content()}
			<Flex gap="3" style="max-width: 500px;">
				<WarningCircleIcon size="24px" weight="fill" color="var(--tomato-9)" class="shrink-0" />
				<Text color="red">{dl.errorMessage}</Text>
			</Flex>
		{/snippet}
	</Floating>
{:else}
	<Badge variant="soft" size="3" color={statusColor[dl.status]} style="">
		{dl.status}
	</Badge>
{/if}
