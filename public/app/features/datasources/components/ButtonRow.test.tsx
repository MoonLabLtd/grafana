import { screen, render } from '@testing-library/react';
import userEvent from '@testing-library/user-event';

import { selectors } from '@grafana/e2e-selectors';

import { ButtonRow, type Props } from './ButtonRow';

const setup = (propOverrides?: object) => {
  const props: Props = {
    canSave: false,
    canDelete: true,
    onDelete: jest.fn(),
    onSubmit: jest.fn(),
    onTest: jest.fn(),
  };

  Object.assign(props, propOverrides);

  return render(<ButtonRow {...props} />);
};

describe('<ButtonRow>', () => {
  it('should render component', () => {
    setup();

    expect(screen.getByTestId(selectors.pages.DataSource.delete)).toBeInTheDocument();
    expect(screen.getByText('Test')).toBeInTheDocument();
  });

  it('should render save & test', () => {
    setup({ canSave: true });

    expect(screen.getByTestId(selectors.pages.DataSource.saveAndTest)).toBeInTheDocument();
  });

  it('should disable Save & test and show tooltip when required fields are invalid', async () => {
    setup({ canSave: true, requiredFieldsValid: false, requiredFieldsMessage: 'Complete all required fields before saving.' });

    const button = screen.getByTestId(selectors.pages.DataSource.saveAndTest);
    expect(button).toBeDisabled();

    await userEvent.hover(button);
    expect(await screen.findByText('Complete all required fields before saving.')).toBeInTheDocument();
  });

  it('should enable Save & test when required fields are valid', () => {
    setup({ canSave: true, requiredFieldsValid: true });

    const button = screen.getByTestId(selectors.pages.DataSource.saveAndTest);
    expect(button).toBeEnabled();
  });

  it('should keep Save & test enabled by default when requiredFieldsValid is not provided', () => {
    setup({ canSave: true });

    const button = screen.getByTestId(selectors.pages.DataSource.saveAndTest);
    expect(button).toBeEnabled();
  });
});
