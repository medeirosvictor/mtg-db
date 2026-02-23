<script lang="ts">
  export let status: string;

  $: badgeClass = {
    'Owned': 'badge-owned',
    'Planned': 'badge-planned',
    'Disassembled': 'badge-disassembled',
  }[status] || 'badge-disassembled';
</script>

<span class="badge {badgeClass}">{status}</span>
