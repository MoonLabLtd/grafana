import * as React from 'react';

import { selectors } from '@grafana/e2e-selectors';
import { Trans } from '@grafana/i18n';
import { Button, Tooltip } from '@grafana/ui';

export interface Props {
  canSave: boolean;
  canDelete: boolean;
  /**
   * Whether the plugin's required fields are populated. When false the
   * "Save & test" button is disabled and a tooltip explains why.
   * Defaults to true for backward compatibility.
   */
  requiredFieldsValid?: boolean;
  /**
   * Message shown in the tooltip when required fields are missing.
   */
  requiredFieldsMessage?: string;
  onDelete: (event: React.MouseEvent<HTMLButtonElement>) => void;
  onSubmit: (event: React.MouseEvent<HTMLButtonElement>) => void;
  onTest: (event: React.MouseEvent<HTMLButtonElement, MouseEvent>) => void;
  children?: React.ReactNode;
}

export function ButtonRow({
  canSave,
  canDelete,
  requiredFieldsValid = true,
  requiredFieldsMessage = 'Complete all required fields before saving.',
  onDelete,
  onSubmit,
  onTest,
  children,
}: Props) {
  const requiredFieldsValidResolved = canSave && requiredFieldsValid;
  const saveButton = (
    <Button
      type="submit"
      variant="primary"
      disabled={!requiredFieldsValidResolved}
      onClick={onSubmit}
      data-testid={selectors.pages.DataSource.saveAndTest}
      id={selectors.pages.DataSource.saveAndTest}
    >
      <Trans i18nKey="datasources.button-row.save-and-test">Save &amp; test</Trans>
    </Button>
  );

  return (
    <div className="gf-form-button-row">
      <Button
        type="button"
        variant="destructive"
        disabled={!canDelete}
        onClick={onDelete}
        data-testid={selectors.pages.DataSource.delete}
      >
        <Trans i18nKey="datasources.button-row.delete">Delete</Trans>
      </Button>

      {children}

      {canSave && (!requiredFieldsValid ? <Tooltip content={requiredFieldsMessage}>{saveButton}</Tooltip> : saveButton)}
      {!canSave && (
        <Button variant="primary" onClick={onTest}>
          <Trans i18nKey="datasources.button-row.test">Test</Trans>
        </Button>
      )}
    </div>
  );
}
