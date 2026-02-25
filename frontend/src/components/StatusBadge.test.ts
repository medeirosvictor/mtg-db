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
    
    // Click the button to open dropdown
    const button = container.querySelector('button');
    if (button) {
      await fireEvent.click(button);
    }
    
    // Check if dropdown options are visible
    expect(container.textContent).toContain('Owned');
    expect(container.textContent).toContain('Planned');
    expect(container.textContent).toContain('Disassembled');
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
