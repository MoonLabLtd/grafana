// Stable identifiers for the five reorderable span detail sections. These IDs are
// persisted in panel JSON and are locale-independent; they are never displayed to
// end users directly. Labels are resolved through the Grafana i18n function `t()`.
export type SectionId = 'spanAttributes' | 'resourceAttributes' | 'events' | 'warnings' | 'stackTraces';

export const DEFAULT_SECTION_ORDER: SectionId[] = [
  'spanAttributes',
  'resourceAttributes',
  'events',
  'warnings',
  'stackTraces',
];

const SECTION_ID_SET = new Set<string>(DEFAULT_SECTION_ORDER);

export function isSectionId(value: string | undefined | null): value is SectionId {
  return value != null && SECTION_ID_SET.has(value);
}

// Metadata used by the panel editor to render the reorder list. `labelKey` is the
// i18n key and `label` the default fallback text; `tag` is a stable non-localized
// hint (not displayed to end users).
export interface SectionOrderMetadata {
  labelKey: string;
  label: string;
  tag: string;
}

export const SECTION_ORDER_METADATA: Record<SectionId, SectionOrderMetadata> = {
  spanAttributes: {
    labelKey: 'explore.span-detail.label-span-attributes',
    label: 'Span attributes',
    tag: 'span',
  },
  resourceAttributes: {
    labelKey: 'explore.span-detail.label-resource-attributes',
    label: 'Resource attributes',
    tag: 'resource',
  },
  events: {
    labelKey: 'explore.accordian-logs.events',
    label: 'Events',
    tag: 'exception',
  },
  warnings: {
    labelKey: 'explore.span-detail.label-warnings',
    label: 'Warnings',
    tag: 'warn',
  },
  stackTraces: {
    labelKey: 'explore.span-detail.label-stack-trace',
    label: 'Stack trace',
    tag: 'stack',
  },
};

/**
 * Normalises a configured section order into a canonical full order.
 * - Unknown IDs are ignored.
 * - Duplicate IDs are deduplicated preserving the first occurrence.
 * - Any known section omitted from the array is appended in default relative order.
 * - `null`, `undefined` and empty arrays fall back to the default order.
 * The result always contains exactly the five known section IDs, each once.
 */
export function normalizeSectionOrder(order?: SectionId[] | null): SectionId[] {
  if (!order || order.length === 0) {
    return [...DEFAULT_SECTION_ORDER];
  }

  const result: SectionId[] = [];
  const seen = new Set<SectionId>();

  for (const id of order) {
    if (isSectionId(id) && !seen.has(id)) {
      seen.add(id);
      result.push(id);
    }
  }

  for (const id of DEFAULT_SECTION_ORDER) {
    if (!seen.has(id)) {
      seen.add(id);
      result.push(id);
    }
  }

  return result;
}
