import { Page, Link } from '@/store/modules/graph/types';

const pages: Page[] = [
  {
    id: 0,
    name: 'France',
  },
  {
    id: 1,
    name: 'Western_Europe',
  },
  {
    id: 2,
    name: 'Belgium',
  },
  {
    id: 3,
    name: 'Luxembourg',
  },
  {
    id: 4,
    name: 'Germany',
  },
];

const links: Link[] = [
  {
    id: 0,
    from: 0,
    to: 1,
  },
  {
    id: 1,
    from: 0,
    to: 2,
  },
  {
    id: 2,
    from: 0,
    to: 3,
  },
  {
    id: 3,
    from: 0,
    to: 4,
  },
];

export const mockedGetPage = async (pageName: string) => {
  const page = pages.find(p => p.name === pageName);
  if (page !== undefined) {
    const pageLinks = links.filter(l => l.from === page.id || l.to === page.id);
    const linkedPages = pages.filter(
      p => pageLinks.find(l => l.from === p.id || l.to === p.id) !== undefined
    );
    return {
      center: page.id,
      pages: linkedPages,
      links: pageLinks,
    };
  }
  return {
    center: undefined,
    pages: [],
    links: [],
  };
};
