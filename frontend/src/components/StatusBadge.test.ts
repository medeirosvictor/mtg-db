import { describe, it, expect } from 'vitest';
import { render, fireEvent, screen } from '@testing-library/svelte';
import StatusBadge from '../components/StatusBadge.svelte';

describe('StatusBadge', () => {
  it('should render with correct status text', () => {
    const { container } = render(StatusBadge, { props: { status: 'Owned' } });
    expect(container.textContent).toContain('Owned');
  });

  it('should render Planned status', () => {
    const { container } = render(StatusBadge, { props: { status: 'Planned' } });
    expect(container.textContent).toContain('Planned');
  });

  it('should render Disassembled status', () => {
    const { container } = render(StatusBadge, { props: { status: 'Disassembled' } });
    expect(container.textContent).toContain('Disassembled');
  });

  it('should show dropdown when clicked', async () => {
    const { container } = render(StatusBadge, { props: { status: 'Owned' } });

    // Before clicking: only the badge button is rendered, no dropdown options
    const beforeButtons = container.querySelectorAll('button');
    expect(beforeButtons.length).toBe(1);

    await fireEvent.click(beforeButtons[0]);

    // After clicking: badge button + one button per status option
    const afterButtons = container.querySelectorAll('button');
    expect(afterButtons.length).toBe(4); // 1 badge + 3 options

    // All status options must be present in the dropdown buttons
    const optionTexts = Array.from(afterButtons)
      .slice(1)
      .map((b) => b.textContent ?? '');
    expect(optionTexts.some((t) => t.includes('Owned'))).toBe(true);
    expect(optionTexts.some((t) => t.includes('Planned'))).toBe(true);
    expect(optionTexts.some((t) => t.includes('Disassembled'))).toBe(true);
  });

  it('should emit change event when status is selected', async () => {
    const { component } = render(StatusBadge, { props: { status: 'Owned' } });
    
    let emitted = false;
    component.$on('change', () => {
      emitted = true;
    });
    
    // Click button to open dropdown
    const button = document.querySelector('button');
    if (button) {
      await fireEvent.click(button);
    }
    
    // Click on Planned option
    const plannedOption = Array.from(document.querySelectorAll('button')).find(
      (btn) => btn.textContent?.includes('Planned')
    );
    if (plannedOption) {
      await fireEvent.click(plannedOption);
    }
    
    expect(emitted).toBe(true);
  });
});
