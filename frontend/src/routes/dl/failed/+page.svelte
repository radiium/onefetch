<script lang="ts">
	import PageLayout from '$lib/components/PageLayout.svelte';
	import { DownloadStatus } from '$lib/types/types';
	import { Flexbox } from 'svxui';
	import DownloadList from '../DownloadList.svelte';
</script>

<PageLayout title="Failed">
	<Flexbox direction="column" gap="5">
		<DownloadList status={[DownloadStatus.FAILED]} type={[]}/> 
	</Flexbox>
</PageLayout>
