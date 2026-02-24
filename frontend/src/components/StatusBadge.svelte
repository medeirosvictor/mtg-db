<script lang="ts">
  export let status: string;

  $: badgeClass = {
    'Owned': 'bg-green text-bg-primary',
    'Planned': 'bg-yellow text-bg-primary',
    'Disassembled': 'bg-text-muted text-bg-primary',
  }[status] || 'bg-text-muted text-bg-primary';
</script>

<span class="inline-block px-2 py-0.5 rounded text-[11px] font-semibold uppercase tracking-wide {badgeClass}">{status}</span>
