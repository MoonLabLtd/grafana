import {
  DEFAULT_SPAN_DETAIL_SECTION_ORDER,
  normalizeSpanDetailHiddenSections,
  normalizeSpanDetailOrdering,
  normalizeSpanDetailSectionOrder,
} from './types';

const DEFAULT = [...DEFAULT_SPAN_DETAIL_SECTION_ORDER];

describe('normalizeSpanDetailSectionOrder', () => {
  it('returns the default order for undefined, null, and empty input', () => {
    expect(normalizeSpanDetailSectionOrder(undefined)).toEqual(DEFAULT);
    expect(normalizeSpanDetailSectionOrder(null as unknown as undefined)).toEqual(DEFAULT);
    expect(normalizeSpanDetailSectionOrder([])).toEqual(DEFAULT);
  });

  it('returns the canonical default order', () => {
    expect(DEFAULT).toEqual(['span', 'resource', 'events', 'warnings', 'stack', 'references', 'flame']);
  });

  it('preserves a full custom order', () => {
    expect(
      normalizeSpanDetailSectionOrder(['events', 'span', 'resource', 'warnings', 'stack', 'references', 'flame'])
    ).toEqual(['events', 'span', 'resource', 'warnings', 'stack', 'references', 'flame']);
  });

  it('appends sections missing from a partial order in canonical relative order', () => {
    expect(normalizeSpanDetailSectionOrder(['events', 'flame'])).toEqual([
      'events',
      'flame',
      'span',
      'resource',
      'warnings',
      'stack',
      'references',
    ]);
  });

  it('deduplicates repeated ids keeping the first occurrence', () => {
    expect(normalizeSpanDetailSectionOrder(['events', 'span', 'events', 'resource'])).toEqual([
      'events',
      'span',
      'resource',
      'warnings',
      'stack',
      'references',
      'flame',
    ]);
  });

  it('ignores unknown ids and appends remaining sections', () => {
    expect(normalizeSpanDetailSectionOrder(['futureSection', 'events', 'unknown'])).toEqual([
      'events',
      'span',
      'resource',
      'warnings',
      'stack',
      'references',
      'flame',
    ]);
  });
});

describe('normalizeSpanDetailHiddenSections', () => {
  it('returns an empty array for undefined input', () => {
    expect(normalizeSpanDetailHiddenSections(undefined)).toEqual([]);
  });

  it('keeps valid ids and drops unknown or duplicate ones', () => {
    expect(normalizeSpanDetailHiddenSections(['events', 'unknown', 'events', 'flame'])).toEqual(['events', 'flame']);
  });
});

describe('normalizeSpanDetailOrdering', () => {
  it('returns the default ordering when options are empty', () => {
    expect(normalizeSpanDetailOrdering(undefined)).toEqual({ sectionOrder: DEFAULT, hiddenSections: [] });
    expect(normalizeSpanDetailOrdering({})).toEqual({ sectionOrder: DEFAULT, hiddenSections: [] });
  });

  it('normalizes both section order and hidden sections', () => {
    expect(normalizeSpanDetailOrdering({ sectionOrder: ['events', 'span'], hiddenSections: ['events'] })).toEqual({
      sectionOrder: ['events', 'span', 'resource', 'warnings', 'stack', 'references', 'flame'],
      hiddenSections: ['events'],
    });
  });
});
