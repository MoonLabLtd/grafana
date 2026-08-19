import { css } from '@emotion/css';
import { memo, useMemo, useState, type DragEvent, type KeyboardEvent } from 'react';

import { type GrafanaTheme2, type StandardEditorProps } from '@grafana/data';
import { t } from '@grafana/i18n';
import { Button, useStyles2 } from '@grafana/ui';
import {
  DEFAULT_SECTION_ORDER,
  SECTION_ORDER_METADATA,
  normalizeSectionOrder,
  type SectionId,
} from 'app/features/explore/TraceView/components/TraceTimelineViewer/SpanDetail/sectionOrder';

type Props = StandardEditorProps<SectionId[], unknown, SectionId[]>;

const arraysEqual = (a: SectionId[], b: SectionId[]) => {
  if (a.length !== b.length) {
    return false;
  }
  for (let i = 0; i < a.length; i++) {
    if (a[i] !== b[i]) {
      return false;
    }
  }
  return true;
};

const getStyles = (theme: GrafanaTheme2) => ({
  list: css({
    display: 'flex',
    flexDirection: 'column',
    gap: '6px',
    margin: 0,
    padding: 0,
    listStyle: 'none',
  }),
  item: css({
    display: 'flex',
    alignItems: 'center',
    gap: '8px',
    background: 'var(--secondary-background)',
    border: '1px solid var(--border-weak)',
    borderRadius: theme.shape.radius.sm,
    padding: '0 8px',
    height: '36px',
    cursor: 'grab',
    userSelect: 'none',
    '&:hover': {
      borderColor: 'var(--border-medium)',
    },
    '&[aria-grabbed="true"]': {
      opacity: 0.45,
      borderColor: 'var(--info)',
    },
    '&:focus-visible': {
      outline: 'none',
      boxShadow: '0 0 0 3px var(--focus-outline)',
      borderColor: 'var(--info)',
    },
  }),
  handle: css({
    color: 'var(--text-secondary)',
    display: 'grid',
    placeItems: 'center',
    cursor: 'grab',
  }),
  index: css({
    fontFamily: 'var(--font-family-monospace)',
    fontSize: '11px',
    color: 'var(--text-secondary)',
    width: '20px',
    textAlign: 'center',
  }),
  name: css({
    fontSize: '12px',
    color: 'var(--text-primary)',
    flex: 1,
    minWidth: 0,
    overflow: 'hidden',
    textOverflow: 'ellipsis',
    whiteSpace: 'nowrap',
    display: 'flex',
    alignItems: 'center',
    gap: '6px',
  }),
  tag: css({
    fontSize: '10px',
    color: 'var(--text-secondary)',
    fontFamily: 'var(--font-family-monospace)',
  }),
  actions: css({
    display: 'flex',
    gap: '2px',
  }),
  footer: css({
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'space-between',
    gap: '8px',
    marginTop: '8px',
  }),
  dirty: css({
    fontSize: '11px',
    color: 'var(--warning-text)',
    fontWeight: 500,
  }),
});

const getSectionLabel = (id: SectionId) => {
  const meta = SECTION_ORDER_METADATA[id];
  return t(meta.labelKey, meta.label);
};

export const SectionOrderEditor = memo<Props>(({ value, onChange }) => {
  const styles = useStyles2(getStyles);
  const appliedOrder = useMemo(() => normalizeSectionOrder(value), [value]);
  const [workingOrder, setWorkingOrder] = useState<SectionId[]>(appliedOrder);

  const isCustomized = !arraysEqual(workingOrder, DEFAULT_SECTION_ORDER);

  const commit = (next: SectionId[]) => {
    setWorkingOrder(next);
    onChange(normalizeSectionOrder(next));
  };

  const move = (from: number, to: number) => {
    if (Number.isNaN(from) || to < 0 || to >= workingOrder.length || from === to) {
      return;
    }
    const next = workingOrder.slice();
    const [moved] = next.splice(from, 1);
    next.splice(to, 0, moved);
    commit(next);
  };

  const handleKeyDown = (event: KeyboardEvent<HTMLDivElement>, index: number) => {
    if (!event.altKey) {
      return;
    }
    if (event.key === 'ArrowUp') {
      event.preventDefault();
      move(index, index - 1);
    } else if (event.key === 'ArrowDown') {
      event.preventDefault();
      move(index, index + 1);
    } else if (event.key === 'Home') {
      event.preventDefault();
      move(index, 0);
    } else if (event.key === 'End') {
      event.preventDefault();
      move(index, workingOrder.length - 1);
    }
  };

  const handleDragStart = (event: DragEvent<HTMLDivElement>, index: number) => {
    event.dataTransfer.effectAllowed = 'move';
    event.dataTransfer.setData('text/plain', String(index));
    event.currentTarget.setAttribute('aria-grabbed', 'true');
  };

  const handleDragEnd = (event: DragEvent<HTMLDivElement>) => {
    event.currentTarget.removeAttribute('aria-grabbed');
  };

  const handleDragOver = (event: DragEvent<HTMLDivElement>) => {
    if (event.dataTransfer.types.includes('text/plain')) {
      event.preventDefault();
      event.dataTransfer.dropEffect = 'move';
    }
  };

  const handleDrop = (event: DragEvent<HTMLDivElement>, to: number) => {
    event.preventDefault();
    const from = Number(event.dataTransfer.getData('text/plain'));
    move(from, to);
  };

  return (
    <div>
      <ul className={styles.list} aria-label={t('traces.section-order.label', 'Span detail section order')}>
        {workingOrder.map((id, index) => {
          const label = getSectionLabel(id);
          const positionLabel = t('traces.section-order.item-position', '{section}, position {position} of {total}', {
            section: label,
            position: index + 1,
            total: workingOrder.length,
          });
          return (
            /* eslint-disable jsx-a11y/no-noninteractive-tabindex, jsx-a11y/no-noninteractive-element-interactions */
            <div
              key={id}
              className={styles.item}
              role="listitem"
              tabIndex={0}
              draggable
              aria-label={positionLabel}
              aria-grabbed="false"
              onDragStart={(event) => handleDragStart(event, index)}
              onDragEnd={handleDragEnd}
              onDragOver={handleDragOver}
              onDrop={(event) => handleDrop(event, index)}
              onKeyDown={(event) => handleKeyDown(event, index)}
            >
              <span className={styles.handle} aria-hidden="true">
                ⋮⋮
              </span>
              <span className={styles.index}>{index + 1}</span>
              <span className={styles.name}>
                {label}
                <span className={styles.tag}>{SECTION_ORDER_METADATA[id].tag}</span>
              </span>
              <span className={styles.actions}>
                <Button
                  variant="secondary"
                  size="sm"
                  fill="text"
                  icon="arrow-up"
                  aria-label={t('traces.section-order.move-up', 'Move up')}
                  disabled={index === 0}
                  onClick={() => move(index, index - 1)}
                />
                <Button
                  variant="secondary"
                  size="sm"
                  fill="text"
                  icon="arrow-down"
                  aria-label={t('traces.section-order.move-down', 'Move down')}
                  disabled={index === workingOrder.length - 1}
                  onClick={() => move(index, index + 1)}
                />
              </span>
            </div>
            /* eslint-enable jsx-a11y/no-noninteractive-tabindex, jsx-a11y/no-noninteractive-element-interactions */
          );
        })}
      </ul>

      <div className={styles.footer}>
        <Button
          variant="secondary"
          size="sm"
          fill="text"
          icon="history"
          disabled={!isCustomized}
          onClick={() => commit([...DEFAULT_SECTION_ORDER])}
        >
          {t('traces.section-order.reset', 'Reset to default')}
        </Button>
        {isCustomized && (
          <span className={styles.dirty} role="status">
            {t('traces.section-order.unsaved', 'Unsaved order')}
          </span>
        )}
      </div>
    </div>
  );
});

SectionOrderEditor.displayName = 'SectionOrderEditor';
