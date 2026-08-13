import { render, screen } from '@testing-library/react';

import { LoadingState, type PanelProps } from '@grafana/data';
import { getDataSourceSrv } from '@grafana/runtime';

import { TracesPanel } from './TracesPanel';

jest.mock('@grafana/runtime', () => ({
  ...jest.requireActual('@grafana/runtime'),
  getDataSourceSrv: jest.fn(),
}));

let lastTraceViewProps: Record<string, unknown> | undefined;
jest.mock('app/features/explore/TraceView/TraceView', () => ({
  TraceView: (props: Record<string, unknown>) => {
    lastTraceViewProps = props;
    return <div data-testid="trace-view" />;
  },
}));

jest.mock('app/features/explore/TraceView/utils/transform', () => ({
  transformDataFrames: jest.fn(() => ({ traceID: 'test-trace' })),
}));

const mockGetDataSourceSrv = getDataSourceSrv as jest.Mock;

describe('TracesPanel', () => {
  beforeEach(() => {
    mockGetDataSourceSrv.mockReturnValue({
      get: jest.fn().mockResolvedValue({ uid: 'tempo-uid', type: 'tempo' }),
    });
  });

  afterEach(() => {
    jest.clearAllMocks();
  });

  it('shows no data message when no data supplied', async () => {
    const props = {
      data: {
        error: undefined,
        series: [],
        state: LoadingState.Done,
      },
    } as unknown as PanelProps;

    render(<TracesPanel {...props} />);

    await screen.findByText('No data found in response');
  });

  it('resolves datasource from data.request when available', async () => {
    const props = {
      data: {
        series: [{ fields: [], length: 0 }],
        state: LoadingState.Done,
        request: {
          targets: [{ datasource: { uid: 'tempo-from-request' } }],
        },
        timeRange: { from: 0, to: 1 },
      },
      options: {},
      replaceVariables: (v: string) => v,
    } as unknown as PanelProps;

    render(<TracesPanel {...props} />);

    await screen.findByTestId('trace-view');

    const getSrv = mockGetDataSourceSrv.mock.results[0].value;
    expect(getSrv.get).toHaveBeenCalledWith('tempo-from-request');
  });

  it('falls back to options.datasource when data.request is undefined', async () => {
    const props = {
      data: {
        series: [{ fields: [], length: 0 }],
        state: LoadingState.Done,
        timeRange: { from: 0, to: 1 },
      },
      options: {
        datasource: { uid: 'tempo-from-options', type: 'tempo' },
      },
      replaceVariables: (v: string) => v,
    } as unknown as PanelProps;

    render(<TracesPanel {...props} />);

    await screen.findByTestId('trace-view');

    const getSrv = mockGetDataSourceSrv.mock.results[0].value;
    expect(getSrv.get).toHaveBeenCalledWith('tempo-from-options');
  });

  it('does not call getDataSourceSrv().get when no uid is available', async () => {
    const getMock = jest.fn();
    mockGetDataSourceSrv.mockReturnValue({ get: getMock });

    const props = {
      data: {
        series: [{ fields: [], length: 0 }],
        state: LoadingState.Done,
        timeRange: { from: 0, to: 1 },
      },
      options: {},
      replaceVariables: (v: string) => v,
    } as unknown as PanelProps;

    render(<TracesPanel {...props} />);

    await screen.findByTestId('trace-view');

    expect(getMock).not.toHaveBeenCalled();
  });

  it('normalizes and forwards spanDetail section order and hidden sections to TraceView', async () => {
    const props = {
      data: {
        series: [{ fields: [], length: 0 }],
        state: LoadingState.Done,
        timeRange: { from: 0, to: 1 },
      },
      options: {
        spanDetail: {
          sectionOrder: ['events', 'span'],
          hiddenSections: ['events'],
        },
      },
      replaceVariables: (v: string) => v,
    } as unknown as PanelProps;

    render(<TracesPanel {...props} />);

    await screen.findByTestId('trace-view');

    expect(lastTraceViewProps?.sectionOrder).toEqual([
      'events',
      'span',
      'resource',
      'warnings',
      'stack',
      'references',
      'flame',
    ]);
    expect(lastTraceViewProps?.hiddenSections).toEqual(['events']);
  });

  it('forwards the default section order when spanDetail is not configured', async () => {
    const props = {
      data: {
        series: [{ fields: [], length: 0 }],
        state: LoadingState.Done,
        timeRange: { from: 0, to: 1 },
      },
      options: {},
      replaceVariables: (v: string) => v,
    } as unknown as PanelProps;

    render(<TracesPanel {...props} />);

    await screen.findByTestId('trace-view');

    expect(lastTraceViewProps?.sectionOrder).toEqual([
      'span',
      'resource',
      'events',
      'warnings',
      'stack',
      'references',
      'flame',
    ]);
    expect(lastTraceViewProps?.hiddenSections).toEqual([]);
  });
});
