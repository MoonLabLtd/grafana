import { useState, useEffect, useMemo } from 'react';
import { SelectableValue } from '@grafana/data';
import { getBackendSrv } from '@grafana/runtime';
import { Subject, from, of } from 'rxjs';
import { debounceTime, switchMap, catchError } from 'rxjs/operators';

export interface UnsavedConfig {
  jsonData: Record<string, any>;
  secureJsonData?: Record<string, any>;
  assumeRoleArn?: string;
  externalId?: string;
  region?: string;
}

interface ResourceRequest {
  resource: string;
  config: UnsavedConfig;
}

export function useResourcePickerWithUnsavedConfig(
  resource: string,
  config: UnsavedConfig
) {
  const [loading, setLoading] = useState(false);
  const [options, setOptions] = useState<Array<SelectableValue<string>>>([]);
  const [error, setError] = useState<string | null>(null);

  const requestSubject = useMemo(() => new Subject<UnsavedConfig>(), []);

  useEffect(() => {
    const subscription = requestSubject
      .pipe(
        debounceTime(300),
        switchMap((currentConfig) => {
          setLoading(true);
          setError(null);
          
          const payload: ResourceRequest = {
            resource,
            config: currentConfig,
          };

          return from(
            getBackendSrv().post('/api/datasources/proxy/0/resources', payload)
          ).pipe(
            catchError((err) => {
              return of({ 
                error: 'Save the data source with valid authentication first to load available options.' 
              });
            })
          );
        })
      )
      .subscribe((response: any) => {
        setLoading(false);
        
        if (response?.error) {
          setError(response.error);
          setOptions([]);
        } else if (response?.data && Array.isArray(response.data)) {
          setOptions(response.data.map((item: string) => ({ label: item, value: item })));
        } else {
          setOptions([]);
        }
      });

    return () => {
      subscription.unsubscribe();
    };
  }, [requestSubject, resource]);

  useEffect(() => {
    requestSubject.next(config);
  }, [config, requestSubject]);

  return {
    loading,
    options,
    error,
  };
}
