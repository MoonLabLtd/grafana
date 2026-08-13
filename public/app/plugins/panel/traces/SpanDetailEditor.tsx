import { css } from '@emotion/css';
import { useRef } from 'react';

import { type GrafanaTheme2, type StandardEditorProps } from '@grafana/data';
import { t } from '@grafana/i18n';
import { Button, Icon, useStyles2 } from '@grafana/ui';
import {
  type SpanDetailOptions,
  type SpanDetailSectionId,
  DEFAULT_SPAN_DETAIL_SECTION_ORDER,
  normalizeSpanDetailHiddenSections,
  normalizeSpanDetailSectionOrder,
} from 'app/features/explore/TraceView/components/TraceTimelineViewer/SpanDetail/types';

type Props = StandardEditorProps<SpanDetailOptions, unknown, unknown>;

const getSectionLabel = (id: SpanDetailSectionId): string => {
  switch (id) {
    case 'span':
      return t('traces.span-detail-editor.label-span', 'Span attributes');
    case 'resource':
      return t('traces.span-detail-editor.label-resource', 'Resource attributes');
    case 'events':
      return t('traces.span-detail-editor.label-events', 'Events');
    case 'warnings':
      return t('traces.span-detail-editor.label-warnings', 'Warnings');
    case 'stack':
      return t('traces.span-detail-editor.label-stack', 'Stack trace');
    case 'references':
      return t('traces.span-detail-editor.label-references', 'References');
    case 'flame':
      return t('traces.span-detail-editor.label-flame', 'Flame graph');
  }
};

const EVENTS_FIRST_ORDER: SpanDetailSectionId[] = [
  'events',
  'span',
  'resource',
  'warnings',
  'stack',
  'references',
  'flame',
];

export const SpanDetailEditor = ({ value, onChange }: Props) => {
  const styles = useStyles2(getStyles);
  const order = normalizeSpanDetailSectionOrder(value?.sectionOrder);
  const hiddenSections = normalizeSpanDetailHiddenSections(value?.hiddenSections);
  const draggedId = useRef<SpanDetailSectionId | null>(null);

  const update = (nextOrder: SpanDetailSectionId[], nextHidden: SpanDetailSectionId[]) => {
    onChange({ sectionOrder: nextOrder, hiddenSections: nextHidden });
  };

  const moveItem = (id: SpanDetailSectionId, direction: -1 | 1) => {
    const index = order.indexOf(id);
    const target = index + direction;
    if (index < 0 || target < 0 || target >= order.length) {
      return;
    }
    const nextOrder = [...order];
    [nextOrder[index], nextOrder[target]] = [nextOrder[target], nextOrder[index]];
    update(nextOrder, hiddenSections);
  };

  const toggleVisible = (id: SpanDetailSectionId) => {
    const nextHidden = hiddenSections.includes(id)
      ? hiddenSections.filter((sectionId) => sectionId !== id)
      : [...hiddenSections, id];
    update(order, nextHidden);
  };

  const setPreset = (preset: 'default' | 'events') => {
    update(preset === 'events' ? [...EVENTS_FIRST_ORDER] : [...DEFAULT_SPAN_DETAIL_SECTION_ORDER], []);
  };

  const reset = () => update([...DEFAULT_SPAN_DETAIL_SECTION_ORDER], []);

  const onKeyDown = (event: React.KeyboardEvent<HTMLDivElement>, id: SpanDetailSectionId) => {
    if (event.key === 'ArrowUp') {
      event.preventDefault();
      moveItem(id, -1);
    } else if (event.key === 'ArrowDown') {
      event.preventDefault();
      moveItem(id, 1);
    }
  };

  return (
    <div>
      <span className={styles.subLabel}>{t('traces.span-detail-editor.quick-preset', 'Quick preset')}</span>
      <div className={styles.presetRow}>
        <Button
          className={styles.preset}
          variant="secondary"
          fill="outline"
          size="sm"
          onClick={() => setPreset('events')}
        >
          {t('traces.span-detail-editor.preset-events', 'Events first')}
        </Button>
        <Button
          className={styles.preset}
          variant="secondary"
          fill="outline"
          size="sm"
          onClick={() => setPreset('default')}
        >
          {t('traces.span-detail-editor.preset-default', 'Default order')}
        </Button>
      </div>

      <span className={styles.subLabel}>{t('traces.span-detail-editor.section-order', 'Section order')}</span>
      <div
        className={styles.orderList}
        role="listbox"
        aria-label={t('traces.span-detail-editor.aria-order', 'Reorder span detail sections')}
      >
        {order.map((id, index) => {
          const label = getSectionLabel(id);
          const isHidden = hiddenSections.includes(id);
          return (
            <div
              key={id}
              className={styles.orderItem}
              draggable
              role="option"
              tabIndex={0}
              aria-selected={false}
              aria-label={t('traces.span-detail-editor.item-label', '{{label}}, position {{position}}', {
                label,
                position: index + 1,
              })}
              onKeyDown={(event) => onKeyDown(event, id)}
              onDragStart={() => {
                draggedId.current = id;
              }}
              onDragOver={(event) => event.preventDefault()}
              onDrop={(event) => {
                event.preventDefault();
                const from = draggedId.current;
                draggedId.current = null;
                if (!from || from === id) {
                  return;
                }
                const nextOrder = order.filter((sectionId) => sectionId !== from);
                nextOrder.splice(nextOrder.indexOf(id), 0, from);
                update(nextOrder, hiddenSections);
              }}
            >
              <span className={styles.dragHandle} aria-hidden="true">
                <Icon name="draggabledots" />
              </span>
              <span className={styles.number}>{index + 1}</span>
              <span className={styles.itemName}>{label}</span>
              <button
                type="button"
                className={`${styles.eye} ${isHidden ? styles.eyeOff : ''}`}
                aria-label={
                  isHidden
                    ? t('traces.span-detail-editor.show-section', 'Show {{label}}', { label })
                    : t('traces.span-detail-editor.hide-section', 'Hide {{label}}', { label })
                }
                onClick={() => toggleVisible(id)}
              >
                <Icon name={isHidden ? 'eye-slash' : 'eye'} />
              </button>
            </div>
          );
        })}
      </div>

      <div className={styles.footer}>
        <Button variant="secondary" size="sm" fill="outline" onClick={reset}>
          {t('traces.span-detail-editor.reset', 'Reset')}
        </Button>
      </div>
    </div>
  );
};

const getStyles = (theme: GrafanaTheme2) => ({
  subLabel: css({
    display: 'block',
    fontSize: '11px',
    color: 'var(--color-text-muted)',
    margin: '12px 0 4px',
  }),
  presetRow: css({
    display: 'flex',
    gap: '6px',
  }),
  preset: css({
    flex: '1',
  }),
  orderList: css({
    display: 'flex',
    flexDirection: 'column',
    gap: '6px',
  }),
  orderItem: css({
    display: 'flex',
    alignItems: 'center',
    gap: '8px',
    minHeight: '36px',
    padding: '0 8px',
    border: '1px solid var(--color-border-muted)',
    borderRadius: theme.shape.radius.sm,
    cursor: 'grab',
  }),
  dragHandle: css({
    color: 'var(--color-text-muted)',
    display: 'flex',
  }),
  number: css({
    color: 'var(--color-text-muted)',
    fontFamily: 'monospace',
    fontSize: '11px',
    width: '16px',
  }),
  itemName: css({
    flex: '1',
    fontSize: '12px',
  }),
  eye: css({
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    width: '24px',
    height: '24px',
    border: 'none',
    background: 'none',
    cursor: 'pointer',
    color: 'var(--color-text-link)',
  }),
  eyeOff: css({
    color: 'var(--color-text-muted)',
  }),
  footer: css({
    marginTop: '10px',
    display: 'flex',
    justifyContent: 'flex-end',
  }),
});
