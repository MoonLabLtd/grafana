export const SPAN_DETAIL_SECTION_IDS = [
  'span',
  'resource',
  'events',
  'warnings',
  'stack',
  'references',
  'flame',
] as const;

export type SpanDetailSectionId = (typeof SPAN_DETAIL_SECTION_IDS)[number];

export const DEFAULT_SPAN_DETAIL_SECTION_ORDER: SpanDetailSectionId[] = [...SPAN_DETAIL_SECTION_IDS];

export const isSpanDetailSectionId = (value: unknown): value is SpanDetailSectionId =>
  typeof value === 'string' && SPAN_DETAIL_SECTION_IDS.some((id) => id === value);

export interface SpanDetailOptions {
  sectionOrder?: SpanDetailSectionId[];
  hiddenSections?: SpanDetailSectionId[];
}

export interface SpanDetailOrdering {
  sectionOrder: SpanDetailSectionId[];
  hiddenSections: SpanDetailSectionId[];
}

/**
 * Normalizes a persisted section order into a complete, de-duplicated, valid
 * seven-section list. Unknown IDs are dropped, duplicates are collapsed to
 * their first occurrence and any section missing from a partial config is
 * appended in the canonical default order. Malformed or empty input resolves
 * to the canonical default order.
 */
export const normalizeSpanDetailSectionOrder = (
  sectionOrder: readonly unknown[] | undefined
): SpanDetailSectionId[] => {
  if (!sectionOrder || sectionOrder.length === 0) {
    return [...DEFAULT_SPAN_DETAIL_SECTION_ORDER];
  }

  const result: SpanDetailSectionId[] = [];
  for (const id of sectionOrder) {
    if (isSpanDetailSectionId(id) && !result.includes(id)) {
      result.push(id);
    }
  }

  for (const id of DEFAULT_SPAN_DETAIL_SECTION_ORDER) {
    if (!result.includes(id)) {
      result.push(id);
    }
  }

  return result;
};

export const normalizeSpanDetailHiddenSections = (
  hiddenSections: readonly unknown[] | undefined
): SpanDetailSectionId[] => {
  if (!hiddenSections) {
    return [];
  }

  const result: SpanDetailSectionId[] = [];
  for (const id of hiddenSections) {
    if (isSpanDetailSectionId(id) && !result.includes(id)) {
      result.push(id);
    }
  }

  return result;
};

export const normalizeSpanDetailOrdering = (options?: SpanDetailOptions): SpanDetailOrdering => ({
  sectionOrder: normalizeSpanDetailSectionOrder(options?.sectionOrder),
  hiddenSections: normalizeSpanDetailHiddenSections(options?.hiddenSections),
});
