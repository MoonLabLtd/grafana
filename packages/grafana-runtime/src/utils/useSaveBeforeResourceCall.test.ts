import { renderHook, act } from '@testing-library/react';

import { getBackendSrv } from '../services';

import { useSaveBeforeResourceCall } from './useSaveBeforeResourceCall';

jest.mock('../services', () => ({
  getBackendSrv: jest.fn(),
}));

describe('useSaveBeforeResourceCall', () => {
  const mockOptions = {
    uid: 'test-uid',
    version: 1,
    name: 'Test Datasource',
    type: 'test',
    id: 1,
    orgId: 1,
    typeName: '',
    typeLogoUrl: '',
    access: 'proxy' as const,
    readOnly: false,
    jsonData: {
      authType: 'default',
      defaultRegion: 'us-east-1',
    },
    secureJsonData: {
      accessKey: 'test-key',
      secretKey: 'test-secret',
    },
  };

  const mockOnOptionsChange = jest.fn();

  beforeEach(() => {
    jest.clearAllMocks();
    (getBackendSrv as jest.Mock).mockReturnValue({
      put: jest.fn().mockResolvedValue({ datasource: { ...mockOptions, version: 2 } }),
    });
  });

  it('should initialize as not saved for version 1', () => {
    const { result } = renderHook(() =>
      useSaveBeforeResourceCall({
        options: mockOptions,
        onOptionsChange: mockOnOptionsChange,
        watchedJsonFields: ['authType', 'defaultRegion'],
        watchedSecureJsonFields: ['accessKey', 'secretKey'],
      })
    );

    expect(result.current.isSaved).toBe(false);
  });

  it('should initialize as saved for version > 1', () => {
    const savedOptions = { ...mockOptions, version: 2 };

    const { result } = renderHook(() =>
      useSaveBeforeResourceCall({
        options: savedOptions,
        onOptionsChange: mockOnOptionsChange,
        watchedJsonFields: ['authType', 'defaultRegion'],
      })
    );

    expect(result.current.isSaved).toBe(true);
  });

  it('should detect changes in watched json fields', async () => {
    const { result, rerender } = renderHook(
      (props) =>
        useSaveBeforeResourceCall({
          options: props.options,
          onOptionsChange: props.onOptionsChange,
          watchedJsonFields: ['authType', 'defaultRegion'],
        }),
      {
        initialProps: {
          options: { ...mockOptions, version: 2 },
          onOptionsChange: mockOnOptionsChange,
        },
      }
    );

    expect(result.current.isSaved).toBe(true);

    // Change a watched field
    const changedOptions = {
      ...mockOptions,
      version: 2,
      jsonData: {
        ...mockOptions.jsonData,
        authType: 'credentials',
      },
    };

    rerender({ options: changedOptions, onOptionsChange: mockOnOptionsChange });

    expect(result.current.isSaved).toBe(false);
  });

  it('should save options when saveIfNeeded is called', async () => {
    const mockPut = jest.fn().mockResolvedValue({ datasource: { ...mockOptions, version: 2 } });
    (getBackendSrv as jest.Mock).mockReturnValue({ put: mockPut });

    const { result } = renderHook(() =>
      useSaveBeforeResourceCall({
        options: mockOptions,
        onOptionsChange: mockOnOptionsChange,
        watchedJsonFields: ['authType'],
      })
    );

    expect(result.current.isSaved).toBe(false);

    await act(async () => {
      await result.current.saveIfNeeded();
    });

    expect(mockPut).toHaveBeenCalledWith(`/api/datasources/uid/${mockOptions.uid}`, mockOptions);
    expect(mockOnOptionsChange).toHaveBeenCalled();
    expect(result.current.isSaved).toBe(true);
  });

  it('should not save options when saveIfNeeded is called if already saved', async () => {
    const mockPut = jest.fn();
    (getBackendSrv as jest.Mock).mockReturnValue({ put: mockPut });

    const savedOptions = { ...mockOptions, version: 2 };

    const { result } = renderHook(() =>
      useSaveBeforeResourceCall({
        options: savedOptions,
        onOptionsChange: mockOnOptionsChange,
        watchedJsonFields: ['authType'],
      })
    );

    await act(async () => {
      await result.current.saveIfNeeded();
    });

    expect(mockPut).not.toHaveBeenCalled();
  });

  it('should throw error when ensureSavedOrError is called if not saved', async () => {
    const { result } = renderHook(() =>
      useSaveBeforeResourceCall({
        options: mockOptions,
        onOptionsChange: mockOnOptionsChange,
      })
    );

    await expect(result.current.ensureSavedOrError()).rejects.toThrow(
      'You need to save the data source before performing this action.'
    );
  });

  it('should not throw error when ensureSavedOrError is called if saved', async () => {
    const savedOptions = { ...mockOptions, version: 2 };

    const { result } = renderHook(() =>
      useSaveBeforeResourceCall({
        options: savedOptions,
        onOptionsChange: mockOnOptionsChange,
      })
    );

    await expect(result.current.ensureSavedOrError()).resolves.not.toThrow();
  });

  it('should throw custom error message', async () => {
    const { result } = renderHook(() =>
      useSaveBeforeResourceCall({
        options: mockOptions,
        onOptionsChange: mockOnOptionsChange,
      })
    );

    await expect(result.current.ensureSavedOrError('Custom error message')).rejects.toThrow('Custom error message');
  });
});
