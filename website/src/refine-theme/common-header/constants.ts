import {
    DocumentsIcon,
    IntegrationsIcon,
    TutorialIcon,
    ExamplesIcon,
    AwesomeIcon,
    ContributeIcon,
    HackathonsIcon,
    AboutUsIcon,
    StoreIcon,
    BlogIcon,
} from "../icons/popover";

export type NavbarPopoverItemType = {
    isPopover: true;
    label: string;
    items: {
        label: string;
        description: string;
        link: string;
        icon: React.FC;
    }[];
};

export type NavbarItemType = {
    isPopover?: false;
    label: string;
    icon?: React.FC;
    href?: string;
};

export type MenuItemType = NavbarPopoverItemType | NavbarItemType;

export const MENU_ITEMS: MenuItemType[] = [
  {
    isPopover: true,
    label: "Resources",
    items: [
      {
        label: "Documentation",
        description: "Setup, configuration, and API reference.",
        link: "/docs/",
        icon: DocumentsIcon,
      },
      {
        label: "How-to Guides",
        description: "Step-by-step guides for common tasks.",
        link: "/docs/articles/intro/",
        icon: TutorialIcon,
      },
      {
        label: "Features",
        description: "See what's included, feature by feature.",
        link: "/features/",
        icon: AwesomeIcon,
      },
      {
        label: "Changelog",
        description: "Every release, documented as it ships.",
        link: "/docs/changelog/intro/",
        icon: IntegrationsIcon,
      },
      {
        label: "Discussions",
        description: "Share ideas, upvote features.",
        link: "https://github.com/stefanpejcic/OpenPanel/discussions",
        icon: ContributeIcon,
      },
      {
        label: "Blog",
        description: "Tips and guides on Linux hosting.",
        link: "/blog",
        icon: BlogIcon,
      },
    ],
  },
  {
    isPopover: false,
    label: "Features",
    href: "/features",
  },
  {
    isPopover: false,
    label: "Live Demo",
    href: "/demo",
  },
  {
    isPopover: true,
    label: "Pricing",
    items: [
      {
        label: "OpenPanel Community",
        description: "Forever free, self-hosted.",
        link: "/community/",
        icon: ExamplesIcon,
      },
      {
        label: "OpenPanel Enterprise",
        description: "Fixed pricing at 14.95€ monthly.",
        link: "/enterprise/",
        icon: StoreIcon,
      },
    ],
  },
];
