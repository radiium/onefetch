<script lang="ts">
	import PageLayout from '$lib/components/PageLayout.svelte';
	import { DownloadStatus } from '$lib/types/types';
	import { Flexbox } from 'svxui';
	import DownloadList from '../DownloadList.svelte';
</script>

<PageLayout title="Paused">
	<Flexbox direction="column" gap="5">
		<DownloadList status={[DownloadStatus.PAUSED]} type={[]} />
	</Flexbox>
</PageLayout>
