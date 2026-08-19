import { render, screen, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';

import type { StandardEditorProps } from '@grafana/data';
import {
  DEFAULT_SECTION_ORDER,
  type SectionId,
} from 'app/features/explore/TraceView/components/TraceTimelineViewer/SpanDetail/sectionOrder';

import { SectionOrderEditor } from './SectionOrderEditor';

type EditorProps = StandardEditorProps<SectionId[], unknown, SectionId[]>;

const renderEditor = (value?: SectionId[]) => {
  const onChange = jest.fn();
  const props: EditorProps = {
    value: value as SectionId[],
    onChange,
    context: { data: [] },
    item: { id: 'spanDetailSectionOrder', title: 'spanDetailSectionOrder', path: 'spanDetailSectionOrder' } as never,
  };
  render(<SectionOrderEditor {...props} />);
  return { onChange };
};

const getItemLabels = () =>
  screen
    .getAllByRole('listitem')
    .map(
      (item) =>
        (within(item).getByText(/Span attributes|Resource attributes|Events|Warnings|Stack trace/).textContent ?? '').replace(
          /(span|resource|exception|warn|stack)$/,
          ''
        )
    );

describe('<SectionOrderEditor>', () => {
  it('renders all five sections in the default order', () => {
    renderEditor();
    const items = screen.getAllByRole('listitem');
    expect(items).toHaveLength(5);
    expect(getItemLabels()).toEqual([
      'Span attributes',
      'Resource attributes',
      'Events',
      'Warnings',
      'Stack trace',
    ]);
  });

  it('renders a custom order when provided', () => {
    renderEditor(['events', 'spanAttributes', 'resourceAttributes', 'warnings', 'stackTraces']);
    expect(getItemLabels()).toEqual(['Events', 'Span attributes', 'Resource attributes', 'Warnings', 'Stack trace']);
  });

  it('keeps reset disabled and hides the unsaved indicator when the order matches default', () => {
    renderEditor();
    expect(screen.getByRole('button', { name: 'Reset to default' })).toBeDisabled();
    expect(screen.queryByText('Unsaved order')).not.toBeInTheDocument();
  });

  it('enables reset and shows the unsaved indicator for a customized order', () => {
    renderEditor(['events', 'spanAttributes', 'resourceAttributes', 'warnings', 'stackTraces']);
    expect(screen.getByRole('button', { name: 'Reset to default' })).toBeEnabled();
    expect(screen.getByText('Unsaved order')).toBeInTheDocument();
  });

  it('moves an item down via the down button and persists the order', async () => {
    const user = userEvent.setup();
    const { onChange } = renderEditor();
    await user.click(screen.getAllByRole('button', { name: 'Move down' })[0]);
    expect(onChange).toHaveBeenCalledWith(['resourceAttributes', 'spanAttributes', 'events', 'warnings', 'stackTraces']);
    // down button for the last item is disabled
    expect(screen.getAllByRole('button', { name: 'Move down' })[4]).toBeDisabled();
  });

  it('disables the up button for the first item and down for the last', () => {
    renderEditor();
    expect(screen.getAllByRole('button', { name: 'Move up' })[0]).toBeDisabled();
    expect(screen.getAllByRole('button', { name: 'Move down' })[4]).toBeDisabled();
  });

  it('moves the focused item with Alt+ArrowDown', async () => {
    const user = userEvent.setup();
    const { onChange } = renderEditor();
    const first = screen.getAllByRole('listitem')[0];
    first.focus();
    await user.keyboard('{Alt>}{ArrowDown}{/Alt}');
    expect(onChange).toHaveBeenCalledWith(['resourceAttributes', 'spanAttributes', 'events', 'warnings', 'stackTraces']);
  });

  it('moves the focused item to the first position with Alt+Home', async () => {
    const user = userEvent.setup();
    const { onChange } = renderEditor();
    const last = screen.getAllByRole('listitem')[4];
    last.focus();
    await user.keyboard('{Alt>}{Home}{/Alt}');
    expect(onChange).toHaveBeenCalledWith(['stackTraces', 'spanAttributes', 'resourceAttributes', 'events', 'warnings']);
  });

  it('resets a customized order back to default via the reset button', async () => {
    const user = userEvent.setup();
    const { onChange } = renderEditor(['events', 'spanAttributes', 'resourceAttributes', 'warnings', 'stackTraces']);
    await user.click(screen.getByRole('button', { name: 'Reset to default' }));
    expect(onChange).toHaveBeenCalledWith(DEFAULT_SECTION_ORDER);
    expect(screen.getByRole('button', { name: 'Reset to default' })).toBeDisabled();
  });
});
