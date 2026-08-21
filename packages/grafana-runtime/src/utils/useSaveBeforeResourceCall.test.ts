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

  it('should detect changes in watched secure json fields', async () => {
    const { result, rerender } = renderHook(
      (props) =>
        useSaveBeforeResourceCall({
          options: props.options,
          onOptionsChange: props.onOptionsChange,
          watchedSecureJsonFields: ['accessKey', 'secretKey'],
        }),
      {
        initialProps: {
          options: { ...mockOptions, version: 2 },
          onOptionsChange: mockOnOptionsChange,
        },
      }
    );

    expect(result.current.isSaved).toBe(true);

    // Change a watched secure field
    const changedOptions = {
      ...mockOptions,
      version: 2,
      secureJsonData: {
        ...mockOptions.secureJsonData,
        accessKey: 'new-key',
      },
    };

    rerender({ options: changedOptions, onOptionsChange: mockOnOptionsChange });

    expect(result.current.isSaved).toBe(false);
  });

  it('should not detect changes in non-watched fields', async () => {
    const { result, rerender } = renderHook(
      (props) =>
        useSaveBeforeResourceCall({
          options: props.options,
          onOptionsChange: props.onOptionsChange,
          watchedJsonFields: ['authType'],
        }),
      {
        initialProps: {
          options: { ...mockOptions, version: 2 },
          onOptionsChange: mockOnOptionsChange,
        },
      }
    );

    expect(result.current.isSaved).toBe(true);

    // Change a non-watched field
    const changedOptions = {
      ...mockOptions,
      version: 2,
      jsonData: {
        ...mockOptions.jsonData,
        defaultRegion: 'eu-west-1',
      },
    };

    rerender({ options: changedOptions, onOptionsChange: mockOnOptionsChange });

    expect(result.current.isSaved).toBe(true);
  });

  it('should save when calling saveIfNeeded on unsaved datasource', async () => {
    const { result } = renderHook(() =>
      useSaveBeforeResourceCall({
        options: mockOptions,
        onOptionsChange: mockOnOptionsChange,
        watchedJsonFields: ['authType', 'defaultRegion'],
      })
    );

    expect(result.current.isSaved).toBe(false);

    await act(async () => {
      await result.current.saveIfNeeded();
    });

    expect(getBackendSrv).toHaveBeenCalled();
    expect(mockOnOptionsChange).toHaveBeenCalled();
  });

  it('should not save when calling saveIfNeeded on saved datasource', async () => {
    const savedOptions = { ...mockOptions, version: 2 };

    const { result } = renderHook(() =>
      useSaveBeforeResourceCall({
        options: savedOptions,
        onOptionsChange: mockOnOptionsChange,
        watchedJsonFields: ['authType', 'defaultRegion'],
      })
    );

    expect(result.current.isSaved).toBe(true);

    await act(async () => {
      await result.current.saveIfNeeded();
    });

    expect(getBackendSrv()).put.not.toHaveBeenCalled();
    expect(mockOnOptionsChange).not.toHaveBeenCalled();
  });

  it('should throw error when calling ensureSavedOrError on unsaved datasource', async () => {
    const { result } = renderHook(() =>
      useSaveBeforeResourceCall({
        options: mockOptions,
        onOptionsChange: mockOnOptionsChange,
        watchedJsonFields: ['authType', 'defaultRegion'],
      })
    );

    expect(result.current.isSaved).toBe(false);

    await act(async () => {
      await expect(result.current.ensureSavedOrError()).rejects.toThrow(
        'You need to save the data source before performing this action. Please save the data source and try again.'
      );
    });
  });

  it('should throw custom error message when calling ensureSavedOrError on unsaved datasource', async () => {
    const { result } = renderHook(() =>
      useSaveBeforeResourceCall({
        options: mockOptions,
        onOptionsChange: mockOnOptionsChange,
        watchedJsonFields: ['authType', 'defaultRegion'],
      })
    );

    expect(result.current.isSaved).toBe(false);

    const customMessage = 'Please save the data source before selecting a region.';

    await act(async () => {
      await expect(result.current.ensureSavedOrError(customMessage)).rejects.toThrow(customMessage);
    });
  });

  it('should not throw error when calling ensureSavedOrError on saved datasource', async () => {
    const savedOptions = { ...mockOptions, version: 2 };

    const { result } = renderHook(() =>
      useSaveBeforeResourceCall({
        options: savedOptions,
        onOptionsChange: mockOnOptionsChange,
        watchedJsonFields: ['authType', 'defaultRegion'],
      })
    );

    expect(result.current.isSaved).toBe(true);

    await act(async () => {
      await expect(result.current.ensureSavedOrError()).resolves.not.toThrow();
    });
  });

  it('should update isSaved when version changes after save', async () => {
    const { result, rerender } = renderHook(
      (props) =>
        useSaveBeforeResourceCall({
          options: props.options,
          onOptionsChange: props.onOptionsChange,
          watchedJsonFields: ['authType', 'defaultRegion'],
        }),
      {
        initialProps: {
          options: mockOptions,
          onOptionsChange: mockOnOptionsChange,
        },
      }
    );

    expect(result.current.isSaved).toBe(false);

    // Simulate version update after save
    const savedOptions = { ...mockOptions, version: 2 };

    rerender({ options: savedOptions, onOptionsChange: mockOnOptionsChange });

    expect(result.current.isSaved).toBe(true);
  });
});
