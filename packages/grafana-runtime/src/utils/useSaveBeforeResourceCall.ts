import { isEqual } from 'lodash';
import { useCallback, useEffect, useRef, useState } from 'react';

import { type DataSourceSettings, type DataSourceJsonData } from '@grafana/data';
import { getBackendSrv } from '@grafana/runtime';

/**
 * Options for the useSaveBeforeResourceCall hook
 */
export interface UseSaveBeforeResourceCallOptions<
  TJsonData extends DataSourceJsonData = DataSourceJsonData,
  TSecureJsonData = Record<string, unknown>,
> {
  /**
   * The current datasource options/settings
   */
  options: DataSourceSettings<TJsonData, TSecureJsonData>;

  /**
   * Callback to update options after saving
   */
  onOptionsChange: (options: DataSourceSettings<TJsonData, TSecureJsonData>) => void;

  /**
   * List of jsonData fields to watch for changes.
   * If any of these fields change, the datasource is considered unsaved.
   */
  watchedJsonFields?: (keyof TJsonData)[];

  /**
   * List of secureJsonData fields to watch for changes.
   * If any of these fields change, the datasource is considered unsaved.
   */
  watchedSecureJsonFields?: (keyof TSecureJsonData)[];
}

/**
 * Hook to manage datasource configuration saving before resource calls.
 * This ensures that resource calls have access to the latest configuration
 * (including credentials) even for unsaved datasources.
 *
 * @param options - Configuration options for the hook
 * @returns An object with:
 *   - isSaved: boolean indicating if the datasource configuration has been saved
 *   - saveIfNeeded: async function that saves configuration if there are unsaved changes
 *   - ensureSavedOrError: async function that throws an error if not saved (useful for UX)
 *
 * @example
 * ```tsx
 * const { isSaved, saveIfNeeded, ensureSavedOrError } = useSaveBeforeResourceCall({
 *   options,
 *   onOptionsChange,
 *   watchedJsonFields: ['authType', 'defaultRegion', 'assumeRoleArn'],
 *   watchedSecureJsonFields: ['accessKey', 'secretKey'],
 * });
 *
 * const handleDropdownOpen = async () => {
 *   // Option 1: Auto-save before opening
 *   await saveIfNeeded();
 *   const data = await datasource.getResource('some-resource');
 *
 *   // Option 2: Show error if not saved
 *   try {
 *     await ensureSavedOrError('You need to save the data source before selecting a value.');
 *     const data = await datasource.getResource('some-resource');
 *   } catch (error) {
 *     // Show error to user
 *   }
 * };
 * ```
 */
export function useSaveBeforeResourceCall<
  TJsonData extends DataSourceJsonData = DataSourceJsonData,
  TSecureJsonData = Record<string, unknown>,
>(optionsConfig: UseSaveBeforeResourceCallOptions<TJsonData, TSecureJsonData>) {
  const { options, onOptionsChange, watchedJsonFields = [], watchedSecureJsonFields = [] } = optionsConfig;

  // Track if the datasource has been saved at least once
  const [isSaved, setIsSaved] = useState(() => {
    // version 1 means it's newly created but not yet configured
    // version > 1 means it has been saved at least once
    return !!options.version && options.version > 1;
  });

  // Keep track of the previous values to detect changes
  const previousValuesRef = useRef({
    jsonData: options.jsonData,
    secureJsonData: options.secureJsonData,
  });

  // Watch for changes in the specified fields
  useEffect(() => {
    const currentWatchedJsonValues = watchedJsonFields.map((field) => options.jsonData[field]);
    const previousWatchedJsonValues = watchedJsonFields.map((field) => previousValuesRef.current.jsonData[field]);

    const currentWatchedSecureValues = watchedSecureJsonFields.map(
      (field) => options.secureJsonData?.[field as string]
    );
    const previousWatchedSecureValues = watchedSecureJsonFields.map(
      (field) => previousValuesRef.current.secureJsonData?.[field as string]
    );

    const hasJsonChanges = !isEqual(currentWatchedJsonValues, previousWatchedJsonValues);
    const hasSecureJsonChanges = !isEqual(currentWatchedSecureValues, previousWatchedSecureValues);

    if (hasJsonChanges || hasSecureJsonChanges) {
      setIsSaved(false);
      previousValuesRef.current = {
        jsonData: options.jsonData,
        secureJsonData: options.secureJsonData,
      };
    }
  }, [options.jsonData, options.secureJsonData, watchedJsonFields, watchedSecureJsonFields]);

  // Update saved state when version changes
  useEffect(() => {
    if (options.version && options.version > 1) {
      setIsSaved(true);
    }
  }, [options.version]);

  // Save the datasource configuration if needed
  const saveIfNeeded = useCallback(async () => {
    if (isSaved) {
      return;
    }

    const result = await getBackendSrv().put<{ datasource: DataSourceSettings<TJsonData, TSecureJsonData> }>(
      `/api/datasources/uid/${options.uid}`,
      options
    );

    // Update the version after successful save
    if (result.datasource.version !== options.version) {
      onOptionsChange(result.datasource);
    }

    setIsSaved(true);
  }, [isSaved, options, onOptionsChange]);

  // Ensure saved or throw an error (useful for UX feedback)
  const ensureSavedOrError = useCallback(
    async (errorMessage = 'You need to save the data source before performing this action.') => {
      if (!isSaved) {
        throw new Error(errorMessage);
      }
    },
    [isSaved]
  );

  return {
    isSaved,
    saveIfNeeded,
    ensureSavedOrError,
  };
}
