<script lang="ts">
	import PageLayout from '$lib/components/PageLayout.svelte';
	import { Flexbox } from 'svxui';
	import DownloadList from '../DownloadList.svelte';
</script>

<PageLayout title="All">
	<Flexbox direction="column" gap="5">
		<DownloadList status={[]} type={[]}/> 
	</Flexbox>
</PageLayout>
