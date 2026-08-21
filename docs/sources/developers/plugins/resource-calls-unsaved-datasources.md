# Resource Calls on Unsaved Data Sources

## Problem

When a user creates a new data source or modifies an existing data source's configuration, the data source has a UID but the configuration (including credentials) may not have been saved to the database yet. This can cause resource calls (e.g., fetching dropdown options from an API) to fail because the backend doesn't have access to the latest configuration.

## Solution

Use the `useSaveBeforeResourceCall` hook from `@grafana/runtime` to ensure the data source configuration is saved before making resource calls.

## Usage

### Basic Example

```tsx
import { useSaveBeforeResourceCall } from '@grafana/runtime';

const ConfigEditor = ({ options, onOptionsChange }) => {
  const { isSaved, saveIfNeeded, ensureSavedOrError } = useSaveBeforeResourceCall({
    options,
    onOptionsChange,
    watchedJsonFields: ['authType', 'defaultRegion', 'assumeRoleArn'],
    watchedSecureJsonFields: ['accessKey', 'secretKey'],
  });

  const handleDropdownOpen = async () => {
    // Option 1: Auto-save before opening (recommended)
    await saveIfNeeded();
    const regions = await datasource.getResource('regions');
    setRegions(regions);
  };

  // Option 2: Show error if not saved
  const handleDropdownOpenWithError = async () => {
    try {
      await ensureSavedOrError('You need to save the data source before selecting a region.');
      const regions = await datasource.getResource('regions');
      setRegions(regions);
    } catch (error) {
      // Show error to user
      setError(error.message);
    }
  };

  return (
    <div>
      {/* Your config UI */}
      <Select
        options={regions}
        onOpen={handleDropdownOpen}
        placeholder={isSaved ? 'Select a region' : 'Save data source first'}
      />
    </div>
  );
};
```

### Configuration Options

- `options`: The current data source options/settings
- `onOptionsChange`: Callback to update options after saving
- `watchedJsonFields`: List of `jsonData` fields to watch for changes
- `watchedSecureJsonFields`: List of `secureJsonData` fields to watch for changes

### Return Values

- `isSaved`: Boolean indicating if the data source configuration has been saved
- `saveIfNeeded()`: Async function that saves configuration if there are unsaved changes
- `ensureSavedOrError(errorMessage?)`: Async function that throws an error if not saved

## Best Practices

1. **Watch all relevant configuration fields**: Include all fields that affect resource calls in the `watchedJsonFields` and `watchedSecureJsonFields` arrays.

2. **Use auto-save for better UX**: Using `saveIfNeeded()` before opening dropdowns provides a seamless experience for users.

3. **Provide clear feedback**: If using `ensureSavedOrError()`, provide a clear error message and disable or update the UI to indicate that the data source needs to be saved first.

4. **Handle errors gracefully**: Always wrap resource calls in try-catch blocks and provide user-friendly error messages.

## Example: AWS Data Source

For AWS-based data sources that need to fetch regions, databases, or workgroups:

```tsx
import { useSaveBeforeResourceCall } from '@grafana/runtime';

const AWSConfigEditor = ({ options, onOptionsChange }) => {
  const { isSaved, saveIfNeeded } = useSaveBeforeResourceCall({
    options,
    onOptionsChange,
    watchedJsonFields: ['authType', 'defaultRegion', 'assumeRoleArn', 'externalId', 'profile'],
    watchedSecureJsonFields: ['accessKey', 'secretKey'],
  });

  const [regions, setRegions] = useState([]);

  const handleRegionDropdownOpen = async () => {
    await saveIfNeeded();
    const data = await datasource.getResource('regions');
    setRegions(data);
  };

  return (
    <Select
      options={regions}
      onOpen={handleRegionDropdownOpen}
      placeholder={isSaved ? 'Select a region' : 'Configuration will be saved automatically'}
    />
  );
};
```

## Related

- [Build a data source plugin](https://grafana.com/developers/plugin-tools/create-a-plugin/develop-a-plugin/build-a-data-source-plugin)
- [Data source resource calls](https://grafana.com/developers/plugin-tools/create-a-plugin/develop-a-plugin/backend-addons)
