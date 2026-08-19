import {
  DEFAULT_SECTION_ORDER,
  normalizeSectionOrder,
  isSectionId,
  type SectionId,
} from './sectionOrder';

describe('normalizeSectionOrder', () => {
  it('returns the default order for undefined, null and empty arrays', () => {
    expect(normalizeSectionOrder()).toEqual(DEFAULT_SECTION_ORDER);
    expect(normalizeSectionOrder(null)).toEqual(DEFAULT_SECTION_ORDER);
    expect(normalizeSectionOrder([])).toEqual(DEFAULT_SECTION_ORDER);
  });

  it('returns the passed order unchanged when already complete and valid', () => {
    const order: SectionId[] = ['events', 'spanAttributes', 'resourceAttributes', 'warnings', 'stackTraces'];
    expect(normalizeSectionOrder(order)).toEqual(order);
  });

  it('ignores unknown ids', () => {
    const result = normalizeSectionOrder([
      'nope' as SectionId,
      'events',
      'spanAttributes',
      'resourceAttributes',
      'warnings',
      'stackTraces',
    ]);
    expect(result).toEqual(['events', 'spanAttributes', 'resourceAttributes', 'warnings', 'stackTraces']);
  });

  it('appends known sections missing from a partial array in default relative order', () => {
    const result = normalizeSectionOrder(['events', 'spanAttributes']);
    expect(result).toEqual(['events', 'spanAttributes', 'resourceAttributes', 'warnings', 'stackTraces']);
  });

  it('deduplicates duplicates preserving the first occurrence', () => {
    const result = normalizeSectionOrder([
      'events',
      'events',
      'spanAttributes',
      'spanAttributes',
      'warnings',
      'stackTraces',
      'resourceAttributes',
    ]);
    expect(result).toEqual(['events', 'spanAttributes', 'warnings', 'stackTraces', 'resourceAttributes']);
  });

  it('truncates an oversized array to exactly five unique known ids in the correct order', () => {
    const huge = Array<SectionId>(1000).fill('events');
    const result = normalizeSectionOrder(huge);
    expect(result).toHaveLength(5);
    expect(result).toEqual(['events', 'spanAttributes', 'resourceAttributes', 'warnings', 'stackTraces']);
  });
});

describe('isSectionId', () => {
  it('returns true for known ids and false otherwise', () => {
    expect(isSectionId('events')).toBe(true);
    expect(isSectionId('stackTraces')).toBe(true);
    expect(isSectionId('bogus')).toBe(false);
    expect(isSectionId(undefined)).toBe(false);
    expect(isSectionId(null)).toBe(false);
  });
});
