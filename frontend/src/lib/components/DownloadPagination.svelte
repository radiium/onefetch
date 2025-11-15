<script lang="ts">
	import type { Pagination } from '$lib/types/types';
	import { Button, Flexbox } from 'svxui';

	type Props = {
		currentPage?: number;
		pagination?: Pagination;
	};

	let { currentPage = $bindable(1), pagination }: Props = $props();
</script>

{#if pagination}
	{currentPage}
	<Flexbox gap="3">
		{#each { length: pagination.totalPages } as _, i}
			{@const page = i + 1}
			<Button
				size="2"
				iconOnly
				variant={page === currentPage ? 'solid' : 'outline'}
				onclick={() => {
					currentPage = page;
				}}
			>
				{page}
			</Button>
		{/each}
	</Flexbox>
{/if}
