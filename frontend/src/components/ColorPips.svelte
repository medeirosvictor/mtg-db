<script lang="ts">
  import { parseColors, getColorHex, type MtgColor } from '../lib/colors';

  export let colors: string;

  $: parsed = parseColors(colors);

  const pipStyle = (color: MtgColor) => {
    const hex = getColorHex(color);
    const border = color === 'W' ? '#ccc' : color === 'B' ? '#555' : hex;
    return `background: ${hex}; border-color: ${border};`;
  };
</script>

<div class="flex gap-1 items-center">
  {#each parsed as color}
    <span 
      class="w-4 h-4 rounded-full border flex-shrink-0 inline-block" 
      style={pipStyle(color)}
      title={color}
    ></span>
  {/each}
</div>
