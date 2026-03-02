import type {SidebarsConfig} from '@docusaurus/plugin-content-docs';

const sidebars: SidebarsConfig = {
  docsSidebar: [
    {
      type: 'category',
      label: 'Getting Started',
      collapsed: false,
      items: [
        'getting-started/introduction',
        'getting-started/installation',
        'getting-started/first-deployment',
        'getting-started/verifying',
        'getting-started/next-steps',
      ],
    },
    {
      type: 'category',
      label: 'Guides',
      collapsed: false,
      items: [
        'guides/node-management',
        'guides/configuration',
        'guides/deploying-components',
        'guides/monitoring',
        'guides/security',
        'guides/docker',
        'guides/kubernetes',
        'guides/repository',
        'guides/troubleshooting',
      ],
    },
    {
      type: 'category',
      label: 'Reference',
      collapsed: true,
      items: [
        'reference/cli',
        'reference/api-overview',
        'reference/api-meta',
        'reference/api-system',
        'reference/api-nodes',
        'reference/api-onramp',
        'reference/configuration',
        'reference/components',
      ],
    },
    {
      type: 'category',
      label: 'Concepts',
      collapsed: true,
      items: [
        'concepts/architecture',
        'concepts/providers',
        'concepts/tasks',
        'concepts/deployment-state',
      ],
    },
    {
      type: 'category',
      label: 'Developer',
      collapsed: true,
      items: [
        'developer/README',
        'developer/architecture',
        'developer/security',
        {
          type: 'category',
          label: 'Providers',
          items: [
            'developer/providers/meta',
            'developer/providers/onramp',
            'developer/providers/system',
          ],
        },
      ],
    },
  ],
};

export default sidebars;
